using DotnetService.Internal.Config;
using DotnetService.Internal.Infrastructure.NoSql.Models;
using MongoDB.Driver;

namespace DotnetService.Internal.Infrastructure.NoSql;

public static class MongoContext
{
    public static IMongoCollection<ProductDocument> CreateCollection(AppConfig config)
    {
        var mongo = config.Database.Mongo;
        var settings = MongoClientSettings.FromConnectionString(mongo.Uri);
        settings.ServerSelectionTimeout = mongo.ConnectTimeout;
        settings.ConnectTimeout = mongo.ConnectTimeout;
        settings.SocketTimeout = mongo.OperationTimeout;
        var client = new MongoClient(settings);
        var database = client.GetDatabase(mongo.Database);
        return database.GetCollection<ProductDocument>(mongo.Collection);
    }
}
