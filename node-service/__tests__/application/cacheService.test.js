const { CacheService } = require('../../internal/application/services/cacheService');

class FakeClient {
  constructor() {
    this.store = new Map();
    this.setCalls = 0;
  }

  async get(key) {
    return this.store.get(key) ?? null;
  }

  async set(key, value) {
    this.setCalls += 1;
    this.store.set(key, value);
  }
}

function encodeEntry(payload, storedAt) {
  return JSON.stringify({
    payload,
    stored_at: storedAt
  });
}

describe('CacheService', () => {
  it('returns bypass when disabled', async () => {
    const service = new CacheService(null, { enabled: false }, {});
    const loader = jest.fn().mockResolvedValue({ hello: 'world' });

    const result = await service.fetch('key', loader);

    expect(result.status).toBe('bypass');
    expect(result.value).toEqual({ hello: 'world' });
    expect(loader).toHaveBeenCalledTimes(1);
  });

  it('stores values on cache miss', async () => {
    const client = new FakeClient();
    let now = Date.now();
    const service = new CacheService(
      client,
      { enabled: true, defaultTTL: 1000, staleTTL: 1000 },
      {},
      () => now
    );
    const loader = jest.fn().mockResolvedValue({ value: 42 });

    const result = await service.fetch('products:list', loader);

    expect(result.status).toBe('miss');
    expect(result.value).toEqual({ value: 42 });
    expect(client.store.has('products:list')).toBe(true);
  });

  it('returns fresh values when cache entry is valid', async () => {
    const client = new FakeClient();
    const base = Date.now();
    client.store.set('products:list', encodeEntry({ cached: true }, base));
    let now = base + 500;
    const service = new CacheService(
      client,
      { enabled: true, defaultTTL: 1000, staleTTL: 1000 },
      {},
      () => now
    );
    const loader = jest.fn();

    const result = await service.fetch('products:list', loader);

    expect(result.status).toBe('fresh');
    expect(result.value).toEqual({ cached: true });
    expect(loader).not.toHaveBeenCalled();
  });

  it('returns stale entries and revalidates in background', async () => {
    const client = new FakeClient();
    const base = Date.now();
    client.store.set('products:list', encodeEntry({ cached: 'old' }, base - 1500));
    let now = base;
    const service = new CacheService(
      client,
      { enabled: true, defaultTTL: 1000, staleTTL: 1000 },
      {},
      () => now
    );
    const loader = jest.fn().mockResolvedValue({ cached: 'new' });

    const result = await service.fetch('products:list', loader);

    expect(result.status).toBe('stale');
    expect(result.value).toEqual({ cached: 'old' });
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(loader).toHaveBeenCalled();
  });
});
