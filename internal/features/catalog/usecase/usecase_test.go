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

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository/mock"
	"github.com/zercle/zercle-go-template/internal/features/catalog/usecase"
)

func TestService_Create_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchProductName("widget")).Return(nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateProductRequest{Name: "  widget  ", PriceCents: 250, Stock: 3})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "widget", resp.Name)
	require.Equal(t, int32(250), resp.PriceCents)
	require.Equal(t, int32(3), resp.Stock)
	parsed, err := uuid.Parse(resp.ID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil(), parsed)
	require.NotEmpty(t, resp.CreatedAt)
	require.NotEmpty(t, resp.UpdatedAt)
}

func TestService_Create_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Create(ctx, &contract.CreateProductRequest{Name: "   ", PriceCents: 100})

	require.ErrorIs(t, err, domain.ErrInvalidProductName)
	require.Nil(t, resp)
}

func TestService_Create_OverLengthName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 5)

	resp, err := svc.Create(ctx, &contract.CreateProductRequest{Name: "abcdef", PriceCents: 100})

	require.ErrorIs(t, err, domain.ErrInvalidProductName)
	require.Nil(t, resp)
}

func TestService_Create_NonPositivePrice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Create(ctx, &contract.CreateProductRequest{Name: "widget", PriceCents: 0})

	require.ErrorIs(t, err, domain.ErrInvalidPrice)
	require.Nil(t, resp)
}

func TestService_Create_RepositoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchProductName("widget")).Return(errors.New("boom"))

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateProductRequest{Name: "widget", PriceCents: 100})

	require.Error(t, err)
	require.Nil(t, resp)
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
	resp, err := svc.Get(ctx, id.String())

	require.NoError(t, err)
	require.Equal(t, &contract.ProductResponse{
		ID:         id.String(),
		Name:       "found",
		PriceCents: 125,
		Stock:      4,
		CreatedAt:  now.Format(time.RFC3339),
		UpdatedAt:  now.Format(time.RFC3339),
	}, resp)
}

func TestService_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrProductNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Get(ctx, id.String())

	require.ErrorIs(t, err, domain.ErrProductNotFound)
	require.Nil(t, resp)
}

func TestService_Get_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Get(ctx, "not-a-uuid")

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)
}

func TestService_List_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	now := time.Now().UTC().Truncate(time.Second)
	expected := []domain.Product{{ID: uuid.New(), Name: "one", PriceCents: 100, Stock: 1, CreatedAt: now, UpdatedAt: now}}
	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(expected, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListProductsRequest{Limit: 10, Offset: 5})

	require.NoError(t, err)
	require.Len(t, resp.Products, 1)
	require.Equal(t, expected[0].ID.String(), resp.Products[0].ID)
	require.Equal(t, "one", resp.Products[0].Name)
	require.Equal(t, now.Format(time.RFC3339), resp.Products[0].CreatedAt)
}

func TestService_List_AppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(20), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListProductsRequest{Limit: 0, Offset: 0})

	require.NoError(t, err)
	require.NotNil(t, resp.Products)
	require.Empty(t, resp.Products)
}

func TestService_List_ClampsOverMaxLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(50), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 10, 50, 255)
	resp, err := svc.List(ctx, &contract.ListProductsRequest{Limit: 999})

	require.NoError(t, err)
	require.Empty(t, resp.Products)
}

func TestService_List_ClampsNegativeOffset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(10), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListProductsRequest{Limit: 10, Offset: -3})

	require.NoError(t, err)
	require.Empty(t, resp.Products)
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
