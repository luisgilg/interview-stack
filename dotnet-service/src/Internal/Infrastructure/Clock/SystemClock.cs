using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Infrastructure.Clock;

public sealed class SystemClock : IClock
{
    public DateTime UtcNow() => DateTime.UtcNow;
}
