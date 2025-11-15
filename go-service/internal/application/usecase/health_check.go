package usecase

import (
	"context"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// HealthCheckUseCase verifies downstream dependencies.
type HealthCheckUseCase struct {
	repo domain.ProductRepository
}

func NewHealthCheckUseCase(repo domain.ProductRepository) *HealthCheckUseCase {
	return &HealthCheckUseCase{repo: repo}
}

func (uc *HealthCheckUseCase) Execute(ctx context.Context) error {
	return uc.repo.Health(ctx)
}
