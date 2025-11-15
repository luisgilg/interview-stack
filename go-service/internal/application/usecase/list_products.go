package usecase

import (
	"context"

	"github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

// ListProductsUseCase handles retrieval of all products.
type ListProductsUseCase struct {
	repo   domain.ProductRepository
	logger domain.Logger
	cache  *cache.Service
}

func NewListProductsUseCase(repo domain.ProductRepository, logger domain.Logger, cacheSvc *cache.Service) *ListProductsUseCase {
	return &ListProductsUseCase{repo: repo, logger: logger, cache: cacheSvc}
}

func (uc *ListProductsUseCase) Execute(ctx context.Context) ([]dto.ProductResponse, error) {
	loader := func(ctx context.Context) ([]dto.ProductResponse, error) {
		products, err := uc.repo.ListProducts(ctx)
		if err != nil {
			return nil, err
		}
		return dto.FromDomainProducts(products), nil
	}

	result, meta, err := cache.Fetch(uc.cache, ctx, "products:list", loader)
	if err != nil {
		return nil, err
	}
	annotateCacheStatus(result, meta.Status)
	uc.logger.Info("list products executed", domain.KV("count", len(result)), domain.KV("cache_status", meta.Status))
	return result, nil
}

func annotateCacheStatus(products []dto.ProductResponse, status cache.Status) {
	if status == "" {
		return
	}
	for i := range products {
		products[i].CacheStatus = string(status)
	}
}
