package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

type productUseCaseStub struct {
	listResult []dto.ProductResponse
	listErr    error

	getResult *dto.ProductResponse
	getErr    error

	createResult *dto.ProductResponse
	createErr    error

	updateResult *dto.ProductResponse
	updateErr    error

	deleteErr error
	healthErr error
}

func (s *productUseCaseStub) Execute(ctx context.Context) ([]dto.ProductResponse, error) {
	return s.listResult, s.listErr
}

func (s *productUseCaseStub) ExecuteGet(ctx context.Context, id string) (*dto.ProductResponse, error) {
	return s.getResult, s.getErr
}

func (s *productUseCaseStub) ExecuteCreate(ctx context.Context, input dto.CreateProductRequest) (*dto.ProductResponse, error) {
	return s.createResult, s.createErr
}

func (s *productUseCaseStub) ExecuteUpdate(ctx context.Context, id string, input dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	return s.updateResult, s.updateErr
}

func (s *productUseCaseStub) ExecuteDelete(ctx context.Context, id string) error {
	return s.deleteErr
}

func (s *productUseCaseStub) ExecuteHealth(ctx context.Context) error {
	return s.healthErr
}

// Adapter methods to satisfy interfaces.
func (s *productUseCaseStub) ExecuteList(ctx context.Context) ([]dto.ProductResponse, error) {
	return s.Execute(ctx)
}

// Interface assertions.
var (
	_ listProductsUseCase  = (*listAdapter)(nil)
	_ getProductUseCase    = (*getAdapter)(nil)
	_ createProductUseCase = (*createAdapter)(nil)
	_ updateProductUseCase = (*updateAdapter)(nil)
	_ deleteProductUseCase = (*deleteAdapter)(nil)
	_ healthCheckUseCase   = (*healthAdapter)(nil)
)

type listAdapter struct{ stub *productUseCaseStub }

func (a listAdapter) Execute(ctx context.Context) ([]dto.ProductResponse, error) {
	return a.stub.Execute(ctx)
}

type getAdapter struct{ stub *productUseCaseStub }

func (a getAdapter) Execute(ctx context.Context, id string) (*dto.ProductResponse, error) {
	return a.stub.ExecuteGet(ctx, id)
}

type createAdapter struct{ stub *productUseCaseStub }

func (a createAdapter) Execute(ctx context.Context, input dto.CreateProductRequest) (*dto.ProductResponse, error) {
	return a.stub.ExecuteCreate(ctx, input)
}

type updateAdapter struct{ stub *productUseCaseStub }

func (a updateAdapter) Execute(ctx context.Context, id string, input dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	return a.stub.ExecuteUpdate(ctx, id, input)
}

type deleteAdapter struct{ stub *productUseCaseStub }

func (a deleteAdapter) Execute(ctx context.Context, id string) error {
	return a.stub.ExecuteDelete(ctx, id)
}

type healthAdapter struct{ stub *productUseCaseStub }

func (a healthAdapter) Execute(ctx context.Context) error {
	return a.stub.ExecuteHealth(ctx)
}

func newControllerWithStub(stub *productUseCaseStub) *ProductController {
	return NewProductController(
		listAdapter{stub},
		getAdapter{stub},
		createAdapter{stub},
		updateAdapter{stub},
		deleteAdapter{stub},
		healthAdapter{stub},
		RequestTimeouts{
			Read:   100 * time.Millisecond,
			Write:  100 * time.Millisecond,
			Health: 100 * time.Millisecond,
		},
	)
}

func TestProductControllerListProducts(t *testing.T) {
	stub := &productUseCaseStub{
		listResult: []dto.ProductResponse{{ID: "1", Name: "Keyboard", Price: 10}},
	}
	controller := newControllerWithStub(stub)

	app := fiber.New()
	app.Get("/products", controller.ListProducts)

	req := httptest.NewRequest("GET", "/products", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200 got %d", resp.StatusCode)
	}
	var payload []dto.ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload) != 1 || payload[0].ID != "1" {
		t.Fatalf("unexpected response: %+v", payload)
	}
}

func TestProductControllerCreateProductValidationError(t *testing.T) {
	stub := &productUseCaseStub{
		createErr: domain.NewValidationError("name is required", nil),
	}
	controller := newControllerWithStub(stub)

	app := fiber.New()
	app.Post("/products", controller.CreateProduct)

	body, _ := json.Marshal(dto.CreateProductRequest{})
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400 got %d", resp.StatusCode)
	}
}
