using DotnetService.Internal.Domain.Entities;

namespace DotnetService.Internal.Application.DTOs;

public sealed record ProductRequestDto(string Name, decimal Price, IReadOnlyCollection<string>? Tags);

public sealed record ProductResponseDto(
    string Id,
    string Name,
    decimal Price,
    IReadOnlyCollection<string> Tags,
    DateTime CreatedAt,
    DateTime UpdatedAt,
    string? CacheStatus = null);

public sealed record ErrorResponseDto(string Error);

public sealed record HealthResponseDto(string Status);

public static class ProductDtoMapper
{
    public static Product ToDomain(this ProductRequestDto dto, string? id = null) =>
        new()
        {
            Id = id ?? string.Empty,
            Name = dto.Name,
            Price = dto.Price,
            Tags = dto.Tags?.ToArray() ?? Array.Empty<string>(),
            CreatedAt = default,
            UpdatedAt = default
        };

    public static ProductResponseDto ToResponse(this Product product) =>
        new(product.Id, product.Name, product.Price, product.Tags, product.CreatedAt, product.UpdatedAt);

    public static Product ToDomain(this ProductResponseDto dto) =>
        new()
        {
            Id = dto.Id,
            Name = dto.Name,
            Price = dto.Price,
            Tags = dto.Tags?.ToArray() ?? Array.Empty<string>(),
            CreatedAt = dto.CreatedAt,
            UpdatedAt = dto.UpdatedAt
        };
}
