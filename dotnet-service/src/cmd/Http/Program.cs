using System.Diagnostics;
using DotnetService.Internal.Application.Services;
using DotnetService.Internal.Application.UseCases;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Interfaces;
using DotnetService.Internal.Infrastructure.Cache;
using DotnetService.Internal.Infrastructure.Clock;
using DotnetService.Internal.Infrastructure.Logging;
using DotnetService.Internal.Infrastructure.NoSql;
using DotnetService.Internal.Infrastructure.Queue;
using DotnetService.Internal.Infrastructure.Repositories;
using DotnetService.Internal.Infrastructure.Sql;
using DotnetService.Internal.Interface.Http;
using DotnetService.Internal.Observability;
using Microsoft.AspNetCore.Routing;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Microsoft.OpenApi.Models;
using Prometheus;
using StackExchange.Redis;

var builder = WebApplication.CreateBuilder(args);
builder.Logging.ClearProviders();
builder.Logging.AddConsole();
const string serviceName = "dotnet-service";

builder.Services.AddOptions<AppConfig>()
    .Bind(builder.Configuration.GetSection("App"))
    .ValidateDataAnnotations()
    .Validate(config => config.Database.Type is "sql" or "mongo", "Database type must be sql or mongo")
    .ValidateOnStart();

builder.Services.AddSingleton(sp => sp.GetRequiredService<IOptions<AppConfig>>().Value);
builder.Services.AddScoped<IClock, SystemClock>();

var cacheEnabled = builder.Configuration.GetValue<bool>("App:Cache:Enabled");
var writeBehindEnabled = builder.Configuration.GetValue<bool>("App:WriteBehind:Enabled");

if (cacheEnabled || writeBehindEnabled)
{
    builder.Services.AddSingleton<IConnectionMultiplexer>(sp =>
    {
        var cacheConfig = sp.GetRequiredService<AppConfig>().Cache;
        var options = new ConfigurationOptions
        {
            EndPoints = { $"{cacheConfig.Redis.Host}:{cacheConfig.Redis.Port}" },
            Password = string.IsNullOrWhiteSpace(cacheConfig.Redis.Password) ? null : cacheConfig.Redis.Password,
            DefaultDatabase = cacheConfig.Redis.Database,
            AbortOnConnectFail = false
        };
        return ConnectionMultiplexer.Connect(options);
    });
}

if (cacheEnabled)
{
    builder.Services.AddSingleton<ICacheClient>(sp =>
    {
        var cacheConfig = sp.GetRequiredService<AppConfig>().Cache;
        var loggerAdapter = sp.GetRequiredService<ILoggerAdapter>();
        var multiplexer = sp.GetRequiredService<IConnectionMultiplexer>();
        return new RedisCacheClient(multiplexer, cacheConfig, loggerAdapter);
    });
}
else
{
    builder.Services.AddSingleton<ICacheClient, NoopCacheClient>();
}

builder.Services.AddSingleton(provider =>
{
    var cfg = provider.GetRequiredService<AppConfig>().Cache;
    var client = provider.GetRequiredService<ICacheClient>();
    var loggerAdapter = provider.GetRequiredService<ILoggerAdapter>();
    var clock = provider.GetRequiredService<IClock>();
    return new CacheService(client, cfg, loggerAdapter, clock);
});

if (writeBehindEnabled)
{
    builder.Services.AddSingleton<IWriteQueueProducer>(sp =>
    {
        var multiplexer = sp.GetRequiredService<IConnectionMultiplexer>();
        var appCfg = sp.GetRequiredService<AppConfig>();
        return new RedisWriteQueueProducer(multiplexer, appCfg);
    });
    builder.Services.AddHostedService<WriteBehindWorker>();
}
else
{
    builder.Services.AddSingleton<IWriteQueueProducer, NoopWriteQueueProducer>();
}

builder.Services.AddSingleton(provider => PostgresFactory.CreateDataSource(provider.GetRequiredService<AppConfig>()));
builder.Services.AddSingleton<PostgresProductStore>();
builder.Services.AddSingleton(provider => MongoContext.CreateCollection(provider.GetRequiredService<AppConfig>()));
builder.Services.AddSingleton<MongoProductStore>();
builder.Services.AddSingleton<IProductStore>(sp =>
{
    var cfg = sp.GetRequiredService<AppConfig>();
    return cfg.Database.Type == "mongo"
        ? sp.GetRequiredService<MongoProductStore>()
        : sp.GetRequiredService<PostgresProductStore>();
});

builder.Services.AddScoped<IProductRepository, ProductRepository>();
builder.Services.AddSingleton<ILoggerAdapter, LoggerAdapter>();

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(options =>
{
    options.SwaggerDoc("v1", new OpenApiInfo
    {
        Title = "Products API (.NET)",
        Version = "v1",
        Description = "Minimal API that powers the .NET slice of the interview stack."
    });
});

builder.Services.AddScoped<ListProductsUseCase>();
builder.Services.AddScoped<GetProductUseCase>();
builder.Services.AddScoped<CreateProductUseCase>();
builder.Services.AddScoped<UpdateProductUseCase>();
builder.Services.AddScoped<DeleteProductUseCase>();
builder.Services.AddScoped<HealthCheckUseCase>();
builder.Services.AddScoped<ProductHttpHandler>();

builder.WebHost.ConfigureKestrel((context, options) =>
{
    var serverConfig = context.Configuration.GetSection("App:Server").Get<ServerConfig>();
    if (serverConfig is null)
    {
        return;
    }

    options.Limits.RequestHeadersTimeout = serverConfig.ReadTimeout;
    options.Limits.KeepAliveTimeout = serverConfig.IdleTimeout;
});

var app = builder.Build();
var appConfig = app.Services.GetRequiredService<AppConfig>();
MetricsRegistry.Initialize(serviceName, appConfig.Metrics.Enabled);
if (appConfig.Database.Type == "mongo")
{
    var mongoStore = app.Services.GetRequiredService<MongoProductStore>();
    await mongoStore.EnsureIndexesAsync(CancellationToken.None);
}
if (appConfig.Metrics.Enabled && appConfig.Metrics.Port != appConfig.Server.Port)
{
    app.Logger.LogWarning("metrics.port differs from server.port; exposing metrics on main HTTP port {Port}", appConfig.Server.Port);
}

app.Use(async (context, next) =>
{
    var stopwatch = Stopwatch.StartNew();
    await next.Invoke();
    stopwatch.Stop();
    var route = context.GetEndpoint() is RouteEndpoint endpoint
        ? endpoint.RoutePattern.RawText ?? context.Request.Path.Value ?? "unknown"
        : context.Request.Path.Value ?? "unknown";
    MetricsRegistry.ObserveHttpRequest(context.Request.Method, route ?? "unknown", context.Response.StatusCode, stopwatch.Elapsed);
});

app.UseSwagger();
app.UseSwaggerUI(options =>
{
    options.SwaggerEndpoint("/swagger/v1/swagger.json", "Products API (.NET) v1");
    options.RoutePrefix = "swagger";
});

if (appConfig.Metrics.Enabled)
{
    app.MapMetrics(appConfig.Metrics.Path);
}

app.MapProductEndpoints();
app.Run($"http://0.0.0.0:{appConfig.Server.Port}");
