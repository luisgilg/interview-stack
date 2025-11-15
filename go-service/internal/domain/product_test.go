package domain

import "testing"

func TestProductValidate(t *testing.T) {
	t.Run("valid product", func(t *testing.T) {
		p := &Product{Name: "Keyboard", Price: 10}
		if err := p.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(p.Tags) != 0 {
			t.Fatalf("expected tags to be initialized")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		p := &Product{Price: 10}
		err := p.Validate()
		if err == nil || err.Code != ErrorCodeValidation {
			t.Fatalf("expected validation error for missing name")
		}
	})

	t.Run("invalid price", func(t *testing.T) {
		p := &Product{Name: "Keyboard", Price: 0}
		err := p.Validate()
		if err == nil || err.Code != ErrorCodeValidation {
			t.Fatalf("expected validation error for invalid price")
		}
	})
}
