package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

type AuthServiceInterface interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	ValidateToken(ctx context.Context, tokenString string) (*domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}
