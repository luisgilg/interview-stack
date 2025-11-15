using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Events;
using DotnetService.Internal.Domain.Errors;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Application.UseCases;

public sealed class DeleteProductUseCase(
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

    public async Task HandleAsync(string id, CancellationToken cancellationToken)
    {
        if (_writeBehindEnabled)
        {
            await EnsureProductExistsAsync(id, cancellationToken).ConfigureAwait(false);
            await _cache.DeleteAsync($"products:{id}", cancellationToken).ConfigureAwait(false);
            await _cache.DeleteAsync("products:list", cancellationToken).ConfigureAwait(false);
            var writeEvent = new WriteEvent(WriteEventType.Delete, id, null, _clock.UtcNow(), Source);
            await _queue.EnqueueAsync(writeEvent, cancellationToken).ConfigureAwait(false);
            _logger.Info("product enqueued for deletion", new { Id = id });
            return;
        }

        var deleted = await _repository.DeleteAsync(id, cancellationToken);
        if (!deleted)
        {
            throw DomainException.NotFound("product not found");
        }
        await _cache.DeleteAsync($"products:{id}", cancellationToken).ConfigureAwait(false);
        await _cache.DeleteAsync("products:list", cancellationToken).ConfigureAwait(false);
        _logger.Info("product deleted", new { Id = id });
    }

    private async Task EnsureProductExistsAsync(string id, CancellationToken cancellationToken)
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

        await _cache.FetchAsync($"products:{id}", Loader, cancellationToken).ConfigureAwait(false);
    }
}
