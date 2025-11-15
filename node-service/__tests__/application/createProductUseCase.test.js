const CreateProductUseCase = require('../../internal/application/usecases/createProduct');
const FakeClock = require('../../internal/core/clock/FakeClock');
const DomainError = require('../../internal/domain/errors');

function createLogger() {
  return {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn()
  };
}

describe('CreateProductUseCase', () => {
  it('calls repository with timestamps sourced from the clock', async () => {
    const repository = { create: jest.fn().mockResolvedValue({ id: 'abc' }) };
    const clock = new FakeClock(new Date('2024-01-02T03:04:05.000Z'));
    const useCase = new CreateProductUseCase(repository, createLogger(), clock, null, null, false, 'node-service');

    const result = await useCase.execute({ name: 'Keyboard', price: 10 });

    expect(repository.create).toHaveBeenCalledTimes(1);
    const savedProduct = repository.create.mock.calls[0][0];
    expect(savedProduct.createdAt).toEqual(new Date('2024-01-02T03:04:05.000Z'));
    expect(savedProduct.updatedAt).toEqual(new Date('2024-01-02T03:04:05.000Z'));
    expect(result).toEqual({ id: 'abc' });
  });

  it('rejects invalid input without calling the repository', async () => {
    const repository = { create: jest.fn() };
    const clock = new FakeClock(new Date('2024-01-02T03:04:05.000Z'));
    const useCase = new CreateProductUseCase(repository, createLogger(), clock, null, null, false, 'node-service');

    await expect(useCase.execute({ name: '', price: 0 })).rejects.toBeInstanceOf(DomainError);
    expect(repository.create).not.toHaveBeenCalled();
  });

  it('enqueues write-behind event when enabled', async () => {
    const repository = { create: jest.fn() };
    const clock = new FakeClock(new Date('2024-03-01T00:00:00.000Z'));
    const cache = { upsertProduct: jest.fn() };
    const queue = { enqueue: jest.fn().mockResolvedValue() };
    const useCase = new CreateProductUseCase(repository, createLogger(), clock, cache, queue, true, 'node-service');

    const result = await useCase.execute({ name: 'Buffered', price: 50 });

    expect(repository.create).not.toHaveBeenCalled();
    expect(cache.upsertProduct).toHaveBeenCalled();
    expect(queue.enqueue).toHaveBeenCalledTimes(1);
    const event = queue.enqueue.mock.calls[0][0];
    expect(event.type).toBe('create');
    expect(event.source).toBe('node-service');
    expect(result.id).toBeDefined();
  });
});
