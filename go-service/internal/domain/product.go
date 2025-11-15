package domain

import (
	"strings"
	"time"
)

// Product represents the aggregate stored in both persistence technologies.
type Product struct {
	ID        string
	Name      string
	Price     float64
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate ensures domain invariants for a product.
func (p *Product) Validate() *Error {
	if strings.TrimSpace(p.Name) == "" {
		return NewValidationError("name is required", nil)
	}
	if p.Price <= 0 {
		return NewValidationError("price must be positive", nil)
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	return nil
}
