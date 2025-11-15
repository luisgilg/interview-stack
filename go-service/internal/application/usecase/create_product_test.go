package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/example/interview-stack/go-service/internal/application/dto"
	"github.com/example/interview-stack/go-service/internal/domain"
	"github.com/example/interview-stack/go-service/internal/infrastructure/clock"
	"github.com/example/interview-stack/go-service/internal/infrastructure/repository"
	"github.com/example/interview-stack/go-service/internal/mocks"
)

const serviceName = "go-service"

type noopLogger struct{}

func (noopLogger) Info(string, ...domain.Field)         {}
func (noopLogger) Warn(string, ...domain.Field)         {}
func (noopLogger) Error(string, error, ...domain.Field) {}

func TestCreateProductUseCase(t *testing.T) {
	t.Run("creates product successfully", func(t *testing.T) {
		fixedTime := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)
		store := mocks.NewProductStoreMock(t)
		store.EXPECT().
			CreateProduct(mock.Anything, mock.MatchedBy(func(product domain.Product) bool {
				require.Equal(t, fixedTime, product.CreatedAt)
				require.Equal(t, fixedTime, product.UpdatedAt)
				return true
			})).
			Return(&domain.Product{ID: "abc", Name: "Keyboard", Price: 10, CreatedAt: fixedTime, UpdatedAt: fixedTime}, nil)

		repo := repository.NewProductRepository(store, clock.NewFakeClock(fixedTime))
		uc := NewCreateProductUseCase(repo, noopLogger{}, clock.NewFakeClock(fixedTime), nil, nil, false, serviceName)

		result, err := uc.Execute(context.Background(), dto.CreateProductRequest{
			Name:  "Keyboard",
			Price: 10,
		})
		require.NoError(t, err)
		require.Equal(t, "abc", result.ID)
	})

	t.Run("validation failure", func(t *testing.T) {
		store := mocks.NewProductStoreMock(t)
		repo := repository.NewProductRepository(store, clock.NewFakeClock(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)))
		uc := NewCreateProductUseCase(repo, noopLogger{}, clock.NewFakeClock(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)), nil, nil, false, serviceName)

		_, err := uc.Execute(context.Background(), dto.CreateProductRequest{
			Name:  "",
			Price: 0,
		})
		require.Error(t, err)
		var domainErr *domain.Error
		require.ErrorAs(t, err, &domainErr)
		require.Equal(t, domain.ErrorCodeValidation, domainErr.Code)
	})
}

type queueStub struct {
	events []domain.WriteEvent
	err    error
}

func (q *queueStub) Enqueue(_ context.Context, event domain.WriteEvent) error {
	if q.err != nil {
		return q.err
	}
	q.events = append(q.events, event)
	return nil
}

func TestCreateProductWriteBehind(t *testing.T) {
	fixed := time.Date(2024, 2, 2, 3, 4, 5, 0, time.UTC)
	queue := &queueStub{}
	clock := clock.NewFakeClock(fixed)
	uc := NewCreateProductUseCase(nil, noopLogger{}, clock, nil, queue, true, serviceName)

	result, err := uc.Execute(context.Background(), dto.CreateProductRequest{
		Name:  "Cached Keyboard",
		Price: 75,
	})
	require.NoError(t, err)
	require.Len(t, queue.events, 1)
	require.Equal(t, serviceName, queue.events[0].Source)
	require.Equal(t, domain.WriteEventCreate, queue.events[0].Type)
	require.Equal(t, result.ID, queue.events[0].ID)
}
