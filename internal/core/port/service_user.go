package port

import (
	"context"

	"github.com/google/uuid"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
)

//go:generate mockgen -destination=./mocks/$GOFILE -package=mocks -source=$GOFILE

// UserService defines the input port for User operations.
type UserService interface {
	Register(ctx context.Context, req *userDto.RegisterRequest) (*userDto.UserResponse, error)
	Login(ctx context.Context, req *userDto.LoginRequest) (string, error) // Returns JWT token
	GetProfile(ctx context.Context, userID uuid.UUID) (*userDto.UserResponse, error)
}
