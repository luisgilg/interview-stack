using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Interfaces;
using StackExchange.Redis;

namespace DotnetService.Internal.Infrastructure.Cache;

public sealed class RedisCacheClient(IConnectionMultiplexer connection, CacheConfig config, ILoggerAdapter logger) : ICacheClient
{
    private readonly IConnectionMultiplexer _connection = connection;
    private readonly CacheConfig _config = config;
    private readonly ILoggerAdapter _logger = logger;

    public async Task<string?> GetAsync(string key, CancellationToken cancellationToken)
    {
        var db = _connection.GetDatabase(_config.Redis.Database);
        var result = await db.StringGetAsync(key).ConfigureAwait(false);
        return result.HasValue ? result.ToString() : null;
    }

    public Task SetAsync(string key, string payload, TimeSpan ttl, CancellationToken cancellationToken)
    {
        var db = _connection.GetDatabase(_config.Redis.Database);
        return db.StringSetAsync(key, payload, ttl);
    }

    public Task DeleteAsync(string key, CancellationToken cancellationToken)
    {
        var db = _connection.GetDatabase(_config.Redis.Database);
        return db.KeyDeleteAsync(key);
    }
}
