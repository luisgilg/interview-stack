const { createClient } = require('redis');
const { loadConfig } = require('../../internal/config/env');
const { createLogger } = require('../../internal/infrastructure/logging/pinoLogger');
const { createPool } = require('../../internal/infrastructure/sql/postgresClient');
const PostgresProductStore = require('../../internal/infrastructure/sql/postgresProductStore');
const { connect: connectMongo, disconnect: disconnectMongo } = require('../../internal/infrastructure/nosql/mongoClient');
const MongoProductStore = require('../../internal/infrastructure/nosql/mongoProductStore');
const ProductRepository = require('../../internal/infrastructure/repository/productRepository');
const SystemClock = require('../../internal/core/clock/SystemClock');

const ListProductsUseCase = require('../../internal/application/usecases/listProducts');
const GetProductUseCase = require('../../internal/application/usecases/getProduct');
const CreateProductUseCase = require('../../internal/application/usecases/createProduct');
const UpdateProductUseCase = require('../../internal/application/usecases/updateProduct');
const DeleteProductUseCase = require('../../internal/application/usecases/deleteProduct');
const HealthCheckUseCase = require('../../internal/application/usecases/healthCheck');
const ProductController = require('../../internal/interface/http/productController');
const { createApp } = require('../../internal/interface/http/app');
const RedisClient = require('../../internal/infrastructure/cache/redisClient');
const { CacheService } = require('../../internal/application/services/cacheService');
const CacheCoordinator = require('../../internal/application/services/cacheCoordinator');
const WriteQueueProducer = require('../../internal/application/services/writeQueue');
const WriteBehindWorker = require('../../internal/application/workers/writeBehindWorker');
const metrics = require('../../internal/observability/metrics');

async function bootstrap() {
  const config = loadConfig();
  const logger = createLogger();
  const clock = new SystemClock();
  const serviceName = 'node-service';
  metrics.configure({ serviceName, enabled: config.metrics.enabled, path: config.metrics.path });

  let pool;
  let closeDatabase = async () => {};
  let repository;
  let store;
  if (config.database.type === 'mongo') {
    await connectMongo(config.database.mongo);
    store = new MongoProductStore(config.database.mongo.operationTimeout);
    await store.ensureIndexes().catch((err) => logger.warn('failed to create indexes', { error: err.message }));
    repository = new ProductRepository(store, clock);
    closeDatabase = async () => disconnectMongo();
  } else {
    pool = createPool(config.database.postgres, config.server.requestTimeouts.write);
    store = new PostgresProductStore(pool);
    repository = new ProductRepository(store, clock);
    closeDatabase = async () => {
      if (pool) {
        await pool.end();
      }
    };
  }

  let cacheService = new CacheService(null, { enabled: false }, logger);
  if (config.cache.enabled) {
    const redisClient = new RedisClient(config.cache.redis, logger);
    try {
      await redisClient.ping();
      cacheService = new CacheService(
        redisClient,
        {
          enabled: true,
          defaultTTL: config.cache.defaultTTL,
          staleTTL: config.cache.staleTTL
        },
        logger
      );
      logger.info('redis cache connected', {
        host: config.cache.redis.host,
        port: config.cache.redis.port
      });
    } catch (err) {
      logger.warn('redis cache disabled', { error: err.message });
    }
  }

  const cacheCoordinator = new CacheCoordinator(cacheService, logger);

  let queueClient;
  let writeQueue = null;
  let stopWorker = async () => {};
  let closeQueue = async () => {};
  if (config.writeBehind.enabled) {
    queueClient = createClient({
      socket: { host: config.cache.redis.host, port: config.cache.redis.port },
      password: config.cache.redis.password || undefined,
      database: config.cache.redis.db || 0
    });
    queueClient.on('error', (err) => logger.warn('redis stream error', { error: err.message }));
    await queueClient.connect();
    writeQueue = new WriteQueueProducer(queueClient, config.writeBehind.streamName, logger);
    closeQueue = async () => {
      try {
        await queueClient.quit();
      } catch (err) {
        logger.warn('failed to close redis stream client', { error: err.message });
      }
    };
    const worker = new WriteBehindWorker({
      client: queueClient,
      streamName: config.writeBehind.streamName,
      group: serviceName,
      consumer: `${serviceName}-${process.pid}-${Date.now()}`,
      batchSize: config.writeBehind.batchSize,
      blockMs: config.writeBehind.flushInterval,
      store,
      logger,
      source: serviceName
    });
    await worker.start();
    stopWorker = () => worker.stop();
  }

  const useCases = {
    list: new ListProductsUseCase(repository, logger, cacheService),
    get: new GetProductUseCase(repository, logger, cacheService),
    create: new CreateProductUseCase(repository, logger, clock, cacheCoordinator, writeQueue, config.writeBehind.enabled, serviceName),
    update: new UpdateProductUseCase(repository, logger, clock, cacheCoordinator, writeQueue, config.writeBehind.enabled, serviceName),
    remove: new DeleteProductUseCase(repository, logger, clock, cacheCoordinator, writeQueue, config.writeBehind.enabled, serviceName),
    health: new HealthCheckUseCase(repository)
  };

  const controller = new ProductController(useCases, {
    read: config.server.requestTimeouts.read,
    write: config.server.requestTimeouts.write,
    health: config.server.requestTimeouts.health
  });
  const app = createApp(config, controller);
  let stopMetricsServer = async () => {};
  if (config.metrics.enabled) {
    if (config.metrics.port === config.server.port) {
      app.get(config.metrics.path, metrics.expressHandler());
    } else {
      metrics.startStandaloneServer(config.metrics.port, logger);
      stopMetricsServer = () => metrics.stopStandaloneServer(logger);
    }
  }

  const server = app.listen(config.server.port, () => {
    logger.info('node-service listening', {
      port: config.server.port,
      dbType: config.database.type
    });
  });
  server.headersTimeout = config.server.readTimeout;
  server.requestTimeout = config.server.writeTimeout;
  server.keepAliveTimeout = config.server.idleTimeout;

  const shutdownHandler = () => shutdown(server, closeDatabase, logger, config, stopWorker, closeQueue, stopMetricsServer);
  process.on('SIGTERM', shutdownHandler);
  process.on('SIGINT', shutdownHandler);
}

async function shutdown(server, closeDatabase, logger, config, stopWorker, closeQueue, stopMetricsServer) {
  logger.info('node-service shutting down');
  const forceExit = setTimeout(() => {
    logger.error('forced shutdown after timeout');
    process.exit(1);
  }, config.server.shutdownTimeout);
  await new Promise((resolve) => server.close(resolve));
  await stopWorker();
  await closeQueue();
  await closeDatabase();
  await stopMetricsServer();
  clearTimeout(forceExit);
  logger.info('node-service stopped');
  process.exit(0);
}

bootstrap().catch((err) => {
  console.error('failed to start node-service', err);
  process.exit(1);
});
