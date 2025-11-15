const express = require('express');
const request = require('supertest');

const DomainError = require('../../internal/domain/errors');
const ProductController = require('../../internal/interface/http/productController');

function buildApp(overrides = {}) {
  const useCases = {
    list: { execute: jest.fn().mockResolvedValue([]) },
    get: { execute: jest.fn().mockResolvedValue(null) },
    create: { execute: jest.fn() },
    update: { execute: jest.fn().mockResolvedValue(null) },
    remove: { execute: jest.fn().mockResolvedValue(null) },
    health: { execute: jest.fn().mockResolvedValue({ status: 'ok' }) },
    ...overrides
  };

  const controller = new ProductController(useCases, { read: 50, write: 50, health: 50 });
  const app = express();
  app.use(express.json());
  controller.register(app);
  return { app, useCases };
}

describe('ProductController', () => {
  it('creates a product and returns 201', async () => {
    const payload = { name: 'Keyboard', price: 10 };
    const created = { id: '123', ...payload };
    const { app, useCases } = buildApp();
    useCases.create.execute.mockResolvedValue(created);

    const res = await request(app).post('/products').send(payload).expect(201);

    expect(useCases.create.execute).toHaveBeenCalledWith(payload);
    expect(res.body).toEqual(created);
  });

  it('maps validation errors to 400 responses', async () => {
    const payload = { name: '', price: 0 };
    const { app, useCases } = buildApp();
    useCases.create.execute.mockRejectedValue(DomainError.validation('missing name'));

    const res = await request(app).post('/products').send(payload).expect(400);

    expect(res.body).toEqual({ error: 'missing name' });
  });
});
