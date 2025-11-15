const DomainError = require('../../domain/errors');

class ProductRepository {
  constructor(store, clock) {
    this.store = store;
    this.clock = clock;
  }

  async list() {
    try {
      return await this.store.listProducts();
    } catch (err) {
      throw DomainError.internal('failed to list products', err);
    }
  }

  async getById(id) {
    try {
      return await this.store.getProduct(id);
    } catch (err) {
      throw DomainError.internal('failed to fetch product', err);
    }
  }

  async create(product) {
    try {
      const now = this.clock?.now();
      if (product.createdAt == null && now) {
        product.createdAt = now;
      }
      if (product.updatedAt == null && now) {
        product.updatedAt = now;
      }
      return await this.store.createProduct(product);
    } catch (err) {
      throw DomainError.internal('failed to create product', err);
    }
  }

  async update(id, product) {
    try {
      if (product.updatedAt == null && this.clock) {
        product.updatedAt = this.clock.now();
      }
      return await this.store.updateProduct(id, product);
    } catch (err) {
      throw DomainError.internal('failed to update product', err);
    }
  }

  async delete(id) {
    try {
      return await this.store.deleteProduct(id);
    } catch (err) {
      throw DomainError.internal('failed to delete product', err);
    }
  }

  async health() {
    await this.store.health();
  }
}

module.exports = ProductRepository;
