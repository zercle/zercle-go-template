package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// Update implements user.Usecase.Update.
func (s *service) Update(ctx context.Context, id string, req *user.UpdateUserDTO) (*user.UserDTO, error) {
	if id == "" {
		return nil, user.ErrUserNotFound
	}
	if req == nil {
		return nil, errors.New("update request is required")
	}

	userID := user.UserID(id)

	// Fetch existing entity
	existing, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to get user for update")
		return nil, err
	}

	// Check for duplicate email if email is being changed
	if req.Email != nil && *req.Email != existing.Email {
		exists, err := s.repo.Exists(ctx, *req.Email)
		if err != nil {
			s.logger.Error().Err(err).Str("email", *req.Email).Msg("failed to check email existence")
			return nil, err
		}
		if exists {
			return nil, user.ErrDuplicateEmail
		}
	}

	// Apply updates
	updated := &user.User{
		ID:           existing.ID,
		Email:        existing.Email,
		PasswordHash: existing.PasswordHash,
		FirstName:    existing.FirstName,
		LastName:     existing.LastName,
		Status:       existing.Status,
		CreatedAt:    existing.CreatedAt,
		UpdatedAt:    time.Now().UTC(),
	}

	if req.Email != nil {
		updated.Email = *req.Email
	}
	if req.FirstName != nil {
		updated.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		updated.LastName = *req.LastName
	}
	if req.Status != nil {
		updated.Status = user.UserStatus(*req.Status)
	}

	// Persist changes
	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to update user")
		return nil, err
	}

	s.logger.Info().Str("user_id", id).Msg("user updated")

	return domainToDTO(result), nil
}
