using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Events;
using DotnetService.Internal.Domain.Errors;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Application.UseCases;

public sealed class UpdateProductUseCase(
    IProductRepository repository,
    ILoggerAdapter logger,
    IClock clock,
    CacheService cache,
    IWriteQueueProducer queue,
    AppConfig config)
{
    private readonly IProductRepository _repository = repository;
    private readonly ILoggerAdapter _logger = logger;
    private readonly IClock _clock = clock;
    private readonly CacheService _cache = cache;
    private readonly IWriteQueueProducer _queue = queue;
    private readonly bool _writeBehindEnabled = config.WriteBehind.Enabled;
    private const string Source = "dotnet-service";

    public async Task<ProductResponseDto> HandleAsync(string id, ProductRequestDto request, CancellationToken cancellationToken)
    {
        var now = _clock.UtcNow();
        var product = request.ToDomain(id).Validate() with { UpdatedAt = now };

        if (_writeBehindEnabled)
        {
            var current = await LoadCurrentProductAsync(id, cancellationToken).ConfigureAwait(false);
            var pending = current with
            {
                Name = product.Name,
                Price = product.Price,
                Tags = product.Tags,
                UpdatedAt = now
            };
            var response = pending.ToResponse();
            await _cache.WriteAsync($"products:{id}", response, cancellationToken).ConfigureAwait(false);
            await _cache.DeleteAsync("products:list", cancellationToken).ConfigureAwait(false);
            var writeEvent = new WriteEvent(WriteEventType.Update, id, pending, now, Source);
            await _queue.EnqueueAsync(writeEvent, cancellationToken).ConfigureAwait(false);
            _logger.Info("product enqueued for update", new { Id = id });
            return response;
        }

        var updated = await _repository.UpdateAsync(id, product, cancellationToken);
        if (updated is null)
        {
            throw DomainException.NotFound("product not found");
        }
        var updatedResponse = updated.ToResponse();
        await _cache.WriteAsync($"products:{id}", updatedResponse, cancellationToken).ConfigureAwait(false);
        await _cache.DeleteAsync("products:list", cancellationToken).ConfigureAwait(false);
        _logger.Info("product updated", new { Id = id });
        return updatedResponse;
    }

    private async Task<DotnetService.Internal.Domain.Entities.Product> LoadCurrentProductAsync(string id, CancellationToken cancellationToken)
    {
        async Task<ProductResponseDto> Loader(CancellationToken token)
        {
            var existing = await _repository.GetByIdAsync(id, token).ConfigureAwait(false);
            if (existing is null)
            {
                throw DomainException.NotFound("product not found");
            }
            return existing.ToResponse();
        }

        var cacheResult = await _cache.FetchAsync($"products:{id}", Loader, cancellationToken).ConfigureAwait(false);
        return cacheResult.Value.ToDomain();
    }
}
