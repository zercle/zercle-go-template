package usecase

import (
	"context"
	"errors"

	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// Create implements user.Usecase.Create.
func (s *service) Create(ctx context.Context, req *user.CreateUserDTO) (*user.UserDTO, error) {
	// Validate input
	if req == nil {
		return nil, errors.New("create request is required")
	}
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}
	if req.FirstName == "" {
		return nil, errors.New("first name is required")
	}
	if req.LastName == "" {
		return nil, errors.New("last name is required")
	}

	// Check for duplicate email
	exists, err := s.repo.Exists(ctx, req.Email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to check email existence")
		return nil, err
	}
	if exists {
		return nil, user.ErrDuplicateEmail
	}

	// Create domain entity
	newUser, err := createDTOToDomain(req)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create user domain entity")
		return nil, err
	}

	// Persist via repository
	created, err := s.repo.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to create user")
		return nil, err
	}

	s.logger.Info().Str("user_id", string(created.ID)).Str("email", created.Email).Msg("user created")

	return domainToDTO(created), nil
}
