using DotnetService.Internal.Domain.Events;
using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Infrastructure.Queue;

public sealed class NoopWriteQueueProducer : IWriteQueueProducer
{
    public Task EnqueueAsync(WriteEvent @event, CancellationToken cancellationToken)
    {
        return Task.CompletedTask;
    }
}
