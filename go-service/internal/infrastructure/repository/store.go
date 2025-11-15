package repository

//go:generate mockery --name=ProductStore --output=../../mocks --outpkg=mocks --filename=product_store_mock.go --structname=ProductStoreMock --with-expecter

import (
	"context"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// ProductStore abstracts the persistence operations for products.
type ProductStore interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error)
	UpdateProduct(ctx context.Context, id string, product domain.Product) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id string) (bool, error)
	Health(ctx context.Context) error
}
