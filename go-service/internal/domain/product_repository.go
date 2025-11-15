package domain

import "context"

// ProductRepository defines the required operations from persistence adapters.
type ProductRepository interface {
	ListProducts(ctx context.Context) ([]Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	CreateProduct(ctx context.Context, product Product) (*Product, error)
	UpdateProduct(ctx context.Context, id string, product Product) (*Product, error)
	DeleteProduct(ctx context.Context, id string) error
	Health(ctx context.Context) error
}
