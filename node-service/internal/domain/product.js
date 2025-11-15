const DomainError = require('./errors');

class Product {
  constructor({ id, name, price, tags = [], createdAt, updatedAt }) {
    this.id = id;
    this.name = name;
    this.price = price;
    this.tags = tags || [];
    this.createdAt = createdAt;
    this.updatedAt = updatedAt;
  }

  validate() {
    if (!this.name || !this.name.trim()) {
      throw DomainError.validation('name is required');
    }
    if (typeof this.price !== 'number' || this.price <= 0) {
      throw DomainError.validation('price must be a positive number');
    }
    if (!Array.isArray(this.tags)) {
      this.tags = [];
    }
  }
}

module.exports = Product;
