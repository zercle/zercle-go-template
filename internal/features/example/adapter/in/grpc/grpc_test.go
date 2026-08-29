//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package grpchandler_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pb "github.com/zercle/zercle-go-template/api/pb/example/v1"
	"github.com/zercle/zercle-go-template/internal/features/example/application/mock"
	"github.com/zercle/zercle-go-template/internal/features/example/contract"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	grpchandler "github.com/zercle/zercle-go-template/internal/features/example/adapter/in/grpc"
)

func TestServer_CreateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	respItem := &contract.ItemResponse{ID: uuid.New().String(), Name: "grpc-item"}
	svc.EXPECT().Create(gomock.Any(), &contract.CreateItemRequest{Name: "grpc-item"}).Return(respItem, nil)

	resp, err := server.CreateItem(context.Background(), &pb.CreateItemRequest{Name: "grpc-item"})
	require.NoError(t, err)
	assert.Equal(t, respItem.ID, resp.Id)
	assert.Equal(t, respItem.Name, resp.Name)
}

func TestServer_CreateItem_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	svc.EXPECT().Create(gomock.Any(), &contract.CreateItemRequest{Name: "bad"}).Return(nil, domain.ErrInvalidName)

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
	respItem := &contract.ItemResponse{ID: id.String(), Name: "grpc-item"}
	svc.EXPECT().Get(gomock.Any(), id.String()).Return(respItem, nil)

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
	svc.EXPECT().Get(gomock.Any(), id.String()).Return(nil, domain.ErrItemNotFound)

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
	respItems := &contract.ListItemsResponse{Items: []contract.ItemResponse{{ID: id.String(), Name: "grpc-item"}}}
	svc.EXPECT().List(gomock.Any(), &contract.ListItemsRequest{Limit: 10, Offset: 0}).Return(respItems, nil)

	resp, err := server.ListItems(context.Background(), &pb.ListItemsRequest{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
}

func TestServer_GetItem_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mock.NewMockService(ctrl)
	server := grpchandler.NewServer(svc)

	svc.EXPECT().Get(gomock.Any(), "not-a-uuid").Return(nil, domain.ErrInvalidID)

	resp, err := server.GetItem(context.Background(), &pb.GetItemRequest{Id: "not-a-uuid"})
	require.Error(t, err)
	assert.Nil(t, resp)
}
