const { RedisConnection } = require('./redisConnection');

class RedisClient {
  constructor(config = {}, logger = console) {
    this.host = config.host || '127.0.0.1';
    this.port = Number.isInteger(config.port) ? config.port : 6379;
    this.password = config.password || '';
    this.db = Number.isInteger(config.db) ? config.db : 0;
    this.logger = logger;
  }

  async ping() {
    const connection = this.createConnection();
    try {
      await this.authenticate(connection);
      const response = await connection.send(['PING']);
      return response;
    } finally {
      connection.close();
    }
  }

  async get(key) {
    const connection = this.createConnection();
    try {
      await this.authenticate(connection);
      const result = await connection.send(['GET', key]);
      return result;
    } finally {
      connection.close();
    }
  }

  async set(key, value, ttlMs) {
    const connection = this.createConnection();
    try {
      await this.authenticate(connection);
      const args = ['SET', key, value];
      if (Number.isFinite(ttlMs) && ttlMs > 0) {
        args.push('PX', Math.trunc(ttlMs).toString());
      }
      await connection.send(args);
    } finally {
      connection.close();
    }
  }

  async del(key) {
    const connection = this.createConnection();
    try {
      await this.authenticate(connection);
      await connection.send(['DEL', key]);
    } finally {
      connection.close();
    }
  }

  createConnection() {
    return new RedisConnection({ host: this.host, port: this.port });
  }

  async authenticate(connection) {
    if (this.password) {
      await connection.send(['AUTH', this.password]);
    }
    if (this.db > 0) {
      await connection.send(['SELECT', this.db.toString()]);
    }
  }
}

module.exports = RedisClient;
