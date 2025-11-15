using System;
using DotnetService.Internal.Config;
using Npgsql;

namespace DotnetService.Internal.Infrastructure.Sql;

public static class PostgresFactory
{
    public static NpgsqlDataSource CreateDataSource(AppConfig config)
    {
        var pg = config.Database.Postgres;
        var connectionString = new NpgsqlConnectionStringBuilder
        {
            Host = pg.Host,
            Port = pg.Port,
            Username = pg.User,
            Password = pg.Password,
            Database = pg.Db,
            Timeout = (int)Math.Ceiling(pg.ConnectTimeout.TotalSeconds),
            CommandTimeout = (int)Math.Ceiling(config.Server.RequestTimeouts.Write.TotalSeconds),
            Pooling = true
        }.ConnectionString;

        var builder = new NpgsqlDataSourceBuilder(connectionString);
        return builder.Build();
    }
}
