package usecase

import (
	"context"

	appcache "github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/domain"
)

// DeleteProductUseCase deletes a product.
type DeleteProductUseCase struct {
	repo         domain.ProductRepository
	logger       domain.Logger
	queue        domain.WriteQueue
	writeBehind  bool
	source       string
	clock        domain.Clock
	cacheWriter  cacheWriter
	cacheEnabled bool
}

func NewDeleteProductUseCase(
	repo domain.ProductRepository,
	logger domain.Logger,
	clock domain.Clock,
	cacheSvc *appcache.Service,
	queue domain.WriteQueue,
	writeBehind bool,
	source string,
) *DeleteProductUseCase {
	return &DeleteProductUseCase{
		repo:         repo,
		logger:       logger,
		queue:        queue,
		writeBehind:  writeBehind,
		source:       source,
		clock:        clock,
		cacheWriter:  newCacheWriter(cacheSvc, logger),
		cacheEnabled: cacheSvc != nil && cacheSvc.Enabled(),
	}
}

func (uc *DeleteProductUseCase) Execute(ctx context.Context, id string) error {
	if uc.writeBehind && uc.queue != nil {
		product, err := uc.repo.GetProduct(ctx, id)
		if err != nil {
			return err
		}
		if product == nil {
			return domain.NewNotFoundError("product not found")
		}
		uc.cacheWriter.deleteProduct(ctx, id)
		uc.cacheWriter.invalidateList(ctx)
		event := domain.WriteEvent{
			Type:      domain.WriteEventDelete,
			ID:        id,
			Timestamp: uc.clock.Now(),
			Source:    uc.source,
		}
		if err := uc.queue.Enqueue(ctx, event); err != nil {
			return domain.NewInternalError("failed to enqueue delete event", err)
		}
		uc.logger.Info("product enqueued for deletion", domain.KV("id", id))
		return nil
	}

	if err := uc.repo.DeleteProduct(ctx, id); err != nil {
		return err
	}
	if uc.cacheEnabled {
		uc.cacheWriter.deleteProduct(ctx, id)
		uc.cacheWriter.invalidateList(ctx)
	}
	uc.logger.Info("product deleted", domain.KV("id", id))
	return nil
}
