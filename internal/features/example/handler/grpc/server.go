// STUB FEATURE — delete internal/features/example to start your project.

package grpchandler

import (
	"context"

	pb "github.com/zercle/zercle-go-template/api/pb/example/v1"
	"github.com/zercle/zercle-go-template/internal/features/example/application"
	"github.com/zercle/zercle-go-template/internal/features/example/contract"
	apperrors "github.com/zercle/zercle-go-template/internal/platform/errors"
)

// nolint:wrapcheck // gRPC handlers return the shared mapper error directly.

// Server implements the example.v1.ExampleService gRPC contract.
type Server struct {
	pb.UnimplementedExampleServiceServer
	service application.Service
}

// NewServer returns a gRPC handler for the example feature.
func NewServer(service application.Service) *Server {
	return &Server{service: service}
}

// CreateItem creates a new item.
func (s *Server) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.Item, error) {
	if req == nil {
		return nil, apperrors.GRPCErr(apperrors.ErrInvalidInput)
	}

	resp, err := s.service.Create(ctx, &contract.CreateItemRequest{Name: req.Name})
	if err != nil {
		return nil, apperrors.GRPCErr(err)
	}

	return mapContractToPB(resp), nil
}

// GetItem retrieves an item by ID.
func (s *Server) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.Item, error) {
	if req == nil {
		return nil, apperrors.GRPCErr(apperrors.ErrInvalidInput)
	}

	resp, err := s.service.Get(ctx, req.Id)
	if err != nil {
		return nil, apperrors.GRPCErr(err)
	}

	return mapContractToPB(resp), nil
}

// ListItems returns a paginated list of items.
func (s *Server) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	if req == nil {
		return nil, apperrors.GRPCErr(apperrors.ErrInvalidInput)
	}

	resp, err := s.service.List(ctx, &contract.ListItemsRequest{Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, apperrors.GRPCErr(err)
	}

	out := &pb.ListItemsResponse{Items: make([]*pb.Item, len(resp.Items))}
	for i := range resp.Items {
		out.Items[i] = mapContractToPB(&resp.Items[i])
	}

	return out, nil
}

func mapContractToPB(resp *contract.ItemResponse) *pb.Item {
	if resp == nil {
		return nil
	}
	return &pb.Item{
		Id:        resp.ID,
		Name:      resp.Name,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}
