using DotnetService.Internal.Domain.Interfaces;

namespace DotnetService.Internal.Infrastructure.Logging;

public sealed class LoggerAdapter(ILoggerFactory factory) : ILoggerAdapter
{
    private readonly ILogger _logger = factory.CreateLogger("Application");

    public void Info(string message, object? context = null)
    {
        _logger.LogInformation("{Message} {@Context}", message, context);
    }

    public void Warn(string message, object? context = null)
    {
        _logger.LogWarning("{Message} {@Context}", message, context);
    }

    public void Error(Exception exception, string message, object? context = null)
    {
        _logger.LogError(exception, "{Message} {@Context}", message, context);
    }
}
