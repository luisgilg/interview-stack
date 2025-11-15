using DotnetService.Internal.Domain.Entities;

namespace DotnetService.Internal.Domain.Events;

public enum WriteEventType
{
    Create,
    Update,
    Delete
}

public sealed record WriteEvent(WriteEventType Type, string Id, Product? Payload, DateTime Timestamp, string Source);
