using MongoDB.Bson.Serialization.Attributes;

namespace DotnetService.Internal.Infrastructure.NoSql.Models;

public sealed class ProductDocument
{
    [BsonId]
    [BsonElement("_id")]
    public string Id { get; set; } = string.Empty;

    [BsonElement("name")]
    public string Name { get; set; } = string.Empty;

    [BsonElement("price")]
    public decimal Price { get; set; }

    [BsonElement("tags")]
    public List<string> Tags { get; set; } = new();

    [BsonElement("created_at")]
    public DateTime CreatedAt { get; set; }

    [BsonElement("updated_at")]
    public DateTime UpdatedAt { get; set; }
}
