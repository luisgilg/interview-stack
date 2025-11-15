const { jest } = require('@jest/globals');

class ProductStoreMock {
  constructor() {
    this.listProducts = jest.fn();
    this.getProduct = jest.fn();
    this.createProduct = jest.fn();
    this.updateProduct = jest.fn();
    this.deleteProduct = jest.fn();
    this.health = jest.fn();
  }
}

module.exports = ProductStoreMock;
