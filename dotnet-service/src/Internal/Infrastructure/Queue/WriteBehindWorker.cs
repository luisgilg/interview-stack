using System.Diagnostics;
using System.Text.Json;
using System.Text.Json.Serialization;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Events;
using DotnetService.Internal.Domain.Interfaces;
using DotnetService.Internal.Observability;
using Microsoft.Extensions.Hosting;
using StackExchange.Redis;

namespace DotnetService.Internal.Infrastructure.Queue;

public sealed class WriteBehindWorker : BackgroundService
{
    private readonly IConnectionMultiplexer _connection;
    private readonly string _streamName;
    private readonly string _groupName;
    private readonly string _consumerName;
    private readonly int _batchSize;
    private readonly TimeSpan _blockTimeout;
    private readonly IProductStore _store;
    private readonly ILoggerAdapter _logger;
    private readonly JsonSerializerOptions _jsonOptions;
    private readonly string _source;

    public WriteBehindWorker(
        IConnectionMultiplexer connection,
        AppConfig config,
        IProductStore store,
        ILoggerAdapter logger)
    {
        _connection = connection;
        _streamName = config.WriteBehind.StreamName;
        _groupName = "dotnet-service";
        _consumerName = $"{_groupName}-{Environment.ProcessId}-{Guid.NewGuid():N}";
        _batchSize = config.WriteBehind.BatchSize;
        _blockTimeout = config.WriteBehind.FlushInterval;
        _store = store;
        _logger = logger;
        _source = _groupName;
        _jsonOptions = new JsonSerializerOptions(JsonSerializerDefaults.Web);
        _jsonOptions.Converters.Add(new JsonStringEnumConverter(JsonNamingPolicy.CamelCase));
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        if (!_connection.IsConnected)
        {
            return;
        }
        await EnsureGroupAsync(stoppingToken).ConfigureAwait(false);
        var db = _connection.GetDatabase();
        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                var entries = await db.StreamReadGroupAsync(
                    _streamName,
                    _groupName,
                    _consumerName,
                    ">",
                    count: _batchSize).ConfigureAwait(false);
                if (entries.Length == 0)
                {
                    await Task.Delay(_blockTimeout, stoppingToken).ConfigureAwait(false);
                    continue;
                }
                var batchTimer = Stopwatch.StartNew();
                var nowMs = DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();
                var maxLagSeconds = 0.0;
                foreach (var entry in entries)
                {
                    var lagSeconds = CalculateLagSeconds(nowMs, entry.Id);
                    if (lagSeconds > maxLagSeconds)
                    {
                        maxLagSeconds = lagSeconds;
                    }
                    await ProcessEntryAsync(db, entry, stoppingToken).ConfigureAwait(false);
                }
                batchTimer.Stop();
                long queueLength = -1;
                try
                {
                    queueLength = await db.StreamLengthAsync(_streamName).ConfigureAwait(false);
                }
                catch (Exception ex)
                {
                    _logger.Warn("failed to read write-behind queue length", new { Error = ex.Message });
                    MetricsRegistry.RecordWriteBehindError();
                }
                MetricsRegistry.RecordWriteBehindBatch(entries.Length, batchTimer.Elapsed, maxLagSeconds, queueLength);
            }
            catch (RedisException ex)
            {
                _logger.Warn("write-behind read failed", new { Error = ex.Message });
                MetricsRegistry.RecordWriteBehindError();
                await Task.Delay(TimeSpan.FromSeconds(1), stoppingToken).ConfigureAwait(false);
            }
        }
    }

    private async Task ProcessEntryAsync(IDatabase db, StreamEntry entry, CancellationToken cancellationToken)
    {
        try
        {
            var payload = entry["event"];
            if (payload.IsNullOrEmpty)
            {
                await AckAsync(db, entry.Id).ConfigureAwait(false);
                return;
            }
            var writeEvent = JsonSerializer.Deserialize<WriteEvent>(payload!, _jsonOptions);
            if (writeEvent is null || (writeEvent.Source is not null && writeEvent.Source != _source))
            {
                await AckAsync(db, entry.Id).ConfigureAwait(false);
                return;
            }
            await ApplyEventAsync(writeEvent, cancellationToken).ConfigureAwait(false);
            await AckAsync(db, entry.Id).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            _logger.Warn("write-behind apply failed", new { Error = ex.Message });
            MetricsRegistry.RecordWriteBehindError();
        }
    }

    private async Task ApplyEventAsync(WriteEvent writeEvent, CancellationToken cancellationToken)
    {
        switch (writeEvent.Type)
        {
            case WriteEventType.Create when writeEvent.Payload is not null:
                await _store.CreateProductAsync(writeEvent.Payload, cancellationToken).ConfigureAwait(false);
                break;
            case WriteEventType.Update when writeEvent.Payload is not null:
                await _store.UpdateProductAsync(writeEvent.Id, writeEvent.Payload, cancellationToken).ConfigureAwait(false);
                break;
            case WriteEventType.Delete:
                await _store.DeleteProductAsync(writeEvent.Id, cancellationToken).ConfigureAwait(false);
                break;
        }
    }

    private async Task AckAsync(IDatabase db, RedisValue id)
    {
        try
        {
            await db.StreamAcknowledgeAsync(_streamName, _groupName, id).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            _logger.Warn("failed to ack write-behind message", new { Error = ex.Message });
            MetricsRegistry.RecordWriteBehindError();
        }
    }

    private async Task EnsureGroupAsync(CancellationToken cancellationToken)
    {
        try
        {
            var db = _connection.GetDatabase();
            await db.StreamCreateConsumerGroupAsync(_streamName, _groupName, "$", true).ConfigureAwait(false);
        }
        catch (RedisServerException ex) when (ex.Message.Contains("BUSYGROUP", StringComparison.OrdinalIgnoreCase))
        {
            // group already exists
        }
    }

    private static double CalculateLagSeconds(long nowMs, RedisValue id)
    {
        if (id.IsNullOrEmpty)
        {
            return 0;
        }
        var raw = id.ToString();
        var dashIndex = raw.IndexOf('-');
        var timestampSegment = dashIndex >= 0 ? raw[..dashIndex] : raw;
        if (!long.TryParse(timestampSegment, out var producedAt))
        {
            return 0;
        }
        var lagMs = Math.Max(0, nowMs - producedAt);
        return lagMs / 1000d;
    }
}
