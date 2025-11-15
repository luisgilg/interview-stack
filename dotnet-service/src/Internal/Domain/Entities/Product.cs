using DotnetService.Internal.Domain.Errors;

namespace DotnetService.Internal.Domain.Entities;

public sealed record Product
{
    public string Id { get; init; } = string.Empty;
    public string Name { get; init; } = string.Empty;
    public decimal Price { get; init; }
    public IReadOnlyCollection<string> Tags { get; init; } = Array.Empty<string>();
    public DateTime CreatedAt { get; init; }
    public DateTime UpdatedAt { get; init; }

    public Product Validate()
    {
        if (string.IsNullOrWhiteSpace(Name))
        {
            throw DomainException.Validation("name is required");
        }

        if (Price <= 0)
        {
            throw DomainException.Validation("price must be positive");
        }

        return this;
    }
}
