const Product = require('../../domain/product');

function cloneTags(tags) {
  if (!Array.isArray(tags)) {
    return [];
  }
  return [...tags];
}

function toDomain(input) {
  return new Product({
    id: input.id,
    name: input.name,
    price: input.price,
    tags: cloneTags(input.tags),
    createdAt: input.createdAt,
    updatedAt: input.updatedAt
  });
}

function toResponse(product) {
  return {
    id: product.id,
    name: product.name,
    price: product.price,
    tags: cloneTags(product.tags),
    created_at: product.createdAt,
    updated_at: product.updatedAt
  };
}

function toResponseList(products) {
  return products.map(toResponse);
}

module.exports = {
  toDomain,
  toResponse,
  toResponseList
};
