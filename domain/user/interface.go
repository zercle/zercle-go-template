//go:generate go run go.uber.org/mock/mockgen@latest -source=$GOFILE -destination=mock/$GOFILE -package=mock

package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/user/model"
	"github.com/zercle/zercle-go-template/domain/user/request"
	"github.com/zercle/zercle-go-template/domain/user/response"
)

// Repository defines the data access interface for users
type Repository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.User, int, error)
}

// Usecase defines the business logic interface for users
type Usecase interface {
	Register(ctx context.Context, req request.RegisterUser) (*response.LoginResponse, error)
	Login(ctx context.Context, req request.LoginUser) (*response.LoginResponse, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*response.UserResponse, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req request.UpdateUser) (*response.UserResponse, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) (*response.ListUsersResponse, error)
}

// Handler defines the HTTP handler interface for users
type Handler interface {
	Register(c echo.Context) error
	Login(c echo.Context) error
	GetProfile(c echo.Context) error
	UpdateProfile(c echo.Context) error
	DeleteAccount(c echo.Context) error
	ListUsers(c echo.Context) error
}
