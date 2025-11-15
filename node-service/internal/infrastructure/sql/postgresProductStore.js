const { v4: uuid } = require('uuid');
const Product = require('../../domain/product');

function mapRow(row) {
  return new Product({
    id: row.id,
    name: row.name,
    price: Number(row.price),
    tags: row.tags || [],
    createdAt: row.created_at,
    updatedAt: row.updated_at
  });
}

class PostgresProductStore {
  constructor(pool) {
    this.pool = pool;
  }

  async listProducts() {
    const query = 'SELECT id, name, price, tags, created_at, updated_at FROM products ORDER BY created_at DESC';
    const { rows } = await this.pool.query(query);
    return rows.map(mapRow);
  }

  async getProduct(id) {
    const query = 'SELECT id, name, price, tags, created_at, updated_at FROM products WHERE id = $1';
    const { rows } = await this.pool.query(query, [id]);
    if (!rows.length) {
      return null;
    }
    return mapRow(rows[0]);
  }

  async createProduct(product) {
    const id = product.id || uuid();
    const query = `INSERT INTO products (id, name, price, tags, created_at, updated_at)
                   VALUES ($1, $2, $3, $4, $5, $6)
                   RETURNING id, name, price, tags, created_at, updated_at`;
    const values = [id, product.name, product.price, product.tags || [], product.createdAt, product.updatedAt];
    const { rows } = await this.pool.query(query, values);
    return mapRow(rows[0]);
  }

  async updateProduct(id, product) {
    const query = `UPDATE products
                   SET name = $1, price = $2, tags = $3, updated_at = $4
                   WHERE id = $5
                   RETURNING id, name, price, tags, created_at, updated_at`;
    const { rows } = await this.pool.query(query, [
      product.name,
      product.price,
      product.tags || [],
      product.updatedAt,
      id
    ]);
    if (!rows.length) {
      return null;
    }
    return mapRow(rows[0]);
  }

  async deleteProduct(id) {
    const query = 'DELETE FROM products WHERE id = $1';
    const { rowCount } = await this.pool.query(query, [id]);
    return rowCount > 0;
  }

  async health() {
    await this.pool.query('SELECT 1');
  }
}

module.exports = PostgresProductStore;
