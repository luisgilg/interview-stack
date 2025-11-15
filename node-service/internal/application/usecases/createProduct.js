const { v4: uuid } = require('uuid');
const { toDomain, toResponse } = require('../dto/productDto');

class CreateProductUseCase {
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

  async execute(input) {
    const product = toDomain(input);
    product.validate();
    const now = this.clock.now();
    product.createdAt = now;
    product.updatedAt = now;

    if (this.writeBehind && this.queue) {
      product.id = product.id || uuid();
      const response = toResponse(product);
      await this.cache?.upsertProduct(response);
      await this.enqueueEvent('create', product, now);
      this.logger.info('product enqueued for creation', { id: product.id });
      return response;
    }

    const created = await this.repository.create(product);
    const response = toResponse(created);
    await this.cache?.upsertProduct(response);
    this.logger.info('product created', { id: created.id });
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

module.exports = CreateProductUseCase;
