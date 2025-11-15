using DotnetService.Internal.Domain.Entities;

namespace DotnetService.Internal.Domain.Interfaces;

public interface IProductRepository
{
    Task<IReadOnlyCollection<Product>> ListAsync(CancellationToken cancellationToken);
    Task<Product?> GetByIdAsync(string id, CancellationToken cancellationToken);
    Task<Product> CreateAsync(Product product, CancellationToken cancellationToken);
    Task<Product?> UpdateAsync(string id, Product product, CancellationToken cancellationToken);
    Task<bool> DeleteAsync(string id, CancellationToken cancellationToken);
    Task HealthAsync(CancellationToken cancellationToken);
}
