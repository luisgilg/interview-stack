package usecase

import (
	"context"
	"fmt"

	"github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

// GetProductUseCase retrieves a single product.
type GetProductUseCase struct {
	repo   domain.ProductRepository
	logger domain.Logger
	cache  *cache.Service
}

func NewGetProductUseCase(repo domain.ProductRepository, logger domain.Logger, cacheSvc *cache.Service) *GetProductUseCase {
	return &GetProductUseCase{repo: repo, logger: logger, cache: cacheSvc}
}

func (uc *GetProductUseCase) Execute(ctx context.Context, id string) (*dto.ProductResponse, error) {
	loader := func(ctx context.Context) (*dto.ProductResponse, error) {
		product, err := uc.repo.GetProduct(ctx, id)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, domain.NewNotFoundError("product not found")
		}
		dtoProduct := dto.FromDomainProduct(product)
		return &dtoProduct, nil
	}

	cacheKey := fmt.Sprintf("products:%s", id)
	result, meta, err := cache.Fetch(uc.cache, ctx, cacheKey, loader)
	if err != nil {
		return nil, err
	}
	if result != nil {
		result.CacheStatus = string(meta.Status)
	}
	uc.logger.Info("retrieved product", domain.KV("id", id), domain.KV("cache_status", meta.Status))
	return result, nil
}
