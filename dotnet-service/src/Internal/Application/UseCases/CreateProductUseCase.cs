using System;
using DotnetService.Internal.Application.DTOs;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Events;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Application.UseCases;

public sealed class CreateProductUseCase(
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

    public async Task<ProductResponseDto> HandleAsync(ProductRequestDto request, CancellationToken cancellationToken)
    {
        var now = _clock.UtcNow();
        var product = request.ToDomain().Validate() with { CreatedAt = now, UpdatedAt = now };

        if (_writeBehindEnabled)
        {
            var pending = product with { Id = string.IsNullOrWhiteSpace(product.Id) ? Guid.NewGuid().ToString() : product.Id };
            var response = pending.ToResponse();
            await _cache.WriteAsync($"products:{pending.Id}", response, cancellationToken).ConfigureAwait(false);
            await _cache.DeleteAsync("products:list", cancellationToken).ConfigureAwait(false);
            var writeEvent = new WriteEvent(WriteEventType.Create, pending.Id, pending, now, Source);
            await _queue.EnqueueAsync(writeEvent, cancellationToken).ConfigureAwait(false);
            _logger.Info("product enqueued for creation", new { pending.Id });
            return response;
        }

        var created = await _repository.CreateAsync(product, cancellationToken);
        var createdResponse = created.ToResponse();
        await _cache.WriteAsync($"products:{created.Id}", createdResponse, cancellationToken).ConfigureAwait(false);
        await _cache.DeleteAsync("products:list", cancellationToken).ConfigureAwait(false);
        _logger.Info("product created", new { created.Id });
        return createdResponse;
    }
}
