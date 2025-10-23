//go:generate go run go.uber.org/mock/mockgen@latest -source=$GOFILE -destination=mock/$GOFILE -package=mock

package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/internal/domain/user/entity"
	"github.com/zercle/zercle-go-template/internal/domain/user/request"
	"github.com/zercle/zercle-go-template/internal/domain/user/response"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, int, error)
}

type UserService interface {
	Register(ctx context.Context, req request.RegisterUser) (*response.LoginResponse, error)
	Login(ctx context.Context, req request.LoginUser) (*response.LoginResponse, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*response.UserResponse, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req request.UpdateUser) (*response.UserResponse, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) (*response.ListUsersResponse, error)
}

type UserHandler interface {
	Register(c echo.Context) error
	Login(c echo.Context) error
	GetProfile(c echo.Context) error
	UpdateProfile(c echo.Context) error
	DeleteAccount(c echo.Context) error
	ListUsers(c echo.Context) error
	RegisterRoutes(api, protected *echo.Group)
}
