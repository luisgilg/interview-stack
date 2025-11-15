using DotnetService.Internal.Application.Services;

namespace DotnetService.Internal.Infrastructure.Cache;

public sealed class NoopCacheClient : ICacheClient
{
    public Task<string?> GetAsync(string key, CancellationToken cancellationToken) => Task.FromResult<string?>(null);

    public Task SetAsync(string key, string payload, TimeSpan ttl, CancellationToken cancellationToken) => Task.CompletedTask;

    public Task DeleteAsync(string key, CancellationToken cancellationToken) => Task.CompletedTask;
}
