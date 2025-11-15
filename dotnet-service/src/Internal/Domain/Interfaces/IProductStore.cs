using DotnetService.Internal.Domain.Entities;

namespace DotnetService.Internal.Domain.Interfaces;

public interface IProductStore
{
    Task<IReadOnlyCollection<Product>> ListProductsAsync(CancellationToken cancellationToken);
    Task<Product?> GetProductAsync(string id, CancellationToken cancellationToken);
    Task<Product> CreateProductAsync(Product product, CancellationToken cancellationToken);
    Task<Product?> UpdateProductAsync(string id, Product product, CancellationToken cancellationToken);
    Task<bool> DeleteProductAsync(string id, CancellationToken cancellationToken);
    Task HealthAsync(CancellationToken cancellationToken);
}
