package usecase

import (
	"context"
	"errors"

	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// GetByID implements user.Usecase.GetByID.
func (s *service) GetByID(ctx context.Context, id string) (*user.UserDTO, error) {
	if id == "" {
		return nil, user.ErrUserNotFound
	}

	userID := user.UserID(id)
	found, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to get user by ID")
		return nil, err
	}

	return domainToDTO(found), nil
}

// GetByEmail implements user.Usecase.GetByEmail.
func (s *service) GetByEmail(ctx context.Context, email string) (*user.UserDTO, error) {
	if email == "" {
		return nil, user.ErrUserNotFound
	}

	found, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("email", email).Msg("failed to get user by email")
		return nil, err
	}

	return domainToDTO(found), nil
}
