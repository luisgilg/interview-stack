package usecase

import (
	"context"
	"fmt"

	"github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

type cacheWriter struct {
	cache  *cache.Service
	logger domain.Logger
}

func newCacheWriter(cacheSvc *cache.Service, logger domain.Logger) cacheWriter {
	return cacheWriter{
		cache:  cacheSvc,
		logger: logger,
	}
}

func (w cacheWriter) upsertProduct(ctx context.Context, product *dto.ProductResponse) {
	if w.cache == nil || product == nil || product.ID == "" {
		return
	}
	key := fmt.Sprintf("products:%s", product.ID)
	if err := w.cache.Store(ctx, key, product); err != nil {
		w.logger.Warn("failed to update product cache", domain.KV("key", key), domain.KV("error", err.Error()))
	}
}

func (w cacheWriter) deleteProduct(ctx context.Context, id string) {
	if w.cache == nil || id == "" {
		return
	}
	key := fmt.Sprintf("products:%s", id)
	if err := w.cache.Delete(ctx, key); err != nil {
		w.logger.Warn("failed to delete product cache", domain.KV("key", key), domain.KV("error", err.Error()))
	}
}

func (w cacheWriter) invalidateList(ctx context.Context) {
	if w.cache == nil {
		return
	}
	if err := w.cache.Delete(ctx, "products:list"); err != nil {
		w.logger.Warn("failed to invalidate product list cache", domain.KV("error", err.Error()))
	}
}
