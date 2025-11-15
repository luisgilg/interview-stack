const test = require('node:test');
const assert = require('node:assert');

const CreateProductUseCase = require('../../internal/application/usecases/createProduct');
const DomainError = require('../../internal/domain/errors');
const FakeClock = require('../../internal/core/clock/FakeClock');

class FakeRepo {
  constructor() {
    this.createdProduct = null;
  }

  async create(product) {
    this.createdProduct = { ...product, id: '123' };
    return this.createdProduct;
  }
}

class NoopLogger {
  info() {}
  warn() {}
  error() {}
}

test('CreateProductUseCase validates and creates product', async () => {
  const repo = new FakeRepo();
  const fixedDate = new Date('2024-01-02T03:04:05.000Z');
  const clock = new FakeClock(fixedDate);
  const useCase = new CreateProductUseCase(repo, new NoopLogger(), clock, null, null, false, 'node-service');

  const result = await useCase.execute({ name: 'Keyboard', price: 10, tags: [] });
  assert.strictEqual(result.id, '123');
  assert.deepStrictEqual(repo.createdProduct.createdAt, fixedDate);
  assert.deepStrictEqual(repo.createdProduct.updatedAt, fixedDate);
});

test('CreateProductUseCase rejects invalid payload', async () => {
  const repo = new FakeRepo();
  const clock = new FakeClock(new Date('2023-01-01T00:00:00.000Z'));
  const useCase = new CreateProductUseCase(repo, new NoopLogger(), clock, null, null, false, 'node-service');

  await assert.rejects(() => useCase.execute({ name: '', price: 0 }), (err) => {
    assert.ok(err instanceof DomainError);
    assert.strictEqual(err.code, 'validation');
    return true;
  });
});
