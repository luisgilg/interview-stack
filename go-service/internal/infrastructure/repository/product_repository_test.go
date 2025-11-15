package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/example/interview-stack/go-service/internal/domain"
	"github.com/example/interview-stack/go-service/internal/infrastructure/clock"
	"github.com/example/interview-stack/go-service/internal/mocks"
)

func TestProductRepositoryDeleteNotFound(t *testing.T) {
	ctx := context.Background()
	store := mocks.NewProductStoreMock(t)
	store.EXPECT().DeleteProduct(mock.Anything, "missing").Return(false, nil)

	repo := NewProductRepository(store, clock.NewFakeClock(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)))

	err := repo.DeleteProduct(ctx, "missing")
	require.Error(t, err)
	var domainErr *domain.Error
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, domain.ErrorCodeNotFound, domainErr.Code)
}

func TestProductRepositoryCreateInternalError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("boom")
	store := mocks.NewProductStoreMock(t)
	store.EXPECT().
		CreateProduct(mock.Anything, mock.AnythingOfType("domain.Product")).
		Return((*domain.Product)(nil), expectedErr)

	repo := NewProductRepository(store, clock.NewFakeClock(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)))

	_, err := repo.CreateProduct(ctx, domain.Product{Name: "Keyboard"})
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestProductRepositorySetsTimestamps(t *testing.T) {
	ctx := context.Background()
	fakeClock := clock.NewFakeClock(time.Date(2023, 5, 6, 7, 8, 9, 0, time.UTC))
	store := mocks.NewProductStoreMock(t)

	var recorded domain.Product
	store.EXPECT().
		CreateProduct(mock.Anything, mock.MatchedBy(func(product domain.Product) bool {
			recorded = product
			return true
		})).
		Return(&domain.Product{}, nil)

	repo := NewProductRepository(store, fakeClock)

	_, err := repo.CreateProduct(ctx, domain.Product{Name: "Keyboard", Price: 10})
	require.NoError(t, err)
	require.Equal(t, fakeClock.Now(), recorded.CreatedAt)
	require.Equal(t, fakeClock.Now(), recorded.UpdatedAt)
}
