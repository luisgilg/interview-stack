package dto

import (
	"time"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// CreateProductRequest represents inbound payload for create operations.
type CreateProductRequest struct {
	Name  string   `json:"name" example:"Premium Widget"`
	Price float64  `json:"price" example:"29.99"`
	Tags  []string `json:"tags" example:"[\"gadget\",\"home\"]"`
}

// UpdateProductRequest represents inbound payload for update operations.
type UpdateProductRequest struct {
	Name  string   `json:"name" example:"Premium Widget"`
	Price float64  `json:"price" example:"29.99"`
	Tags  []string `json:"tags" example:"[\"gadget\",\"home\"]"`
}

// ProductResponse is the outbound DTO exposed through HTTP.
type ProductResponse struct {
	ID        string    `json:"id" example:"a1b2c3"`
	Name      string    `json:"name" example:"Premium Widget"`
	Price     float64   `json:"price" example:"29.99"`
	Tags      []string  `json:"tags" example:"[\"gadget\",\"home\"]"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-02T15:04:05Z"`
	CacheStatus string  `json:"cache_status,omitempty" example:"fresh"`
}

// ToDomain converts the inbound DTO into a domain entity.
func (c CreateProductRequest) ToDomain() domain.Product {
	return domain.Product{
		Name:  c.Name,
		Price: c.Price,
		Tags:  cloneTags(c.Tags),
	}
}

// ToDomain converts an update request into a domain entity.
func (u UpdateProductRequest) ToDomain() domain.Product {
	return domain.Product{
		Name:  u.Name,
		Price: u.Price,
		Tags:  cloneTags(u.Tags),
	}
}

// FromDomainProduct maps a domain product into a response DTO.
func FromDomainProduct(p *domain.Product) ProductResponse {
	tags := cloneTags(p.Tags)
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Price:       p.Price,
		Tags:        tags,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		CacheStatus: "",
	}
}

// FromDomainProducts maps a slice of domain entities to DTOs.
func FromDomainProducts(products []domain.Product) []ProductResponse {
	result := make([]ProductResponse, 0, len(products))
	for i := range products {
		product := products[i] // avoid pointer aliasing
		result = append(result, FromDomainProduct(&product))
	}
	return result
}

func cloneTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	out := make([]string, len(tags))
	copy(out, tags)
	return out
}
