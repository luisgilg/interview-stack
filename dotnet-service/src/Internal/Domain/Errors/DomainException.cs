namespace DotnetService.Internal.Domain.Errors;

public enum DomainErrorCode
{
    Validation,
    NotFound,
    Internal
}

public sealed class DomainException : Exception
{
    public DomainErrorCode Code { get; }

    private DomainException(DomainErrorCode code, string message, Exception? inner = null) : base(message, inner)
    {
        Code = code;
    }

    public static DomainException Validation(string message) => new(DomainErrorCode.Validation, message);

    public static DomainException NotFound(string message) => new(DomainErrorCode.NotFound, message);

    public static DomainException Internal(string message, Exception? inner = null) =>
        new(DomainErrorCode.Internal, message, inner);
}
