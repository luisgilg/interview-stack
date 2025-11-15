const { Pool } = require('pg');

function createPool(pgConfig, statementTimeout) {
  return new Pool({
    host: pgConfig.host,
    port: pgConfig.port,
    user: pgConfig.user,
    password: pgConfig.password,
    database: pgConfig.database,
    connectionTimeoutMillis: pgConfig.connectTimeout,
    statement_timeout: statementTimeout,
    query_timeout: statementTimeout
  });
}

module.exports = {
  createPool
};
