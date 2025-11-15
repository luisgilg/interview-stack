using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Domain.Errors;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Application.UseCases;

public sealed class GetProductUseCase(IProductRepository repository, ILoggerAdapter logger, CacheService cache)
{
    private readonly IProductRepository _repository = repository;
    private readonly ILoggerAdapter _logger = logger;
    private readonly CacheService _cache = cache;

    public async Task<ProductResponseDto> HandleAsync(string id, CancellationToken cancellationToken)
    {
        async Task<ProductResponseDto> Loader(CancellationToken token)
        {
            var product = await _repository.GetByIdAsync(id, token).ConfigureAwait(false);
            if (product is null)
            {
                throw DomainException.NotFound("product not found");
            }
            return product.ToResponse();
        }

        var cacheResult = await _cache.FetchAsync($"products:{id}", Loader, cancellationToken).ConfigureAwait(false);
        var response = cacheResult.Value with { CacheStatus = cacheResult.Status.ToString().ToLowerInvariant() };
        _logger.Info("retrieved product", new { Id = id, CacheStatus = cacheResult.Status });
        return response;
    }
}
