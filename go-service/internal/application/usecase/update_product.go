package usecase

import (
	"context"

	appcache "github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

// UpdateProductUseCase updates an existing product.
type UpdateProductUseCase struct {
	repo         domain.ProductRepository
	logger       domain.Logger
	clock        domain.Clock
	queue        domain.WriteQueue
	writeBehind  bool
	source       string
	cacheWriter  cacheWriter
	cacheEnabled bool
}

func NewUpdateProductUseCase(
	repo domain.ProductRepository,
	logger domain.Logger,
	clock domain.Clock,
	cacheSvc *appcache.Service,
	queue domain.WriteQueue,
	writeBehind bool,
	source string,
) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		repo:         repo,
		logger:       logger,
		clock:        clock,
		queue:        queue,
		writeBehind:  writeBehind,
		source:       source,
		cacheWriter:  newCacheWriter(cacheSvc, logger),
		cacheEnabled: cacheSvc != nil && cacheSvc.Enabled(),
	}
}

func (uc *UpdateProductUseCase) Execute(ctx context.Context, id string, input dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product := input.ToDomain()
	if err := product.Validate(); err != nil {
		return nil, err
	}
	if uc.writeBehind && uc.queue != nil {
		current, err := uc.repo.GetProduct(ctx, id)
		if err != nil {
			return nil, err
		}
		if current == nil {
			return nil, domain.NewNotFoundError("product not found")
		}
		current.Name = product.Name
		current.Price = product.Price
		current.Tags = product.Tags
		current.UpdatedAt = uc.clock.Now()
		dtoProduct := dto.FromDomainProduct(current)
		uc.cacheWriter.upsertProduct(ctx, &dtoProduct)
		uc.cacheWriter.invalidateList(ctx)
		payload := *current
		event := domain.WriteEvent{
			Type:      domain.WriteEventUpdate,
			ID:        id,
			Payload:   &payload,
			Timestamp: current.UpdatedAt,
			Source:    uc.source,
		}
		if err := uc.queue.Enqueue(ctx, event); err != nil {
			return nil, domain.NewInternalError("failed to enqueue update event", err)
		}
		uc.logger.Info("product enqueued for update", domain.KV("id", id))
		return &dtoProduct, nil
	}

	product.UpdatedAt = uc.clock.Now()

	updated, err := uc.repo.UpdateProduct(ctx, id, product)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, domain.NewNotFoundError("product not found")
	}
	dtoProduct := dto.FromDomainProduct(updated)
	if uc.cacheEnabled {
		uc.cacheWriter.upsertProduct(ctx, &dtoProduct)
		uc.cacheWriter.invalidateList(ctx)
	}
	uc.logger.Info("product updated", domain.KV("id", id))
	return &dtoProduct, nil
}
