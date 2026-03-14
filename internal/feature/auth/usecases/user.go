package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
	"github.com/zercle/zercle-go-template/internal/feature/auth/ports"
)

// UserService handles user-related business logic.
type UserService struct {
	userRepo ports.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo ports.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserByID retrieves a user by ID.
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.PublicUser, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToPublic(), nil
}
