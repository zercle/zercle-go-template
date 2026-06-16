//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package grpchandler_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pb "github.com/zercle/zercle-go-template/api/pb/example/v1"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	grpchandler "github.com/zercle/zercle-go-template/internal/features/example/handler/grpc"
	"github.com/zercle/zercle-go-template/internal/features/example/service/mock"
)

func TestServer_CreateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	item := &domain.Item{ID: uuid.New(), Name: "grpc-item", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	svc.EXPECT().Create(gomock.Any(), "grpc-item").Return(item, nil)

	resp, err := server.CreateItem(context.Background(), &pb.CreateItemRequest{Name: "grpc-item"})
	require.NoError(t, err)
	assert.Equal(t, item.ID.String(), resp.Id)
	assert.Equal(t, item.Name, resp.Name)
}

func TestServer_CreateItem_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	svc.EXPECT().Create(gomock.Any(), "bad").Return(nil, domain.ErrInvalidName)

	resp, err := server.CreateItem(context.Background(), &pb.CreateItemRequest{Name: "bad"})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_GetItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	item := &domain.Item{ID: id, Name: "grpc-item", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	svc.EXPECT().Get(gomock.Any(), id).Return(item, nil)

	resp, err := server.GetItem(context.Background(), &pb.GetItemRequest{Id: id.String()})
	require.NoError(t, err)
	assert.Equal(t, id.String(), resp.Id)
}

func TestServer_GetItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	svc.EXPECT().Get(gomock.Any(), id).Return(nil, domain.ErrItemNotFound)

	resp, err := server.GetItem(context.Background(), &pb.GetItemRequest{Id: id.String()})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_ListItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	id := uuid.New()
	items := []domain.Item{{ID: id, Name: "grpc-item", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}
	svc.EXPECT().List(gomock.Any(), int32(10), int32(0)).Return(items, nil)

	resp, err := server.ListItems(context.Background(), &pb.ListItemsRequest{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
}

func TestServer_GetItem_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	resp, err := server.GetItem(context.Background(), &pb.GetItemRequest{Id: "not-a-uuid"})
	require.Error(t, err)
	assert.Nil(t, resp)
}
