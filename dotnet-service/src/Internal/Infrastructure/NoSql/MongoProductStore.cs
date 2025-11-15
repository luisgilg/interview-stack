using DotnetService.Internal.Domain.Entities;
using DotnetService.Internal.Domain.Interfaces;
using DotnetService.Internal.Infrastructure.NoSql.Models;
using MongoDB.Bson;
using MongoDB.Driver;

namespace DotnetService.Internal.Infrastructure.NoSql;

public sealed class MongoProductStore(IMongoCollection<ProductDocument> collection) : IProductStore
{
    private readonly IMongoCollection<ProductDocument> _collection = collection;

    public async Task<IReadOnlyCollection<Product>> ListProductsAsync(CancellationToken cancellationToken)
    {
        var docs = await _collection.Find(FilterDefinition<ProductDocument>.Empty)
            .SortByDescending(p => p.CreatedAt)
            .ToListAsync(cancellationToken);
        return docs.Select(ToDomain).ToList();
    }

    public async Task<Product?> GetProductAsync(string id, CancellationToken cancellationToken)
    {
        var doc = await _collection.Find(Builders<ProductDocument>.Filter.Eq(p => p.Id, id))
            .FirstOrDefaultAsync(cancellationToken);
        return doc is null ? null : ToDomain(doc);
    }

    public async Task<Product> CreateProductAsync(Product product, CancellationToken cancellationToken)
    {
        var doc = new ProductDocument
        {
            Id = string.IsNullOrWhiteSpace(product.Id) ? Guid.NewGuid().ToString() : product.Id,
            Name = product.Name,
            Price = product.Price,
            Tags = product.Tags?.ToList() ?? new List<string>(),
            CreatedAt = product.CreatedAt,
            UpdatedAt = product.UpdatedAt
        };

        await _collection.InsertOneAsync(doc, cancellationToken: cancellationToken);
        return ToDomain(doc);
    }

    public async Task<Product?> UpdateProductAsync(string id, Product product, CancellationToken cancellationToken)
    {
        var update = Builders<ProductDocument>.Update
            .Set(p => p.Name, product.Name)
            .Set(p => p.Price, product.Price)
            .Set(p => p.Tags, product.Tags?.ToList() ?? new List<string>())
            .Set(p => p.UpdatedAt, product.UpdatedAt);

        var options = new FindOneAndUpdateOptions<ProductDocument> { ReturnDocument = ReturnDocument.After };
        var doc = await _collection.FindOneAndUpdateAsync(
            Builders<ProductDocument>.Filter.Eq(p => p.Id, id),
            update,
            options,
            cancellationToken);

        return doc is null ? null : ToDomain(doc);
    }

    public async Task<bool> DeleteProductAsync(string id, CancellationToken cancellationToken)
    {
        var result = await _collection.DeleteOneAsync(Builders<ProductDocument>.Filter.Eq(p => p.Id, id), cancellationToken);
        return result.DeletedCount > 0;
    }

    public async Task EnsureIndexesAsync(CancellationToken cancellationToken)
    {
        var keys = Builders<ProductDocument>.IndexKeys.Ascending(p => p.Name);
        try
        {
            await _collection.Indexes.CreateOneAsync(new CreateIndexModel<ProductDocument>(keys), cancellationToken: cancellationToken);
        }
        catch (MongoCommandException ex) when (ex.Code == 85 || string.Equals(ex.CodeName, "IndexOptionsConflict", StringComparison.OrdinalIgnoreCase))
        {
            // index already exists, ignore
        }
    }

    public async Task HealthAsync(CancellationToken cancellationToken)
    {
        await _collection.Database.RunCommandAsync<BsonDocument>(new BsonDocument("ping", 1), cancellationToken: cancellationToken);
    }

    private static Product ToDomain(ProductDocument doc) =>
        new()
        {
            Id = doc.Id,
            Name = doc.Name,
            Price = doc.Price,
            Tags = doc.Tags?.ToArray() ?? Array.Empty<string>(),
            CreatedAt = doc.CreatedAt,
            UpdatedAt = doc.UpdatedAt
        };
}
