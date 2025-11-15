const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const DEFAULT_CONFIG_PATH = path.join(__dirname, '../../config.yaml');
const DURATION_UNITS = {
  ms: 1,
  s: 1000,
  m: 60_000
};

function parseDuration(value, fallback) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string') {
    const match = value.trim().match(/^(\d+(?:\.\d+)?)(ms|s|m)$/i);
    if (match) {
      const amount = parseFloat(match[1]);
      const unit = match[2].toLowerCase();
      return Math.round(amount * DURATION_UNITS[unit]);
    }
  }
  return fallback;
}

class RequestTimeouts {
  constructor(input = {}) {
    this.read = parseDuration(input.read, 2000);
    this.write = parseDuration(input.write, 5000);
    this.health = parseDuration(input.health, 2000);
  }
}

class ServerConfig {
  constructor(input = {}) {
    this.port = Number.isInteger(input.port) ? input.port : 8082;
    this.readTimeout = parseDuration(input.readTimeout, 5000);
    this.writeTimeout = parseDuration(input.writeTimeout, 5000);
    this.idleTimeout = parseDuration(input.idleTimeout, 120000);
    this.shutdownTimeout = parseDuration(input.shutdownTimeout, 5000);
    this.requestTimeouts = new RequestTimeouts(input.requestTimeouts);
  }
}

class PostgresConfig {
  constructor(input = {}) {
    this.host = input.host || 'postgres';
    this.port = Number.isInteger(input.port) ? input.port : 5432;
    this.user = input.user || 'postgres';
    this.password = input.password || 'postgres';
    this.database = input.db || input.database || 'productsdb';
    this.connectTimeout = parseDuration(input.connectTimeout, 5000);
  }
}

class MongoConfig {
  constructor(input = {}) {
    this.uri = input.uri || 'mongodb://mongo:27017';
    this.database = input.database || 'productsdb';
    this.collection = input.collection || 'products';
    this.connectTimeout = parseDuration(input.connectTimeout, 10000);
    this.operationTimeout = parseDuration(input.operationTimeout, 5000);
  }
}

class DatabaseConfig {
  constructor(input = {}) {
    const type = (input.type || 'sql').toLowerCase();
    this.type = type === 'mongo' ? 'mongo' : 'sql';
    this.postgres = new PostgresConfig(input.postgres);
    this.mongo = new MongoConfig(input.mongo);
  }
}

class RedisConfig {
  constructor(input = {}) {
    this.host = input.host || 'redis';
    this.port = Number.isInteger(input.port) ? input.port : 6379;
    this.password = typeof input.password === 'string' ? input.password : '';
    this.db = Number.isInteger(input.db) ? input.db : 0;
  }
}

class CacheConfig {
  constructor(input = {}) {
    this.enabled = input.enabled !== undefined ? Boolean(input.enabled) : true;
    this.defaultTTL = parseDuration(input.defaultTTL, 30000);
    this.staleTTL = parseDuration(input.staleTTL, 60000);
    this.redis = new RedisConfig(input.redis);
  }
}

class WriteBehindConfig {
  constructor(input = {}) {
    this.enabled = input.enabled !== undefined ? Boolean(input.enabled) : true;
    this.batchSize = Number.isInteger(input.batchSize) ? input.batchSize : 50;
    this.flushInterval = parseDuration(input.flushInterval, 1000);
    this.streamName = input.streamName || 'products_write_queue';
  }
}

class MetricsConfig {
  constructor(input = {}, fallbackPort = 8082) {
    this.enabled = input.enabled !== undefined ? Boolean(input.enabled) : true;
    const defaultPort = Number.isInteger(fallbackPort) ? fallbackPort : 8082;
    this.port = Number.isInteger(input.port) && input.port > 0 ? input.port : defaultPort;
    const rawPath = typeof input.path === 'string' ? input.path.trim() : '';
    if (!rawPath) {
      this.path = '/metrics';
    } else {
      this.path = rawPath.startsWith('/') ? rawPath : `/${rawPath}`;
    }
  }
}

class AppConfig {
  constructor(raw = {}) {
    this.server = new ServerConfig(raw.server);
    this.database = new DatabaseConfig(raw.database);
    this.cache = new CacheConfig(raw.cache);
    this.writeBehind = new WriteBehindConfig(raw.writeBehind);
    this.metrics = new MetricsConfig(raw.metrics, this.server.port);
  }

  validate() {
    if (this.server.port <= 0) {
      throw new Error('server.port must be greater than zero');
    }
    if (this.database.type !== 'sql' && this.database.type !== 'mongo') {
      throw new Error('database.type must be either sql or mongo');
    }
    if (this.cache.enabled) {
      if (this.cache.defaultTTL <= 0) {
        throw new Error('cache.defaultTTL must be greater than zero when cache is enabled');
      }
      if (this.cache.staleTTL <= 0) {
        throw new Error('cache.staleTTL must be greater than zero when cache is enabled');
      }
      if (!this.cache.redis.host) {
        throw new Error('cache.redis.host must be set when cache is enabled');
      }
      if (this.cache.redis.port <= 0) {
        throw new Error('cache.redis.port must be greater than zero when cache is enabled');
      }
    }
    if (this.writeBehind.enabled) {
      if (this.writeBehind.batchSize <= 0) {
        throw new Error('writeBehind.batchSize must be greater than zero when enabled');
      }
      if (this.writeBehind.flushInterval <= 0) {
        throw new Error('writeBehind.flushInterval must be greater than zero when enabled');
      }
      if (!this.writeBehind.streamName) {
        throw new Error('writeBehind.streamName must be set when enabled');
      }
    }
    if (this.metrics.enabled) {
      if (this.metrics.port <= 0) {
        throw new Error('metrics.port must be greater than zero when enabled');
      }
      if (!this.metrics.path || !this.metrics.path.startsWith('/')) {
        throw new Error('metrics.path must start with / when enabled');
      }
    }
    return this;
  }
}

function loadConfig(customPath) {
  const filePath = customPath || process.env.NODE_CONFIG_PATH || DEFAULT_CONFIG_PATH;
  const rawContent = fs.readFileSync(filePath, 'utf8');
  const parsed = yaml.load(rawContent) || {};
  return new AppConfig(parsed).validate();
}

module.exports = {
  loadConfig,
  AppConfig
};
