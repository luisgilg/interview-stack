package repository

import (
	"context"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// ProductRepository bridges the domain port with the configured store implementation.
type ProductRepository struct {
	store ProductStore
	clock domain.Clock
}

func NewProductRepository(store ProductStore, clock domain.Clock) *ProductRepository {
	return &ProductRepository{
		store: store,
		clock: clock,
	}
}

func (r *ProductRepository) ListProducts(ctx context.Context) ([]domain.Product, error) {
	products, err := r.store.ListProducts(ctx)
	if err != nil {
		return nil, domain.NewInternalError("failed to list products", err)
	}
	return products, nil
}

func (r *ProductRepository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := r.store.GetProduct(ctx, id)
	if err != nil {
		return nil, domain.NewInternalError("failed to fetch product", err)
	}
	return product, nil
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error) {
	product = r.ensureTimestamps(product, true)
	created, err := r.store.CreateProduct(ctx, product)
	if err != nil {
		return nil, domain.NewInternalError("failed to save product", err)
	}
	return created, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, id string, product domain.Product) (*domain.Product, error) {
	product = r.ensureTimestamps(product, false)
	updated, err := r.store.UpdateProduct(ctx, id, product)
	if err != nil {
		return nil, domain.NewInternalError("failed to update product", err)
	}
	return updated, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
	deleted, err := r.store.DeleteProduct(ctx, id)
	if err != nil {
		return domain.NewInternalError("failed to delete product", err)
	}
	if !deleted {
		return domain.NewNotFoundError("product not found")
	}
	return nil
}

func (r *ProductRepository) Health(ctx context.Context) error {
	if err := r.store.Health(ctx); err != nil {
		return domain.NewInternalError("store unhealthy", err)
	}
	return nil
}

func (r *ProductRepository) ensureTimestamps(product domain.Product, isCreate bool) domain.Product {
	if r.clock == nil {
		return product
	}
	if product.CreatedAt.IsZero() && isCreate {
		product.CreatedAt = r.clock.Now()
	}
	if product.UpdatedAt.IsZero() {
		if isCreate {
			product.UpdatedAt = product.CreatedAt
		} else {
			product.UpdatedAt = r.clock.Now()
		}
	}
	return product
}
