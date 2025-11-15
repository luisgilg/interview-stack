using System.Text.Json;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Interfaces;
using Moq;

namespace DotnetService.Tests;

public sealed class CacheServiceTests
{
    private static CacheConfig CreateConfig(bool enabled = true) => new()
    {
        Enabled = enabled,
        DefaultTtl = TimeSpan.FromSeconds(1),
        StaleTtl = TimeSpan.FromSeconds(1),
        Redis = new RedisConfig()
    };

    [Fact]
    public async Task FetchAsync_BypassesWhenDisabled()
    {
        var client = new Mock<ICacheClient>();
        var logger = new Mock<ILoggerAdapter>();
        var clock = new Mock<IClock>();
        var service = new CacheService(client.Object, CreateConfig(false), logger.Object, clock.Object);

        var result = await service.FetchAsync("key", _ => Task.FromResult("value"), CancellationToken.None);

        Assert.Equal(CacheStatus.Bypass, result.Status);
        Assert.Equal("value", result.Value);
        client.Verify(c => c.GetAsync(It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Never);
    }

    [Fact]
    public async Task FetchAsync_StoresValueOnMiss()
    {
        var client = new Mock<ICacheClient>();
        var logger = new Mock<ILoggerAdapter>();
        var clock = new Mock<IClock>();
        clock.Setup(c => c.UtcNow()).Returns(new DateTime(2024, 01, 01, 0, 0, 0, DateTimeKind.Utc));
        var service = new CacheService(client.Object, CreateConfig(), logger.Object, clock.Object);

        var result = await service.FetchAsync("products:list", _ => Task.FromResult("payload"), CancellationToken.None);

        Assert.Equal(CacheStatus.Miss, result.Status);
        Assert.Equal("payload", result.Value);
        client.Verify(c => c.SetAsync("products:list", It.IsAny<string>(), It.IsAny<TimeSpan>(), It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task FetchAsync_ReturnsFreshCacheHit()
    {
        var now = new DateTime(2024, 01, 01, 0, 0, 10, DateTimeKind.Utc);
        var client = new Mock<ICacheClient>();
        client.Setup(c => c.GetAsync("products:list", It.IsAny<CancellationToken>()))
            .ReturnsAsync(EncodeEnvelope("cached", now));
        var logger = new Mock<ILoggerAdapter>();
        var clock = new Mock<IClock>();
        clock.Setup(c => c.UtcNow()).Returns(now.AddMilliseconds(500));
        var service = new CacheService(client.Object, CreateConfig(), logger.Object, clock.Object);
        var loader = new Mock<Func<CancellationToken, Task<string>>>();

        var result = await service.FetchAsync("products:list", loader.Object, CancellationToken.None);

        Assert.Equal(CacheStatus.Fresh, result.Status);
        Assert.Equal("cached", result.Value);
        loader.Verify(l => l(It.IsAny<CancellationToken>()), Times.Never);
    }

    [Fact]
    public async Task FetchAsync_ReturnsStaleAndRefreshes()
    {
        var now = new DateTime(2024, 01, 01, 0, 0, 2, DateTimeKind.Utc);
        var client = new Mock<ICacheClient>();
        client.Setup(c => c.GetAsync("products:list", It.IsAny<CancellationToken>()))
            .ReturnsAsync(EncodeEnvelope("old", now.AddMilliseconds(-1500)));
        var setTcs = new TaskCompletionSource<bool>();
        client.Setup(c => c.SetAsync("products:list", It.IsAny<string>(), It.IsAny<TimeSpan>(), It.IsAny<CancellationToken>()))
            .Returns(() =>
            {
                setTcs.TrySetResult(true);
                return Task.CompletedTask;
            });

        var logger = new Mock<ILoggerAdapter>();
        var clock = new Mock<IClock>();
        clock.Setup(c => c.UtcNow()).Returns(now);
        var service = new CacheService(client.Object, CreateConfig(), logger.Object, clock.Object);
        var loader = new Mock<Func<CancellationToken, Task<string>>>();
        loader.Setup(l => l(It.IsAny<CancellationToken>())).ReturnsAsync("new value");

        var result = await service.FetchAsync("products:list", loader.Object, CancellationToken.None);

        Assert.Equal(CacheStatus.Stale, result.Status);
        Assert.Equal("old", result.Value);
        await setTcs.Task.WaitAsync(TimeSpan.FromSeconds(1));
        loader.Verify(l => l(It.IsAny<CancellationToken>()), Times.Once);
    }

    private static string EncodeEnvelope<T>(T payload, DateTime storedAt) =>
        JsonSerializer.Serialize(new { Payload = payload, StoredAt = storedAt });
}
