const { v4: uuid } = require('uuid');
const Product = require('../../domain/product');
const ProductModel = require('./productModel');

function toDomain(doc) {
  return new Product({
    id: doc._id,
    name: doc.name,
    price: Number(doc.price),
    tags: Array.isArray(doc.tags) ? [...doc.tags] : [],
    createdAt: doc.created_at,
    updatedAt: doc.updated_at
  });
}

class MongoProductStore {
  constructor(operationTimeout) {
    this.operationTimeout = operationTimeout;
  }

  async listProducts() {
    const docs = await ProductModel.find()
      .sort({ created_at: -1 })
      .maxTimeMS(this.operationTimeout)
      .lean()
      .exec();
    return docs.map(toDomain);
  }

  async getProduct(id) {
    const doc = await ProductModel.findById(id).maxTimeMS(this.operationTimeout).lean().exec();
    return doc ? toDomain(doc) : null;
  }

  async createProduct(product) {
    const payload = {
      _id: product.id || uuid(),
      name: product.name,
      price: product.price,
      tags: product.tags || [],
      created_at: product.createdAt,
      updated_at: product.updatedAt
    };
    const doc = new ProductModel(payload);
    await doc.save({ maxTimeMS: this.operationTimeout });
    return toDomain(doc.toObject());
  }

  async updateProduct(id, product) {
    const doc = await ProductModel.findByIdAndUpdate(
      id,
      {
        name: product.name,
        price: product.price,
        tags: product.tags || [],
        updated_at: product.updatedAt
      },
      { new: true, maxTimeMS: this.operationTimeout }
    )
      .lean()
      .exec();
    return doc ? toDomain(doc) : null;
  }

  async deleteProduct(id) {
    const deleted = await ProductModel.findByIdAndDelete(id, { maxTimeMS: this.operationTimeout }).exec();
    return Boolean(deleted);
  }

  async ensureIndexes() {
    try {
      await ProductModel.collection.createIndex(
        { name: 1 },
        { name: 'idx_products_name', maxTimeMS: this.operationTimeout }
      );
    } catch (err) {
      if (err?.code === 85 || err?.codeName === 'IndexOptionsConflict') {
        return;
      }
      throw err;
    }
  }

  async health() {
    if (ProductModel.db.readyState !== 1) {
      throw new Error('mongo not connected');
    }
    await ProductModel.db.db.command({ ping: 1, maxTimeMS: this.operationTimeout });
  }
}

module.exports = MongoProductStore;
