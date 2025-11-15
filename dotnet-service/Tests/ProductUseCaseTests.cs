using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Application.UseCases;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Entities;
using DotnetService.Internal.Domain.Errors;
using DotnetService.Internal.Domain.Interfaces;
using DotnetService.Internal.Infrastructure.Cache;
using Moq;

namespace DotnetService.Tests;

public sealed class ProductUseCaseTests
{
    [Fact]
    public async Task CreateProductUseCase_DelegatesToRepositoryAndLogs()
    {
        var now = new DateTime(2024, 02, 02, 10, 0, 0, DateTimeKind.Utc);
        var repository = new Mock<IProductRepository>(MockBehavior.Strict);
        var logger = new Mock<ILoggerAdapter>(MockBehavior.Loose);
        var clock = new Mock<IClock>();
        clock.Setup(c => c.UtcNow()).Returns(now);
        repository
            .Setup(r => r.CreateAsync(It.IsAny<Product>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync((Product product, CancellationToken _) => product with { Id = "abc" });
        var cache = new CacheService(new NoopCacheClient(), new CacheConfig { Enabled = false }, logger.Object, clock.Object);
        var queue = new Mock<IWriteQueueProducer>();
        var config = new AppConfig { Cache = new CacheConfig { Enabled = false }, WriteBehind = new WriteBehindConfig { Enabled = false } };
        var useCase = new CreateProductUseCase(repository.Object, logger.Object, clock.Object, cache, queue.Object, config);
        var response = await useCase.HandleAsync(new ProductRequestDto("Keyboard", 10, Array.Empty<string>()), CancellationToken.None);

        repository.Verify(r => r.CreateAsync(
                It.Is<Product>(p => p.CreatedAt == now && p.UpdatedAt == now && p.Name == "Keyboard"),
                It.IsAny<CancellationToken>()),
            Times.Once);
        logger.Verify(l => l.Info("product created", It.IsAny<object?>()), Times.Once);
        Assert.Equal("abc", response.Id);
    }

    [Fact]
    public async Task CreateProductUseCase_ThrowsOnValidationErrors()
    {
        var repository = new Mock<IProductRepository>();
        var logger = new Mock<ILoggerAdapter>();
        var clock = new Mock<IClock>();
        var cache = new CacheService(new NoopCacheClient(), new CacheConfig { Enabled = false }, logger.Object, clock.Object);
        var queue = new Mock<IWriteQueueProducer>();
        var config = new AppConfig { Cache = new CacheConfig { Enabled = false }, WriteBehind = new WriteBehindConfig { Enabled = false } };
        var useCase = new CreateProductUseCase(repository.Object, logger.Object, clock.Object, cache, queue.Object, config);

        await Assert.ThrowsAsync<DomainException>(() =>
            useCase.HandleAsync(new ProductRequestDto(string.Empty, 0, null), CancellationToken.None));
        repository.Verify(r => r.CreateAsync(It.IsAny<Product>(), It.IsAny<CancellationToken>()), Times.Never);
    }

    [Fact]
    public async Task CreateProductUseCase_EnqueuesWhenWriteBehindEnabled()
    {
        var now = new DateTime(2024, 05, 01, 0, 0, 0, DateTimeKind.Utc);
        var repository = new Mock<IProductRepository>(MockBehavior.Strict);
        var logger = new Mock<ILoggerAdapter>(MockBehavior.Loose);
        var clock = new Mock<IClock>();
        clock.Setup(c => c.UtcNow()).Returns(now);
        var cache = new CacheService(new NoopCacheClient(), new CacheConfig { Enabled = false }, logger.Object, clock.Object);
        var queue = new Mock<IWriteQueueProducer>();
        var config = new AppConfig { WriteBehind = new WriteBehindConfig { Enabled = true } };
        var useCase = new CreateProductUseCase(repository.Object, logger.Object, clock.Object, cache, queue.Object, config);

        var response = await useCase.HandleAsync(new ProductRequestDto("Buffered", 25, null), CancellationToken.None);

        queue.Verify(q => q.EnqueueAsync(It.IsAny<DotnetService.Internal.Domain.Events.WriteEvent>(), It.IsAny<CancellationToken>()), Times.Once);
        repository.Verify(r => r.CreateAsync(It.IsAny<Product>(), It.IsAny<CancellationToken>()), Times.Never);
        Assert.False(string.IsNullOrEmpty(response.Id));
    }
}
