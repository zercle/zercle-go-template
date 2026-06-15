// STUB FEATURE — delete internal/features/example to start your project.

package grpchandler

import (
	"context"
	"time"

	"github.com/google/uuid"

	pb "github.com/zercle/zercle-go-template/api/pb/example/v1"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// nolint:wrapcheck // gRPC handlers return the shared mapper error directly.

// Server implements the example.v1.ExampleService gRPC contract.
type Server struct {
	pb.UnimplementedExampleServiceServer
	service domain.Service
}

// NewServer returns a gRPC handler for the example feature.
func NewServer(service domain.Service) *Server {
	return &Server{service: service}
}

// CreateItem creates a new item.
func (s *Server) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.Item, error) {
	item, err := s.service.Create(ctx, req.Name)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}

	return mapDomainToPB(item), nil
}

// GetItem retrieves an item by ID.
func (s *Server) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.Item, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, sharederrors.GRPCErr(domain.ErrInvalidName)
	}

	item, err := s.service.Get(ctx, id)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}

	return mapDomainToPB(item), nil
}

// ListItems returns a paginated list of items.
func (s *Server) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	items, err := s.service.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, sharederrors.GRPCErr(err)
	}

	resp := &pb.ListItemsResponse{Items: make([]*pb.Item, len(items))}
	for i, item := range items {
		resp.Items[i] = mapDomainToPB(&item)
	}

	return resp, nil
}

func mapDomainToPB(item *domain.Item) *pb.Item {
	return &pb.Item{
		Id:        item.ID.String(),
		Name:      item.Name,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
}
