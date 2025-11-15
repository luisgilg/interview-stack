class CacheCoordinator {
  constructor(cacheService, logger = console) {
    this.cache = cacheService;
    this.logger = logger;
  }

  get enabled() {
    return Boolean(this.cache?.isEnabled());
  }

  async upsertProduct(product) {
    if (!this.enabled || !product?.id) {
      return;
    }
    const key = `products:${product.id}`;
    try {
      await this.cache.write(key, product);
    } catch (err) {
      this.logger?.warn?.('failed to update product cache', { key, error: err.message });
    }
    await this.invalidateList();
  }

  async deleteProduct(id) {
    if (!this.cache || !id) {
      return;
    }
    const key = `products:${id}`;
    try {
      await this.cache.invalidate(key);
    } catch (err) {
      this.logger?.warn?.('failed to delete product cache', { key, error: err.message });
    }
    await this.invalidateList();
  }

  async invalidateList() {
    if (!this.enabled) {
      return;
    }
    try {
      await this.cache.invalidate('products:list');
    } catch (err) {
      this.logger?.warn?.('failed to invalidate product list cache', { error: err.message });
    }
  }
}

module.exports = CacheCoordinator;
