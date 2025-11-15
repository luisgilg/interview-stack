namespace DotnetService.Internal.Domain.Interfaces;

public interface ILoggerAdapter
{
    void Info(string message, object? context = null);
    void Warn(string message, object? context = null);
    void Error(Exception exception, string message, object? context = null);
}
