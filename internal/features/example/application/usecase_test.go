//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/example/application"
	"github.com/zercle/zercle-go-template/internal/features/example/contract"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/port/mock"
)

func TestService_Create_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchItemName("stub")).Return(nil)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateItemRequest{Name: "stub"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "stub", resp.Name)
	parsed, err := uuid.Parse(resp.ID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, parsed)
	require.NotEmpty(t, resp.CreatedAt)
	require.NotEmpty(t, resp.UpdatedAt)
}

func TestService_Create_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := application.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Create(ctx, &contract.CreateItemRequest{Name: ""})

	require.ErrorIs(t, err, domain.ErrInvalidName)
	require.Nil(t, resp)
}

func TestService_Create_WhitespaceName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateItemRequest{Name: "   "})
	require.ErrorIs(t, err, domain.ErrInvalidName)
	require.Nil(t, resp)
}

func TestService_Get_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	now := time.Now().UTC().Truncate(time.Second)
	expected := &domain.Item{ID: id, Name: "found", CreatedAt: now, UpdatedAt: now}
	repo.EXPECT().GetByID(ctx, id).Return(expected, nil)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Get(ctx, id.String())

	require.NoError(t, err)
	require.Equal(t, &contract.ItemResponse{
		ID:        id.String(),
		Name:      "found",
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	}, resp)
}

func TestService_Get_MapsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrItemNotFound)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Get(ctx, id.String())

	require.ErrorIs(t, err, domain.ErrItemNotFound)
	require.Nil(t, resp)
}

func TestService_Get_InvalidIDRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := application.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Get(ctx, "not-a-uuid")
	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)

	resp, err = svc.Get(ctx, uuid.Nil.String())
	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)
}

func TestService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	now := time.Now().UTC().Truncate(time.Second)
	expected := []domain.Item{{ID: uuid.New(), Name: "one", CreatedAt: now, UpdatedAt: now}}
	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(expected, nil)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListItemsRequest{Limit: 10, Offset: 5})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, expected[0].ID.String(), resp.Items[0].ID)
	require.Equal(t, "one", resp.Items[0].Name)
	require.Equal(t, now.Format(time.RFC3339), resp.Items[0].CreatedAt)
}

func TestService_List_EmptyResultKeepsNonNullItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(nil, nil)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListItemsRequest{Limit: 10, Offset: 5})

	require.NoError(t, err)
	require.NotNil(t, resp.Items)
	require.Empty(t, resp.Items)
}

func TestService_List_AppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	expected := []domain.Item{{ID: uuid.New(), Name: "default"}}
	repo.EXPECT().List(ctx, int32(20), int32(5)).Return(expected, nil)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListItemsRequest{Limit: 0, Offset: 5})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "default", resp.Items[0].Name)
}

func TestService_List_ClampsOverMaxLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	expected := []domain.Item{{ID: uuid.New(), Name: "clamped"}}
	repo.EXPECT().List(ctx, int32(100), int32(0)).Return(expected, nil)

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListItemsRequest{Limit: 999, Offset: -5})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "clamped", resp.Items[0].Name)
}

func TestService_List_RespectsConfiguredMaxPageSize(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	expected := []domain.Item{{ID: uuid.New(), Name: "clamped"}}
	repo.EXPECT().List(ctx, int32(50), int32(0)).Return(expected, nil)

	svc := application.NewUsecase(repo, 10, 50, 255)
	resp, err := svc.List(ctx, &contract.ListItemsRequest{Limit: 999})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "clamped", resp.Items[0].Name)
}

func TestService_Create_RepositoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchItemName("stub")).Return(errors.New("boom"))

	svc := application.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateItemRequest{Name: "stub"})

	require.Error(t, err)
	require.Nil(t, resp)
}

func matchItemName(name string) any {
	return matchItemByName{name: name}
}

type matchItemByName struct {
	name string
}

func (m matchItemByName) Matches(x any) bool {
	item, ok := x.(*domain.Item)
	return ok && item.Name == m.name
}

func (m matchItemByName) String() string {
	return "is item named " + m.name
}
