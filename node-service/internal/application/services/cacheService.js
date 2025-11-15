const metrics = require('../../observability/metrics');

const CacheStatuses = {
  BYPASS: 'bypass',
  MISS: 'miss',
  FRESH: 'fresh',
  STALE: 'stale'
};

class CacheService {
  constructor(client, config = {}, logger = console, clock = () => Date.now()) {
    this.client = client;
    this.logger = logger;
    this.clock = clock;
    this.enabled = Boolean(config.enabled) && Boolean(client);
    this.defaultTTL = Number(config.defaultTTL) || 0;
    this.staleTTL = Number(config.staleTTL) || 0;
  }

  isEnabled() {
    return this.enabled && this.defaultTTL > 0 && this.client;
  }

  async fetch(key, loader) {
    if (!this.isEnabled() || !key || typeof loader !== 'function') {
      const value = await loader();
      return this.track({ value, status: CacheStatuses.BYPASS });
    }

    try {
      const cached = await this.client.get(key);
      if (cached) {
        const parsed = JSON.parse(cached);
        const age = this.clock() - Number(parsed.stored_at || 0);
        if (age <= this.defaultTTL) {
          return this.track({ value: parsed.payload, status: CacheStatuses.FRESH });
        }
        if (age <= this.defaultTTL + this.staleTTL) {
          this.refresh(key, loader);
          return this.track({ value: parsed.payload, status: CacheStatuses.STALE });
        }
      }
    } catch (err) {
      this.logger?.warn?.('cache fetch failed', { key, error: err.message });
      const value = await loader();
      return this.track({ value, status: CacheStatuses.BYPASS });
    }

    return this.loadAndStore(key, loader);
  }

  async loadAndStore(key, loader) {
    const value = await loader();
    try {
      await this.store(key, value);
    } catch (err) {
      this.logger?.warn?.('cache store failed', { key, error: err.message });
    }
    return this.track({ value, status: CacheStatuses.MISS });
  }

  async store(key, value) {
    if (!this.isEnabled()) {
      return;
    }
    const ttl = this.defaultTTL + this.staleTTL;
    if (ttl <= 0) {
      return;
    }
    const entry = JSON.stringify({
      payload: value,
      stored_at: this.clock()
    });
    await this.client.set(key, entry, ttl);
  }

  async write(key, value) {
    try {
      await this.store(key, value);
    } catch (err) {
      this.logger?.warn?.('cache store failed', { key, error: err.message });
    }
  }

  async invalidate(key) {
    if (!this.isEnabled() || typeof this.client?.del !== 'function') {
      return;
    }
    try {
      await this.client.del(key);
    } catch (err) {
      this.logger?.warn?.('cache delete failed', { key, error: err.message });
    }
  }

  refresh(key, loader) {
    setImmediate(async () => {
      try {
        const value = await loader();
        await this.store(key, value);
      } catch (err) {
        this.logger?.warn?.('cache refresh failed', { key, error: err.message });
      }
    });
  }

  track(result) {
    metrics.recordCacheStatus(result.status);
    return result;
  }
}

module.exports = {
  CacheService,
  CacheStatuses
};
