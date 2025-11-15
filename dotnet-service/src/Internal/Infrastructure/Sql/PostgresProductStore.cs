using Dapper;
using DotnetService.Internal.Domain.Entities;
using DotnetService.Internal.Domain.Interfaces;
using Npgsql;

namespace DotnetService.Internal.Infrastructure.Sql;

public sealed class PostgresProductStore(NpgsqlDataSource dataSource) : IProductStore
{
    private readonly NpgsqlDataSource _dataSource = dataSource;

    public async Task<IReadOnlyCollection<Product>> ListProductsAsync(CancellationToken cancellationToken)
    {
        const string sql = @"SELECT id::text AS Id, name, price, tags, created_at AS CreatedAt, updated_at AS UpdatedAt
                             FROM products ORDER BY created_at DESC";
        await using var connection = await _dataSource.OpenConnectionAsync(cancellationToken);
        var rows = await connection.QueryAsync<ProductRow>(new CommandDefinition(sql, cancellationToken: cancellationToken));
        return rows.Select(MapProduct).ToList();
    }

    public async Task<Product?> GetProductAsync(string id, CancellationToken cancellationToken)
    {
        if (!Guid.TryParse(id, out var productId))
        {
            return null;
        }
        const string sql = @"SELECT id::text AS Id, name, price, tags, created_at AS CreatedAt, updated_at AS UpdatedAt
                             FROM products WHERE id = @Id";
        await using var connection = await _dataSource.OpenConnectionAsync(cancellationToken);
        var row = await connection.QuerySingleOrDefaultAsync<ProductRow>(
            new CommandDefinition(sql, new { Id = productId }, cancellationToken: cancellationToken));
        return row is null ? null : MapProduct(row);
    }

    public async Task<Product> CreateProductAsync(Product product, CancellationToken cancellationToken)
    {
        var id = Guid.NewGuid();
        var tags = product.Tags?.ToArray() ?? Array.Empty<string>();
        const string sql = @"INSERT INTO products (id, name, price, tags, created_at, updated_at)
                             VALUES (@Id, @Name, @Price, @Tags, @CreatedAt, @UpdatedAt)
                             RETURNING id::text AS Id, name, price, tags, created_at AS CreatedAt, updated_at AS UpdatedAt";
        await using var connection = await _dataSource.OpenConnectionAsync(cancellationToken);
        var row = await connection.QuerySingleAsync<ProductRow>(
            new CommandDefinition(sql, new
            {
                Id = id,
                product.Name,
                product.Price,
                Tags = tags,
                CreatedAt = product.CreatedAt,
                UpdatedAt = product.UpdatedAt
            }, cancellationToken: cancellationToken));
        return MapProduct(row);
    }

    public async Task<Product?> UpdateProductAsync(string id, Product product, CancellationToken cancellationToken)
    {
        if (!Guid.TryParse(id, out var productId))
        {
            return null;
        }
        const string sql = @"UPDATE products
                             SET name = @Name, price = @Price, tags = @Tags, updated_at = @UpdatedAt
                             WHERE id = @Id
                             RETURNING id::text AS Id, name, price, tags, created_at AS CreatedAt, updated_at AS UpdatedAt";
        await using var connection = await _dataSource.OpenConnectionAsync(cancellationToken);
        var row = await connection.QuerySingleOrDefaultAsync<ProductRow>(
            new CommandDefinition(sql, new
            {
                Id = productId,
                product.Name,
                product.Price,
                Tags = product.Tags?.ToArray() ?? Array.Empty<string>(),
                UpdatedAt = product.UpdatedAt
            }, cancellationToken: cancellationToken));
        return row is null ? null : MapProduct(row);
    }

    public async Task<bool> DeleteProductAsync(string id, CancellationToken cancellationToken)
    {
        if (!Guid.TryParse(id, out var productId))
        {
            return false;
        }
        const string sql = "DELETE FROM products WHERE id = @Id";
        await using var connection = await _dataSource.OpenConnectionAsync(cancellationToken);
        var rows = await connection.ExecuteAsync(new CommandDefinition(sql, new { Id = productId }, cancellationToken: cancellationToken));
        return rows > 0;
    }

    public async Task HealthAsync(CancellationToken cancellationToken)
    {
        await using var connection = await _dataSource.OpenConnectionAsync(cancellationToken);
        await connection.ExecuteScalarAsync<int>(new CommandDefinition("SELECT 1", cancellationToken: cancellationToken));
    }

    private static Product MapProduct(ProductRow row) =>
        new()
        {
            Id = row.Id,
            Name = row.Name,
            Price = row.Price,
            Tags = row.Tags?.ToArray() ?? Array.Empty<string>(),
            CreatedAt = row.CreatedAt,
            UpdatedAt = row.UpdatedAt
        };

    private sealed record ProductRow
    {
        public string Id { get; init; } = string.Empty;
        public string Name { get; init; } = string.Empty;
        public decimal Price { get; init; }
        public string[]? Tags { get; init; }
        public DateTime CreatedAt { get; init; }
        public DateTime UpdatedAt { get; init; }
    }
}
