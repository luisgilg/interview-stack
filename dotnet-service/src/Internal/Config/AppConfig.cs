using System;
using System.ComponentModel.DataAnnotations;
using Microsoft.Extensions.Options;

namespace DotnetService.Internal.Config;

public sealed class RequestTimeoutsConfig
{
    [Required]
    public TimeSpan Read { get; init; } = TimeSpan.FromSeconds(2);

    [Required]
    public TimeSpan Write { get; init; } = TimeSpan.FromSeconds(5);

    [Required]
    public TimeSpan Health { get; init; } = TimeSpan.FromSeconds(2);
}

public sealed class ServerConfig
{
    [Range(1, 65535)]
    public int Port { get; init; } = 8083;

    [Range(typeof(TimeSpan), "00:00:01", "12:00:00")]
    public TimeSpan ReadTimeout { get; init; } = TimeSpan.FromSeconds(5);

    [Range(typeof(TimeSpan), "00:00:01", "12:00:00")]
    public TimeSpan WriteTimeout { get; init; } = TimeSpan.FromSeconds(5);

    [Range(typeof(TimeSpan), "00:00:01", "12:00:00")]
    public TimeSpan IdleTimeout { get; init; } = TimeSpan.FromMinutes(2);

    [Range(typeof(TimeSpan), "00:00:01", "12:00:00")]
    public TimeSpan ShutdownTimeout { get; init; } = TimeSpan.FromSeconds(5);

    public RequestTimeoutsConfig RequestTimeouts { get; init; } = new();
}

public sealed class PostgresConfig
{
    [Required]
    public string Host { get; init; } = "postgres";

    [Range(1, 65535)]
    public int Port { get; init; } = 5432;

    [Required]
    public string User { get; init; } = "postgres";

    [Required]
    public string Password { get; init; } = "postgres";

    [Required]
    public string Db { get; init; } = "productsdb";

    [Range(typeof(TimeSpan), "00:00:01", "00:05:00")]
    public TimeSpan ConnectTimeout { get; init; } = TimeSpan.FromSeconds(5);
}

public sealed class MongoConfig
{
    [Required]
    public string Uri { get; init; } = "mongodb://mongo:27017";

    [Required]
    public string Database { get; init; } = "productsdb";

    [Required]
    public string Collection { get; init; } = "products";

    [Range(typeof(TimeSpan), "00:00:01", "00:05:00")]
    public TimeSpan ConnectTimeout { get; init; } = TimeSpan.FromSeconds(10);

    [Range(typeof(TimeSpan), "00:00:01", "00:05:00")]
    public TimeSpan OperationTimeout { get; init; } = TimeSpan.FromSeconds(5);
}

public sealed class DatabaseConfig
{
    private string _type = "sql";

    [Required]
    public string Type
    {
        get => _type;
        init => _type = value?.Equals("mongo", StringComparison.OrdinalIgnoreCase) == true ? "mongo" : "sql";
    }

    public PostgresConfig Postgres { get; init; } = new();

    public MongoConfig Mongo { get; init; } = new();
}

public sealed class RedisConfig
{
    [Required]
    public string Host { get; init; } = "redis";

    [Range(1, 65535)]
    public int Port { get; init; } = 6379;

    public string? Password { get; init; }

    [Range(0, 16)]
    public int Database { get; init; }
}

public sealed class CacheConfig
{
    public bool Enabled { get; init; } = true;

    [Range(typeof(TimeSpan), "00:00:01", "1.00:00:00")]
    public TimeSpan DefaultTtl { get; init; } = TimeSpan.FromSeconds(30);

    [Range(typeof(TimeSpan), "00:00:01", "1.00:00:00")]
    public TimeSpan StaleTtl { get; init; } = TimeSpan.FromMinutes(1);

    [Required]
    public RedisConfig Redis { get; init; } = new();
}

public sealed class WriteBehindConfig
{
    public bool Enabled { get; init; } = true;

    [Range(1, 1000)]
    public int BatchSize { get; init; } = 50;

    [Range(typeof(TimeSpan), "00:00:01", "1.00:00:00")]
    public TimeSpan FlushInterval { get; init; } = TimeSpan.FromSeconds(1);

    [Required]
    public string StreamName { get; init; } = "products_write_queue";
}

public sealed class MetricsConfig
{
    public bool Enabled { get; init; } = true;

    [Range(1, 65535)]
    public int Port { get; init; } = 8083;

    [Required]
    [RegularExpression("^/.+")]
    public string Path { get; init; } = "/metrics";
}

public sealed class AppConfig
{
    public ServerConfig Server { get; init; } = new();

    public DatabaseConfig Database { get; init; } = new();

    public CacheConfig Cache { get; init; } = new();

    public WriteBehindConfig WriteBehind { get; init; } = new();

    public MetricsConfig Metrics { get; init; } = new();
}
