using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Application.UseCases;

public sealed class HealthCheckUseCase(IProductRepository repository)
{
    private readonly IProductRepository _repository = repository;

    public Task HandleAsync(CancellationToken cancellationToken) => _repository.HealthAsync(cancellationToken);
}
