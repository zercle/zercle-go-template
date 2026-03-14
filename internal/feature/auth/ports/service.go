//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../mocks/service.mock.go -package=mocks

package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
)

// RegisterInput holds the data needed for user registration.
type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

// LoginInput holds the data needed for user login.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult holds the result of an authentication operation.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
	ExpiresAt    int64
}

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	ValidateToken(ctx context.Context, tokenString string) (*domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}
