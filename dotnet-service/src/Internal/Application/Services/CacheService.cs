using System;
using System.Text.Json;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Interfaces;
using DotnetService.Internal.Observability;

namespace DotnetService.Internal.Application.Services;

public enum CacheStatus
{
    Bypass,
    Miss,
    Fresh,
    Stale
}

public sealed record CacheResult<T>(T Value, CacheStatus Status);

public interface ICacheClient
{
    Task<string?> GetAsync(string key, CancellationToken cancellationToken);
    Task SetAsync(string key, string payload, TimeSpan ttl, CancellationToken cancellationToken);
    Task DeleteAsync(string key, CancellationToken cancellationToken);
}

public sealed class CacheService
{
    private  readonly  ICacheClient _client;
    private readonly ILoggerAdapter _logger;
    private readonly IClock _clock;
    private readonly bool _enabled;
    private readonly TimeSpan _defaultTtl;
    private readonly TimeSpan _staleTtl;

    public CacheService(ICacheClient client, CacheConfig config, ILoggerAdapter logger, IClock clock)
    {
        ArgumentNullException.ThrowIfNull(client);
        ArgumentNullException.ThrowIfNull(logger);
        ArgumentNullException.ThrowIfNull(clock);
        ArgumentNullException.ThrowIfNull(config);

        _client = client;
        _logger = logger;
        _clock = clock;
        _defaultTtl = config.DefaultTtl;
        _staleTtl = config.StaleTtl;
        _enabled = config.Enabled && _client is not null && _defaultTtl > TimeSpan.Zero;
    }

    public bool Enabled => _enabled;

    public async Task WriteAsync<T>(string key, T value, CancellationToken cancellationToken)
    {
        if (!_enabled)
        {
            return;
        }
        try
        {
            await PersistAsync(key, value, cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            _logger.Warn("cache write failed", new { Key = key, Error = ex.Message });
        }
    }

    public async Task DeleteAsync(string key, CancellationToken cancellationToken)
    {
        if (!_enabled)
        {
            return;
        }
        try
        {
            await _client.DeleteAsync(key, cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            _logger.Warn("cache delete failed", new { Key = key, Error = ex.Message });
        }
    }

    public async Task<CacheResult<T>> FetchAsync<T>(string key, Func<CancellationToken, Task<T>> loader, CancellationToken cancellationToken)
    {
        if (!_enabled || string.IsNullOrWhiteSpace(key))
        {
            var value = await loader(cancellationToken).ConfigureAwait(false);
            return Track(new CacheResult<T>(value, CacheStatus.Bypass));
        }

        try
        {
            var cached = await _client.GetAsync(key, cancellationToken).ConfigureAwait(false);
            if (!string.IsNullOrWhiteSpace(cached))
            {
                var envelope = JsonSerializer.Deserialize<CacheEnvelope<T>>(cached);
                if (envelope is not null)
                {
                    var age = _clock.UtcNow() - envelope.StoredAt;
                    if (age <= _defaultTtl)
                    {
                        return Track(new CacheResult<T>(envelope.Payload, CacheStatus.Fresh));
                    }

                    if (age <= _defaultTtl + _staleTtl)
                    {
                        _ = RefreshAsync(key, loader);
                        return Track(new CacheResult<T>(envelope.Payload, CacheStatus.Stale));
                    }
                }
            }
        }
        catch (Exception ex)
        {
            _logger.Warn("cache fetch failed", new { Key = key, Error = ex.Message });
            var fallback = await loader(cancellationToken).ConfigureAwait(false);
            return Track(new CacheResult<T>(fallback, CacheStatus.Bypass));
        }

        return await LoadAndStoreAsync(key, loader, cancellationToken).ConfigureAwait(false);
    }

    private async Task<CacheResult<T>> LoadAndStoreAsync<T>(string key, Func<CancellationToken, Task<T>> loader, CancellationToken cancellationToken)
    {
        var value = await loader(cancellationToken).ConfigureAwait(false);
        try
        {
            await PersistAsync(key, value, cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            _logger.Warn("cache store failed", new { Key = key, Error = ex.Message });
        }
        return Track(new CacheResult<T>(value, CacheStatus.Miss));
    }

    private async Task PersistAsync<T>(string key, T value, CancellationToken cancellationToken)
    {
        if (!_enabled)
        {
            return;
        }

        var ttl = _defaultTtl + _staleTtl;
        if (ttl <= TimeSpan.Zero)
        {
            return;
        }

        var envelope = new CacheEnvelope<T>(value, _clock.UtcNow());
        var payload = JsonSerializer.Serialize(envelope);
        await _client.SetAsync(key, payload, ttl, cancellationToken).ConfigureAwait(false);
    }

    private Task RefreshAsync<T>(string key, Func<CancellationToken, Task<T>> loader)
    {
        return Task.Run(async () =>
        {
            try
            {
                var refreshed = await loader(CancellationToken.None).ConfigureAwait(false);
                await PersistAsync(key, refreshed, CancellationToken.None).ConfigureAwait(false);
            }
            catch (Exception ex)
            {
                _logger.Warn("cache refresh failed", new { Key = key, Error = ex.Message });
            }
        });
    }

    private static CacheResult<T> Track<T>(CacheResult<T> result)
    {
        MetricsRegistry.RecordCacheStatus(result.Status);
        return result;
    }

    private sealed record CacheEnvelope<T>(T Payload, DateTime StoredAt);
}
