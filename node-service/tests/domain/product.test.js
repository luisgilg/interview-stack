const test = require('node:test');
const assert = require('node:assert');

const Product = require('../../internal/domain/product');
const DomainError = require('../../internal/domain/errors');

test('product validation succeeds', () => {
  const product = new Product({ name: 'Keyboard', price: 100 });
  assert.doesNotThrow(() => product.validate());
  assert.deepStrictEqual(product.tags, []);
});

test('product validation fails for invalid input', () => {
  const product = new Product({ name: '', price: -10 });
  assert.throws(() => product.validate(), (err) => {
    assert.ok(err instanceof DomainError);
    assert.strictEqual(err.code, 'validation');
    return true;
  });
});
