const test = require('node:test');
const assert = require('node:assert');
const express = require('express');

const ProductController = require('../../internal/interface/http/productController');
const DomainError = require('../../internal/domain/errors');

function createApp(controller) {
  const app = express();
  app.use(express.json());
  controller.register(app);
  return app;
}

function buildController(overrides = {}) {
  return new ProductController(
    {
      list: { execute: async () => [] },
      get: { execute: async () => ({}) },
      create: { execute: async () => ({}) },
      update: { execute: async () => ({}) },
      remove: { execute: async () => {} },
      health: { execute: async () => {} },
      ...overrides
    },
    { read: 50, write: 50, health: 50 }
  );
}

test('GET /products returns products', async () => {
  const controller = buildController({
    list: {
      execute: async () => [
        {
          id: '1',
          name: 'Keyboard',
          price: 10,
          tags: [],
          created_at: new Date('2023-01-01T00:00:00.000Z'),
          updated_at: new Date('2023-01-01T00:00:00.000Z')
        }
      ]
    }
  });

  const app = createApp(controller);
  const server = app.listen(0);
  const port = server.address().port;

  const response = await fetch(`http://127.0.0.1:${port}/products`);
  const body = await response.json();
  assert.strictEqual(response.status, 200);
  assert.strictEqual(body.length, 1);
  server.close();
});

test('POST /products maps validation errors', async () => {
  const controller = buildController({
    create: { execute: async () => { throw DomainError.validation('invalid'); } }
  });

  const app = createApp(controller);
  const server = app.listen(0);
  const port = server.address().port;

  const response = await fetch(`http://127.0.0.1:${port}/products`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: '', price: 0 })
  });
  assert.strictEqual(response.status, 400);
  const payload = await response.json();
  assert.strictEqual(payload.error, 'invalid');
  server.close();
});
