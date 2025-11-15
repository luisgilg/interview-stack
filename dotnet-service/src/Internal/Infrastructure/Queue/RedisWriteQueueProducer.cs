using System.Text.Json;
using System.Text.Json.Serialization;
using DotnetService.Internal.Config;
using DotnetService.Internal.Domain.Events;
using DotnetService.Internal.Domain.Interfaces;
using StackExchange.Redis;

namespace DotnetService.Internal.Infrastructure.Queue;

public sealed class RedisWriteQueueProducer : IWriteQueueProducer
{
    private readonly IConnectionMultiplexer _connection;
    private readonly string _streamName;
    private readonly JsonSerializerOptions _jsonOptions;

    public RedisWriteQueueProducer(IConnectionMultiplexer connection, AppConfig config)
    {
        _connection = connection;
        _streamName = config.WriteBehind.StreamName;
        _jsonOptions = new JsonSerializerOptions(JsonSerializerDefaults.Web);
        _jsonOptions.Converters.Add(new JsonStringEnumConverter(JsonNamingPolicy.CamelCase));
    }

    public async Task EnqueueAsync(WriteEvent @event, CancellationToken cancellationToken)
    {
        if (cancellationToken.IsCancellationRequested)
        {
            cancellationToken.ThrowIfCancellationRequested();
        }
        var db = _connection.GetDatabase();
        var payload = JsonSerializer.Serialize(@event, _jsonOptions);
        await db.StreamAddAsync(_streamName, new[] { new NameValueEntry("event", payload) }).ConfigureAwait(false);
    }
}
