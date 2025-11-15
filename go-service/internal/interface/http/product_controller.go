package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/example/interview-stack/go-service/internal/application/dto"
)

// RequestTimeouts configures how long handlers wait for downstream use cases.
type RequestTimeouts struct {
	Read   time.Duration
	Write  time.Duration
	Health time.Duration
}

type listProductsUseCase interface {
	Execute(ctx context.Context) ([]dto.ProductResponse, error)
}

type getProductUseCase interface {
	Execute(ctx context.Context, id string) (*dto.ProductResponse, error)
}

type createProductUseCase interface {
	Execute(ctx context.Context, input dto.CreateProductRequest) (*dto.ProductResponse, error)
}

type updateProductUseCase interface {
	Execute(ctx context.Context, id string, input dto.UpdateProductRequest) (*dto.ProductResponse, error)
}

type deleteProductUseCase interface {
	Execute(ctx context.Context, id string) error
}

type healthCheckUseCase interface {
	Execute(ctx context.Context) error
}

// ProductController translates HTTP payloads to use case invocations.
type ProductController struct {
	listUC   listProductsUseCase
	getUC    getProductUseCase
	createUC createProductUseCase
	updateUC updateProductUseCase
	deleteUC deleteProductUseCase
	healthUC healthCheckUseCase
	timeouts RequestTimeouts
}

func NewProductController(
	listUC listProductsUseCase,
	getUC getProductUseCase,
	createUC createProductUseCase,
	updateUC updateProductUseCase,
	deleteUC deleteProductUseCase,
	healthUC healthCheckUseCase,
	timeouts RequestTimeouts,
) *ProductController {
	return &ProductController{
		listUC:   listUC,
		getUC:    getUC,
		createUC: createUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
		healthUC: healthUC,
		timeouts: timeouts,
	}
}

// ListProducts godoc
// @Summary List products
// @Description Returns all stored products
// @Tags Products
// @Produce json
// @Success 200 {array} dto.ProductResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products [get]
func (c *ProductController) ListProducts(ctx *fiber.Ctx) error {
	reqCtx, cancel := context.WithTimeout(ctx.Context(), c.timeouts.Read)
	defer cancel()

	products, err := c.listUC.Execute(reqCtx)
	if err != nil {
		return handleError(ctx, err)
	}
	return ctx.JSON(products)
}

// GetProduct godoc
// @Summary Get product
// @Description Fetch a product by its ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{id} [get]
func (c *ProductController) GetProduct(ctx *fiber.Ctx) error {
	reqCtx, cancel := context.WithTimeout(ctx.Context(), c.timeouts.Read)
	defer cancel()

	product, err := c.getUC.Execute(reqCtx, ctx.Params("id"))
	if err != nil {
		return handleError(ctx, err)
	}
	return ctx.JSON(product)
}

// CreateProduct godoc
// @Summary Create product
// @Description Creates a new product from the provided payload
// @Tags Products
// @Accept json
// @Produce json
// @Param product body dto.CreateProductRequest true "Product payload"
// @Success 201 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products [post]
func (c *ProductController) CreateProduct(ctx *fiber.Ctx) error {
	var input dto.CreateProductRequest
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	reqCtx, cancel := context.WithTimeout(ctx.Context(), c.timeouts.Write)
	defer cancel()

	product, err := c.createUC.Execute(reqCtx, input)
	if err != nil {
		return handleError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(product)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Updates an existing product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body dto.UpdateProductRequest true "Updated product payload"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{id} [put]
func (c *ProductController) UpdateProduct(ctx *fiber.Ctx) error {
	var input dto.UpdateProductRequest
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	reqCtx, cancel := context.WithTimeout(ctx.Context(), c.timeouts.Write)
	defer cancel()

	product, err := c.updateUC.Execute(reqCtx, ctx.Params("id"), input)
	if err != nil {
		return handleError(ctx, err)
	}
	return ctx.JSON(product)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Removes a product permanently
// @Tags Products
// @Param id path string true "Product ID"
// @Success 204
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{id} [delete]
func (c *ProductController) DeleteProduct(ctx *fiber.Ctx) error {
	reqCtx, cancel := context.WithTimeout(ctx.Context(), c.timeouts.Write)
	defer cancel()

	if err := c.deleteUC.Execute(reqCtx, ctx.Params("id")); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// Health godoc
// @Summary Health check
// @Description Returns the health status of the service
// @Tags Health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /health [get]
func (c *ProductController) Health(ctx *fiber.Ctx) error {
	reqCtx, cancel := context.WithTimeout(ctx.Context(), c.timeouts.Health)
	defer cancel()

	if err := c.healthUC.Execute(reqCtx); err != nil {
		return handleError(ctx, err)
	}
	return ctx.JSON(dto.HealthResponse{Status: "ok"})
}
