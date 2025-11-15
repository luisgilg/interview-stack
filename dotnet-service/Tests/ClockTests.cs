using DotnetService.Internal.Infrastructure.Clock;

namespace DotnetService.Tests;

public sealed class ClockTests
{
    [Fact]
    public void FakeClock_ReturnsDeterministicValues()
    {
        var instant = new DateTime(2024, 03, 03, 0, 0, 0, DateTimeKind.Utc);
        var clock = new FakeClock(instant);

        Assert.Equal(instant, clock.UtcNow());

        clock.Advance(TimeSpan.FromMinutes(30));
        Assert.Equal(instant.AddMinutes(30), clock.UtcNow());

        var expected = new DateTime(2025, 01, 01, 12, 0, 0, DateTimeKind.Utc);
        clock.Set(expected);
        Assert.Equal(expected, clock.UtcNow());
    }
}
