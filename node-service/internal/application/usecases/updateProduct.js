const DomainError = require('../../domain/errors');
const { toDomain, toResponse } = require('../dto/productDto');

class UpdateProductUseCase {
  /**
   * @param {import('../../domain/productRepository')} repository
   * @param {import('../../domain/logger')} logger
   * @param {import('../../core/clock/Clock')} clock
   * @param {import('../services/cacheCoordinator')} cacheCoordinator
   * @param {import('../services/writeQueue')} writeQueue
   */
  constructor(repository, logger, clock, cacheCoordinator, writeQueue, writeBehindEnabled, source) {
    this.repository = repository;
    this.logger = logger;
    this.clock = clock;
    this.cache = cacheCoordinator;
    this.queue = writeQueue;
    this.writeBehind = Boolean(writeBehindEnabled);
    this.source = source || 'node-service';
  }

  async execute(id, input) {
    const changes = toDomain({ ...input, id });
    changes.validate();
    const now = this.clock.now();
    changes.updatedAt = now;

    if (this.writeBehind && this.queue) {
      const current = await this.repository.getById(id);
      if (!current) {
        throw DomainError.notFound('product not found');
      }
      current.name = changes.name;
      current.price = changes.price;
      current.tags = changes.tags;
      current.updatedAt = now;
      const response = toResponse(current);
      await this.cache?.upsertProduct(response);
      await this.enqueueEvent('update', current, now);
      this.logger.info('product enqueued for update', { id });
      return response;
    }

    const updated = await this.repository.update(id, changes);
    if (!updated) {
      throw DomainError.notFound('product not found');
    }
    const response = toResponse(updated);
    await this.cache?.upsertProduct(response);
    this.logger.info('product updated', { id });
    return response;
  }

  async enqueueEvent(type, product, timestamp) {
    const event = {
      type,
      id: product.id,
      payload: product,
      timestamp: timestamp?.toISOString?.() || new Date().toISOString(),
      source: this.source
    };
    await this.queue.enqueue(event);
  }
}

module.exports = UpdateProductUseCase;
