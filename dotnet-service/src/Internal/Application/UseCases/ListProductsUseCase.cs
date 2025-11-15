using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Application.UseCases;

public sealed class ListProductsUseCase(IProductRepository repository, ILoggerAdapter logger, CacheService cache)
{
    private readonly IProductRepository _repository = repository;
    private readonly ILoggerAdapter _logger = logger;
    private readonly CacheService _cache = cache;

    public async Task<IReadOnlyCollection<ProductResponseDto>> HandleAsync(CancellationToken cancellationToken)
    {
        async Task<IReadOnlyCollection<ProductResponseDto>> Loader(CancellationToken token)
        {
            var products = await _repository.ListAsync(token).ConfigureAwait(false);
            return products.Select(p => p.ToResponse()).ToList();
        }

        var cacheResult = await _cache.FetchAsync("products:list", Loader, cancellationToken).ConfigureAwait(false);
        var response = cacheResult.Value
            .Select(dto => dto with { CacheStatus = cacheResult.Status.ToString().ToLowerInvariant() })
            .ToList();
        _logger.Info("listed products", new { Count = response.Count, CacheStatus = cacheResult.Status });
        return response;
    }
}
