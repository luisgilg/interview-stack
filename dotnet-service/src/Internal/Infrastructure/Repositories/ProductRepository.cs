using DotnetService.Internal.Domain.Entities;
using DotnetService.Internal.Domain.Errors;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Infrastructure.Repositories;

public sealed class ProductRepository(IProductStore store, IClock clock) : IProductRepository
{
    private readonly IProductStore _store = store;
    private readonly IClock _clock = clock;

    public async Task<IReadOnlyCollection<Product>> ListAsync(CancellationToken cancellationToken)
    {
        try
        {
            return await _store.ListProductsAsync(cancellationToken);
        }
        catch (Exception ex)
        {
            throw DomainException.Internal("failed to list products", ex);
        }
    }

    public async Task<Product?> GetByIdAsync(string id, CancellationToken cancellationToken)
    {
        try
        {
            return await _store.GetProductAsync(id, cancellationToken);
        }
        catch (Exception ex)
        {
            throw DomainException.Internal("failed to fetch product", ex);
        }
    }

    public async Task<Product> CreateAsync(Product product, CancellationToken cancellationToken)
    {
        product = EnsureCreateTimestamps(product);
        try
        {
            return await _store.CreateProductAsync(product, cancellationToken);
        }
        catch (Exception ex)
        {
            throw DomainException.Internal("failed to create product", ex);
        }
    }

    public async Task<Product?> UpdateAsync(string id, Product product, CancellationToken cancellationToken)
    {
        product = EnsureUpdateTimestamp(product);
        try
        {
            return await _store.UpdateProductAsync(id, product, cancellationToken);
        }
        catch (Exception ex)
        {
            throw DomainException.Internal("failed to update product", ex);
        }
    }

    public async Task<bool> DeleteAsync(string id, CancellationToken cancellationToken)
    {
        try
        {
            return await _store.DeleteProductAsync(id, cancellationToken);
        }
        catch (Exception ex)
        {
            throw DomainException.Internal("failed to delete product", ex);
        }
    }

    public async Task HealthAsync(CancellationToken cancellationToken)
    {
        await _store.HealthAsync(cancellationToken);
    }

    private Product EnsureCreateTimestamps(Product product)
    {
        var createdAt = product.CreatedAt == default ? _clock.UtcNow() : product.CreatedAt;
        var updatedAt = product.UpdatedAt == default ? createdAt : product.UpdatedAt;
        return product with { CreatedAt = createdAt, UpdatedAt = updatedAt };
    }

    private Product EnsureUpdateTimestamp(Product product)
    {
        var updatedAt = product.UpdatedAt == default ? _clock.UtcNow() : product.UpdatedAt;
        return product with { UpdatedAt = updatedAt };
    }
}
