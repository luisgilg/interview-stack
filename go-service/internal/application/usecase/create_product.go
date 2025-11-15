package usecase

import (
	"context"

	"github.com/google/uuid"

	appcache "github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
)

// CreateProductUseCase encapsulates business logic for creating products.
type CreateProductUseCase struct {
	repo         domain.ProductRepository
	logger       domain.Logger
	clock        domain.Clock
	queue        domain.WriteQueue
	writeBehind  bool
	source       string
	cacheWriter  cacheWriter
	cacheEnabled bool
}

func NewCreateProductUseCase(
	repo domain.ProductRepository,
	logger domain.Logger,
	clock domain.Clock,
	cacheSvc *appcache.Service,
	queue domain.WriteQueue,
	writeBehind bool,
	source string,
) *CreateProductUseCase {
	return &CreateProductUseCase{
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

func (uc *CreateProductUseCase) Execute(ctx context.Context, input dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := input.ToDomain()
	if err := product.Validate(); err != nil {
		return nil, err
	}
	now := uc.clock.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	if uc.writeBehind && uc.queue != nil {
		if product.ID == "" {
			product.ID = uuid.NewString()
		}
		dtoProduct := dto.FromDomainProduct(&product)
		uc.cacheWriter.upsertProduct(ctx, &dtoProduct)
		uc.cacheWriter.invalidateList(ctx)
		eventPayload := product
		event := domain.WriteEvent{
			Type:      domain.WriteEventCreate,
			ID:        product.ID,
			Payload:   &eventPayload,
			Timestamp: now,
			Source:    uc.source,
		}
		if err := uc.queue.Enqueue(ctx, event); err != nil {
			return nil, domain.NewInternalError("failed to enqueue create event", err)
		}
		uc.logger.Info("product enqueued for creation", domain.KV("id", product.ID))
		return &dtoProduct, nil
	}

	created, err := uc.repo.CreateProduct(ctx, product)
	if err != nil {
		return nil, err
	}
	dtoProduct := dto.FromDomainProduct(created)
	if uc.cacheEnabled {
		uc.cacheWriter.upsertProduct(ctx, &dtoProduct)
		uc.cacheWriter.invalidateList(ctx)
	}
	uc.logger.Info("product created", domain.KV("id", created.ID))
	return &dtoProduct, nil
}
