//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/port/mock"
	"github.com/zercle/zercle-go-template/internal/features/catalog/usecase"
)

func TestService_Create_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchProductName("widget")).Return(nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	product, err := svc.Create(ctx, "  widget  ", 250, 3)

	require.NoError(t, err)
	require.NotNil(t, product)
	require.Equal(t, "widget", product.Name)
	require.Equal(t, int32(250), product.PriceCents)
	require.Equal(t, int32(3), product.Stock)
	require.NotEqual(t, uuid.Nil(), product.ID)
	require.False(t, product.CreatedAt.IsZero())
	require.False(t, product.UpdatedAt.IsZero())
}

func TestService_Create_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	product, err := svc.Create(ctx, "   ", 100, 0)

	require.ErrorIs(t, err, domain.ErrInvalidProductName)
	require.Nil(t, product)
}

func TestService_Create_OverLengthName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 5)

	product, err := svc.Create(ctx, "abcdef", 100, 0)

	require.ErrorIs(t, err, domain.ErrInvalidProductName)
	require.Nil(t, product)
}

func TestService_Create_NonPositivePrice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	product, err := svc.Create(ctx, "widget", 0, 0)

	require.ErrorIs(t, err, domain.ErrInvalidPrice)
	require.Nil(t, product)
}

func TestService_Create_RepositoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchProductName("widget")).Return(errors.New("boom"))

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	product, err := svc.Create(ctx, "widget", 100, 0)

	require.Error(t, err)
	require.Nil(t, product)
}

func TestService_Get_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	now := time.Now().UTC().Truncate(time.Second)
	expected := &domain.Product{ID: id, Name: "found", PriceCents: 125, Stock: 4, CreatedAt: now, UpdatedAt: now}
	repo.EXPECT().GetByID(ctx, id).Return(expected, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	product, err := svc.Get(ctx, id.String())

	require.NoError(t, err)
	require.Equal(t, expected, product)
}

func TestService_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrProductNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	product, err := svc.Get(ctx, id.String())

	require.ErrorIs(t, err, domain.ErrProductNotFound)
	require.Nil(t, product)
}

func TestService_Get_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	product, err := svc.Get(ctx, "not-a-uuid")

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, product)
}

func TestService_List_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	now := time.Now().UTC().Truncate(time.Second)
	expected := []domain.Product{{ID: uuid.New(), Name: "one", PriceCents: 100, Stock: 1, CreatedAt: now, UpdatedAt: now}}
	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(expected, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	products, err := svc.List(ctx, 10, 5)

	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Equal(t, expected[0].ID, products[0].ID)
	require.Equal(t, "one", products[0].Name)
	require.Equal(t, now, products[0].CreatedAt)
}

func TestService_List_AppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(20), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	products, err := svc.List(ctx, 0, 0)

	require.NoError(t, err)
	require.Empty(t, products)
}

func TestService_List_ClampsOverMaxLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(50), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 10, 50, 255)
	products, err := svc.List(ctx, 999, 0)

	require.NoError(t, err)
	require.Empty(t, products)
}

func TestService_List_ClampsNegativeOffset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(10), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	products, err := svc.List(ctx, 10, -3)

	require.NoError(t, err)
	require.Empty(t, products)
}

func matchProductName(name string) any {
	return matchProductByName{name: name}
}

type matchProductByName struct {
	name string
}

func (m matchProductByName) Matches(x any) bool {
	product, ok := x.(*domain.Product)
	return ok && product.Name == m.name
}

func (m matchProductByName) String() string {
	return "is product named " + m.name
}
