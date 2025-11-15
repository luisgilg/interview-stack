const ProductRepository = require('../../internal/infrastructure/repository/productRepository');
const ProductStoreMock = require('../../__mocks__/ProductStore');
const ClockMock = require('../../__mocks__/Clock');

describe('ProductRepository', () => {
  it('creates products with injected timestamps', async () => {
    const now = new Date('2024-01-01T00:00:00.000Z');
    const clock = new ClockMock(now);
    const store = new ProductStoreMock();
    store.createProduct.mockResolvedValue({ id: '1', name: 'Keyboard', price: 10, createdAt: now, updatedAt: now });
    const repository = new ProductRepository(store, clock);

    await repository.create({ name: 'Keyboard', price: 10 });

    expect(store.createProduct).toHaveBeenCalledTimes(1);
    const savedProduct = store.createProduct.mock.calls[0][0];
    expect(savedProduct.createdAt).toEqual(now);
    expect(savedProduct.updatedAt).toEqual(now);
  });

  it('uses the clock when updating', async () => {
    const now = new Date('2024-05-05T10:11:12.000Z');
    const clock = new ClockMock(now);
    const store = new ProductStoreMock();
    store.updateProduct.mockResolvedValue({ id: '1', name: 'Keyboard', price: 25, updatedAt: now });
    const repository = new ProductRepository(store, clock);

    await repository.update('1', { name: 'Keyboard', price: 25 });

    expect(store.updateProduct).toHaveBeenCalledTimes(1);
    const updatedProduct = store.updateProduct.mock.calls[0][1];
    expect(updatedProduct.updatedAt).toEqual(now);
  });
});
