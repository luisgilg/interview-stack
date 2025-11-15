using DotnetService.Internal.Domain.Entities;
using DotnetService.Internal.Domain.Interfaces;
using DotnetService.Internal.Infrastructure.Repositories;
using Moq;

namespace DotnetService.Tests;

public sealed class ProductRepositoryTests
{
    [Fact]
    public async Task CreateAsync_UsesClockTimestampsAndPersistsThroughStore()
    {
        var now = new DateTime(2024, 01, 01, 0, 0, 0, DateTimeKind.Utc);
        var storeMock = new Mock<IProductStore>(MockBehavior.Strict);
        var clockMock = new Mock<IClock>();
        clockMock.Setup(c => c.UtcNow()).Returns(now);
        storeMock
            .Setup(s => s.CreateProductAsync(It.IsAny<Product>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync((Product product, CancellationToken _) => product);

        var repository = new ProductRepository(storeMock.Object, clockMock.Object);

        var created = await repository.CreateAsync(new Product { Name = "Keyboard", Price = 99 }, CancellationToken.None);

        storeMock.Verify(
            s => s.CreateProductAsync(
                It.Is<Product>(p => p.CreatedAt == now && p.UpdatedAt == now),
                It.IsAny<CancellationToken>()),
            Times.Once);
        Assert.Equal(now, created.CreatedAt);
        Assert.Equal(now, created.UpdatedAt);
    }

    [Fact]
    public async Task UpdateAsync_UsesClockWhenUpdatedAtMissing()
    {
        var now = new DateTime(2024, 05, 05, 12, 0, 0, DateTimeKind.Utc);
        var storeMock = new Mock<IProductStore>(MockBehavior.Strict);
        var clockMock = new Mock<IClock>();
        clockMock.Setup(c => c.UtcNow()).Returns(now);
        storeMock
            .Setup(s => s.UpdateProductAsync("product-1", It.IsAny<Product>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync((string _, Product product, CancellationToken __) => product);

        var repository = new ProductRepository(storeMock.Object, clockMock.Object);
        var updated = await repository.UpdateAsync("product-1", new Product { Name = "Keyboard", Price = 101 }, CancellationToken.None);

        storeMock.Verify(
            s => s.UpdateProductAsync(
                "product-1",
                It.Is<Product>(p => p.UpdatedAt == now),
                It.IsAny<CancellationToken>()),
            Times.Once);
        Assert.NotNull(updated);
        Assert.Equal(now, updated!.UpdatedAt);
    }
}
