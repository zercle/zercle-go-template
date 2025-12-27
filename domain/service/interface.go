//go:generate go run go.uber.org/mock/mockgen@latest -source=$GOFILE -destination=mock/$GOFILE -package=mock

package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/service/model"
	"github.com/zercle/zercle-go-template/domain/service/request"
	"github.com/zercle/zercle-go-template/domain/service/response"
)

// Repository defines the data access interface for services
type Repository interface {
	Create(ctx context.Context, service *model.Service) (*model.Service, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error)
	Update(ctx context.Context, service *model.Service) (*model.Service, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, isActive bool, limit, offset int) ([]*model.Service, error)
	SearchByName(ctx context.Context, name string, isActive bool, limit int) ([]*model.Service, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Service, error)
}

// Usecase defines the business logic interface for services
type Usecase interface {
	CreateService(ctx context.Context, req request.CreateService) (*response.ServiceResponse, error)
	GetService(ctx context.Context, id uuid.UUID) (*response.ServiceResponse, error)
	UpdateService(ctx context.Context, id uuid.UUID, req request.UpdateService) (*response.ServiceResponse, error)
	DeleteService(ctx context.Context, id uuid.UUID) error
	ListServices(ctx context.Context, req request.ListServices) (*response.ListServicesResponse, error)
	SearchServices(ctx context.Context, name string, isActive bool, limit int) ([]response.ServiceResponse, error)
}

// Handler defines the HTTP handler interface for services
type Handler interface {
	CreateService(c echo.Context) error
	GetService(c echo.Context) error
	UpdateService(c echo.Context) error
	DeleteService(c echo.Context) error
	ListServices(c echo.Context) error
	SearchServices(c echo.Context) error
}
