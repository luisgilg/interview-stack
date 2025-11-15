using DotnetService.Internal.Application.DTOs;
using Microsoft.AspNetCore.Mvc;

namespace DotnetService.Internal.Interface.Http;

public static class ProductEndpoints
{
    public static IEndpointRouteBuilder MapProductEndpoints(this IEndpointRouteBuilder endpoints)
    {
        endpoints.MapGet("/health", Health)
            .WithSummary("Health check")
            .WithDescription("Returns the health status of the service.")
            .Produces<HealthResponseDto>(StatusCodes.Status200OK)
            .Produces<ErrorResponseDto>(StatusCodes.Status503ServiceUnavailable);

        endpoints.MapGet("/products", ListProducts)
            .WithSummary("Gets all products")
            .WithDescription("Lists every product stored in the backing repository.")
            .Produces<IEnumerable<ProductResponseDto>>(StatusCodes.Status200OK)
            .Produces<ErrorResponseDto>(StatusCodes.Status500InternalServerError);

        endpoints.MapGet("/products/{id}", GetProduct)
            .WithSummary("Gets a product by ID")
            .WithDescription("Returns a single product given its identifier.")
            .Produces<ProductResponseDto>(StatusCodes.Status200OK)
            .Produces<ErrorResponseDto>(StatusCodes.Status404NotFound)
            .Produces<ErrorResponseDto>(StatusCodes.Status500InternalServerError);

        endpoints.MapPost("/products", CreateProduct)
            .WithSummary("Creates a product")
            .WithDescription("Persists a new product and returns the enriched representation.")
            .Produces<ProductResponseDto>(StatusCodes.Status201Created)
            .Produces<ErrorResponseDto>(StatusCodes.Status400BadRequest)
            .Produces<ErrorResponseDto>(StatusCodes.Status500InternalServerError);

        endpoints.MapPut("/products/{id}", UpdateProduct)
            .WithSummary("Updates a product")
            .WithDescription("Updates a product in place and returns the latest details.")
            .Produces<ProductResponseDto>(StatusCodes.Status200OK)
            .Produces<ErrorResponseDto>(StatusCodes.Status400BadRequest)
            .Produces<ErrorResponseDto>(StatusCodes.Status404NotFound)
            .Produces<ErrorResponseDto>(StatusCodes.Status500InternalServerError);

        endpoints.MapDelete("/products/{id}", DeleteProduct)
            .WithSummary("Deletes a product")
            .WithDescription("Removes a product permanently.")
            .Produces(StatusCodes.Status204NoContent)
            .Produces<ErrorResponseDto>(StatusCodes.Status404NotFound)
            .Produces<ErrorResponseDto>(StatusCodes.Status500InternalServerError);

        return endpoints;
    }

    [ProducesResponseType(typeof(HealthResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status503ServiceUnavailable)]
    private static Task<IResult> Health(ProductHttpHandler handler, CancellationToken ct) =>
        handler.HealthAsync(ct);

    [ProducesResponseType(typeof(IEnumerable<ProductResponseDto>), StatusCodes.Status200OK)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status500InternalServerError)]
    private static Task<IResult> ListProducts(ProductHttpHandler handler, CancellationToken ct) =>
        handler.ListAsync(ct);

    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status404NotFound)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status500InternalServerError)]
    private static Task<IResult> GetProduct(ProductHttpHandler handler, string id, CancellationToken ct) =>
        handler.GetAsync(id, ct);

    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status201Created)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status400BadRequest)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status500InternalServerError)]
    private static Task<IResult> CreateProduct(ProductHttpHandler handler, ProductRequestDto dto, CancellationToken ct) =>
        handler.CreateAsync(dto, ct);

    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status400BadRequest)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status404NotFound)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status500InternalServerError)]
    private static Task<IResult> UpdateProduct(ProductHttpHandler handler, string id, ProductRequestDto dto, CancellationToken ct) =>
        handler.UpdateAsync(id, dto, ct);

    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status404NotFound)]
    [ProducesResponseType(typeof(ErrorResponseDto), StatusCodes.Status500InternalServerError)]
    private static Task<IResult> DeleteProduct(ProductHttpHandler handler, string id, CancellationToken ct) =>
        handler.DeleteAsync(id, ct);
}
