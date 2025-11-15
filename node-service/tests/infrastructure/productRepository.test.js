const test = require('node:test');
const assert = require('node:assert');

const ProductRepository = require('../../internal/infrastructure/repository/productRepository');
const DomainError = require('../../internal/domain/errors');
const Product = require('../../internal/domain/product');
const FakeClock = require('../../internal/core/clock/FakeClock');

class StoreStub {
  constructor(overrides = {}) {
    this.listProducts = overrides.listProducts || (async () => []);
    this.getProduct = overrides.getProduct || (async () => null);
    this.createProduct =
      overrides.createProduct ||
      (async (product) => ({ ...product, id: 'abc' }));
    this.updateProduct = overrides.updateProduct || (async () => null);
    this.deleteProduct = overrides.deleteProduct || (async () => true);
    this.health = overrides.health || (async () => {});
  }
}

test('ProductRepository delegates create to the configured store', async () => {
  const clock = new FakeClock(new Date('2024-02-03T04:05:06.000Z'));
  let recordedProduct;
  const repo = new ProductRepository(
    new StoreStub({
      createProduct: async (product) => {
        recordedProduct = product;
        return { ...product, id: 'abc' };
      }
    }),
    clock
  );
  const product = new Product({ name: 'Keyboard', price: 10 });
  const created = await repo.create(product);
  assert.strictEqual(created.id, 'abc');
  assert.deepStrictEqual(recordedProduct.createdAt, clock.now());
  assert.deepStrictEqual(recordedProduct.updatedAt, clock.now());
});

test('ProductRepository wraps store errors as DomainError', async () => {
  const repo = new ProductRepository(
    new StoreStub({
      createProduct: async () => {
        throw new Error('boom');
      }
    }),
    new FakeClock(new Date('2023-01-01T00:00:00.000Z'))
  );

  await assert.rejects(() => repo.create(new Product({ name: 'Mouse', price: 20 })), (err) => {
    assert.ok(err instanceof DomainError);
    assert.strictEqual(err.code, 'internal');
    assert.strictEqual(err.message, 'failed to create product');
    return true;
  });
});
