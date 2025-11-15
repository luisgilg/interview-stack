using System.Threading;
using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.UseCases;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Errors;

namespace DotnetService.Internal.Interface.Http;

public sealed class ProductHttpHandler(
    ListProductsUseCase list,
    GetProductUseCase get,
    CreateProductUseCase create,
    UpdateProductUseCase update,
    DeleteProductUseCase delete,
    HealthCheckUseCase health,
    AppConfig config)
{
    private readonly ListProductsUseCase _list = list;
    private readonly GetProductUseCase _get = get;
    private readonly CreateProductUseCase _create = create;
    private readonly UpdateProductUseCase _update = update;
    private readonly DeleteProductUseCase _delete = delete;
    private readonly HealthCheckUseCase _health = health;
    private readonly RequestTimeoutsConfig _timeouts = config.Server.RequestTimeouts;

    public async Task<IResult> ListAsync(CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CreateTimeoutSource(cancellationToken, RequestType.Read);
            var result = await _list.HandleAsync(timeout.Token);
            return Results.Json(result);
        }
        catch (DomainException ex)
        {
            return MapError(ex);
        }
        catch (Exception ex)
        {
            return InternalError(ex);
        }
    }

    public async Task<IResult> GetAsync(string id, CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CreateTimeoutSource(cancellationToken, RequestType.Read);
            var product = await _get.HandleAsync(id, timeout.Token);
            return Results.Json(product);
        }
        catch (DomainException ex)
        {
            return MapError(ex);
        }
        catch (Exception ex)
        {
            return InternalError(ex);
        }
    }

    public async Task<IResult> CreateAsync(ProductRequestDto request, CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CreateTimeoutSource(cancellationToken, RequestType.Write);
            var product = await _create.HandleAsync(request, timeout.Token);
            return Results.Created($"/products/{product.Id}", product);
        }
        catch (DomainException ex)
        {
            return MapError(ex);
        }
        catch (Exception ex)
        {
            return InternalError(ex);
        }
    }

    public async Task<IResult> UpdateAsync(string id, ProductRequestDto request, CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CreateTimeoutSource(cancellationToken, RequestType.Write);
            var product = await _update.HandleAsync(id, request, timeout.Token);
            return Results.Json(product);
        }
        catch (DomainException ex)
        {
            return MapError(ex);
        }
        catch (Exception ex)
        {
            return InternalError(ex);
        }
    }

    public async Task<IResult> DeleteAsync(string id, CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CreateTimeoutSource(cancellationToken, RequestType.Write);
            await _delete.HandleAsync(id, timeout.Token);
            return Results.NoContent();
        }
        catch (DomainException ex)
        {
            return MapError(ex);
        }
        catch (Exception ex)
        {
            return InternalError(ex);
        }
    }

    public async Task<IResult> HealthAsync(CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CreateTimeoutSource(cancellationToken, RequestType.Health);
            await _health.HandleAsync(timeout.Token);
            return Results.Json(new HealthResponseDto("ok"));
        }
        catch (Exception ex)
        {
            return Results.Json(new ErrorResponseDto(ex.Message), statusCode: StatusCodes.Status503ServiceUnavailable);
        }
    }

    private static IResult MapError(DomainException exception) =>
        exception.Code switch
        {
            DomainErrorCode.Validation => Results.BadRequest(new ErrorResponseDto(exception.Message)),
            DomainErrorCode.NotFound => Results.NotFound(new ErrorResponseDto(exception.Message)),
            _ => Results.Json(new ErrorResponseDto(exception.Message), statusCode: StatusCodes.Status500InternalServerError)
        };

    private static IResult InternalError(Exception exception) =>
        Results.Json(new ErrorResponseDto(exception.Message), statusCode: StatusCodes.Status500InternalServerError);

    private enum RequestType
    {
        Read,
        Write,
        Health
    }

    private CancellationTokenSource CreateTimeoutSource(CancellationToken cancellationToken, RequestType type)
    {
        var deadline = type switch
        {
            RequestType.Read => _timeouts.Read,
            RequestType.Write => _timeouts.Write,
            RequestType.Health => _timeouts.Health,
            _ => _timeouts.Write
        };
        var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeout.CancelAfter(deadline);
        return timeout;
    }
}
