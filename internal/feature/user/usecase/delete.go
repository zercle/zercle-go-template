package usecase

import (
	"context"
	"errors"

	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// Delete implements user.Usecase.Delete.
func (s *service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return user.ErrUserNotFound
	}

	userID := user.UserID(id)

	// Check if user exists first
	exists, err := s.repo.ExistsByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to check user existence")
		return err
	}
	if !exists {
		return user.ErrUserNotFound
	}

	// Delete via repository
	if err := s.repo.Delete(ctx, userID); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to delete user")
		return err
	}

	s.logger.Info().Str("user_id", id).Msg("user deleted")

	return nil
}
