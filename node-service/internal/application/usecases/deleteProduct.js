const DomainError = require('../../domain/errors');

class DeleteProductUseCase {
  constructor(repository, logger, clock, cacheCoordinator, writeQueue, writeBehindEnabled, source) {
    this.repository = repository;
    this.logger = logger;
    this.clock = clock;
    this.cache = cacheCoordinator;
    this.queue = writeQueue;
    this.writeBehind = Boolean(writeBehindEnabled);
    this.source = source || 'node-service';
  }

  async execute(id) {
    if (this.writeBehind && this.queue) {
      const existing = await this.repository.getById(id);
      if (!existing) {
        throw DomainError.notFound('product not found');
      }
      await this.cache?.deleteProduct?.(id);
      await this.enqueueEvent('delete', { id }, this.clock.now());
      this.logger.info('product enqueued for deletion', { id });
      return;
    }

    const deleted = await this.repository.delete(id);
    if (!deleted) {
      throw DomainError.notFound('product not found');
    }
    await this.cache?.deleteProduct?.(id);
    this.logger.info('product deleted', { id });
  }

  async enqueueEvent(type, product, timestamp) {
    const event = {
      type,
      id: product.id || product.ID,
      timestamp: timestamp?.toISOString?.() || new Date().toISOString(),
      source: this.source
    };
    await this.queue.enqueue(event);
  }
}

module.exports = DeleteProductUseCase;
