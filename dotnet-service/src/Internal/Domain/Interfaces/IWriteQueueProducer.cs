using DotnetService.Internal.Domain.Events;

namespace DotnetService.Internal.Domain.Interfaces;

public interface IWriteQueueProducer
{
    Task EnqueueAsync(WriteEvent @event, CancellationToken cancellationToken);
}
