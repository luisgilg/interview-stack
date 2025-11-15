using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Infrastructure.Clock;

public sealed class FakeClock(DateTime? initial = null) : IClock
{
    private DateTime _current = initial ?? DateTime.UnixEpoch;

    public DateTime UtcNow() => _current;

    public void Set(DateTime instant)
    {
        _current = instant;
    }

    public void Advance(TimeSpan amount)
    {
        _current = _current.Add(amount);
    }
}
