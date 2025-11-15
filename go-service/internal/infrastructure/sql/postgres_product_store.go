package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// ProductStore encapsulates Postgres operations for products.
type ProductStore struct {
	pool *pgxpool.Pool
}

func NewProductStore(pool *pgxpool.Pool) *ProductStore {
	return &ProductStore{pool: pool}
}

func (s *ProductStore) ListProducts(ctx context.Context) ([]domain.Product, error) {
	const query = `SELECT id, name, price, tags, created_at, updated_at FROM products ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Tags, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (s *ProductStore) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	const query = `SELECT id, name, price, tags, created_at, updated_at FROM products WHERE id = $1`
	var product domain.Product
	if err := s.pool.QueryRow(ctx, query, id).Scan(&product.ID, &product.Name, &product.Price, &product.Tags, &product.CreatedAt, &product.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error) {
	if product.ID == "" {
		product.ID = uuid.NewString()
	}
	if product.Tags == nil {
		product.Tags = []string{}
	}

	const query = `INSERT INTO products (id, name, price, tags, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := s.pool.Exec(ctx, query, product.ID, product.Name, product.Price, product.Tags, product.CreatedAt, product.UpdatedAt); err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) UpdateProduct(ctx context.Context, id string, product domain.Product) (*domain.Product, error) {
	if product.Tags == nil {
		product.Tags = []string{}
	}

	const query = `UPDATE products SET name = $1, price = $2, tags = $3, updated_at = $4 WHERE id = $5 RETURNING id, name, price, tags, created_at, updated_at`
	var updated domain.Product
	err := s.pool.QueryRow(ctx, query, product.Name, product.Price, product.Tags, product.UpdatedAt, id).Scan(
		&updated.ID, &updated.Name, &updated.Price, &updated.Tags, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *ProductStore) DeleteProduct(ctx context.Context, id string) (bool, error) {
	const query = `DELETE FROM products WHERE id = $1`
	cmd, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (s *ProductStore) Health(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
