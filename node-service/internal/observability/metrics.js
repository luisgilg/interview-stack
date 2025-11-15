const http = require('http');
const client = require('prom-client');

class ObservabilityMetrics {
  constructor() {
    this.serviceName = 'node-service';
    this.enabled = true;
    this.path = '/metrics';
    this.registry = new client.Registry();
    this.metricsServer = null;
    this.initialized = false;
    this.httpRequestsTotal = null;
    this.httpRequestDurationSeconds = null;
    this.cacheHitsTotal = null;
    this.cacheMissesTotal = null;
    this.writeBehindLagSeconds = null;
    this.writeBehindQueueLength = null;
    this.writeBehindBatchSize = null;
    this.writeBehindBatchDurationSeconds = null;
    this.writeBehindErrorsTotal = null;
  }

  configure({ serviceName, enabled = true, path = '/metrics' } = {}) {
    this.serviceName = serviceName || this.serviceName;
    this.enabled = enabled;
    this.path = this.normalizePath(path);
    if (this.initialized) {
      return;
    }
    this.initialized = true;
    this.registry.setDefaultLabels({ service: this.serviceName });
    client.collectDefaultMetrics({ register: this.registry });

    this.httpRequestsTotal = new client.Counter({
      name: 'http_requests_total',
      help: 'Total HTTP requests handled by the API',
      labelNames: ['method', 'route', 'status'],
      registers: [this.registry]
    });
    this.httpRequestDurationSeconds = new client.Histogram({
      name: 'http_request_duration_seconds',
      help: 'Duration of HTTP requests in seconds',
      labelNames: ['method', 'route'],
      buckets: client.exponentialBuckets(0.005, 2, 10),
      registers: [this.registry]
    });
    this.cacheHitsTotal = new client.Counter({
      name: 'cache_hits_total',
      help: 'Total number of cache hits',
      labelNames: ['status'],
      registers: [this.registry]
    });
    this.cacheMissesTotal = new client.Counter({
      name: 'cache_misses_total',
      help: 'Total number of cache misses or bypass operations',
      labelNames: ['status'],
      registers: [this.registry]
    });
    this.writeBehindLagSeconds = new client.Gauge({
      name: 'write_behind_lag_seconds',
      help: 'Age of the oldest message in the latest processed batch',
      registers: [this.registry]
    });
    this.writeBehindQueueLength = new client.Gauge({
      name: 'write_behind_queue_length',
      help: 'Current Redis stream length for write-behind operations',
      registers: [this.registry]
    });
    this.writeBehindBatchSize = new client.Gauge({
      name: 'write_behind_batch_size',
      help: 'Number of events processed in the latest batch',
      registers: [this.registry]
    });
    this.writeBehindBatchDurationSeconds = new client.Histogram({
      name: 'write_behind_batch_duration_seconds',
      help: 'Processing latency for write-behind batches',
      buckets: client.exponentialBuckets(0.01, 2, 10),
      registers: [this.registry]
    });
    this.writeBehindErrorsTotal = new client.Counter({
      name: 'write_behind_errors_total',
      help: 'Total number of write-behind errors',
      registers: [this.registry]
    });
  }

  normalizePath(path) {
    if (typeof path !== 'string' || path.trim() === '') {
      return '/metrics';
    }
    return path.startsWith('/') ? path : `/${path}`;
  }

  httpMiddleware() {
    return (req, res, next) => {
      if (!this.enabled) {
        next();
        return;
      }
      const start = process.hrtime.bigint();
      res.once('finish', () => {
        const duration = Number(process.hrtime.bigint() - start) / 1e9;
        const route = (req.route?.path || req.path || req.url || 'unknown').toString();
        const status = res.statusCode || 0;
        this.httpRequestsTotal?.inc({ method: req.method, route, status: String(status) });
        this.httpRequestDurationSeconds?.observe({ method: req.method, route }, duration);
      });
      next();
    };
  }

  expressHandler() {
    return async (_req, res) => {
      if (!this.enabled) {
        res.status(404).json({ error: 'metrics disabled' });
        return;
      }
      try {
        const payload = await this.registry.metrics();
        res.set('Content-Type', this.registry.contentType);
        res.send(payload);
      } catch (err) {
        res.status(500).json({ error: 'failed to collect metrics' });
      }
    };
  }

  startStandaloneServer(port, logger = console) {
    if (!this.enabled || !port || port <= 0 || this.metricsServer) {
      return;
    }
    const handler = async (req, res) => {
      if (req.url === this.path) {
        try {
          const payload = await this.registry.metrics();
          res.writeHead(200, { 'Content-Type': this.registry.contentType });
          res.end(payload);
        } catch (_err) {
          res.writeHead(500);
          res.end('metrics collection failed');
        }
        return;
      }
      res.writeHead(404);
      res.end();
    };
    this.metricsServer = http.createServer(handler);
    this.metricsServer.listen(port, () => {
      logger?.info?.('metrics server listening', { port, path: this.path });
    });
  }

  async stopStandaloneServer(logger = console) {
    if (!this.metricsServer) {
      return;
    }
    await new Promise((resolve) => this.metricsServer.close(resolve));
    logger?.info?.('metrics server stopped');
    this.metricsServer = null;
  }

  recordCacheStatus(status) {
    if (!this.enabled || !status) {
      return;
    }
    const normalized = String(status).toLowerCase();
    if (normalized === 'fresh' || normalized === 'stale') {
      this.cacheHitsTotal?.inc({ status: normalized });
    } else {
      this.cacheMissesTotal?.inc({ status: normalized });
    }
  }

  recordWriteBehindBatch({ size = 0, durationSeconds = 0, lagSeconds, queueLength } = {}) {
    if (!this.enabled || size <= 0) {
      return;
    }
    this.writeBehindBatchSize?.set(size);
    this.writeBehindBatchDurationSeconds?.observe(durationSeconds);
    if (typeof lagSeconds === 'number' && lagSeconds >= 0) {
      this.writeBehindLagSeconds?.set(lagSeconds);
    }
    if (typeof queueLength === 'number' && queueLength >= 0) {
      this.writeBehindQueueLength?.set(queueLength);
    }
  }

  recordWriteBehindError() {
    if (!this.enabled) {
      return;
    }
    this.writeBehindErrorsTotal?.inc();
  }
}

module.exports = new ObservabilityMetrics();
