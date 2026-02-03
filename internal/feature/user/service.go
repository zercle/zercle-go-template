package user

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
)

// Service provides user business logic.
type Service struct {
	repo   Repository
	logger zerolog.Logger
}

// NewService creates a new user service.
func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zerolog.Nop(),
	}
}

// NewServiceWithLogger creates a new user service with a custom logger.
func NewServiceWithLogger(repo Repository, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new user.
func (s *Service) Create(ctx context.Context, req *CreateUserInput) (*Response, error) {
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

	exists, err := s.repo.Exists(ctx, req.Email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to check email existence")
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateEmail
	}

	newUser, err := createDTOToDomain(req)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create user domain entity")
		return nil, err
	}

	created, err := s.repo.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to create user")
		return nil, err
	}

	s.logger.Info().Str("user_id", string(created.ID)).Str("email", created.Email).Msg("user created")

	return domainToDTO(created), nil
}

// GetByID retrieves a user by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, ErrUserNotFound
	}

	userID := ID(id)
	found, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to get user by ID")
		return nil, err
	}

	return domainToDTO(found), nil
}

// GetByEmail retrieves a user by email address.
func (s *Service) GetByEmail(ctx context.Context, email string) (*Response, error) {
	if email == "" {
		return nil, ErrUserNotFound
	}

	found, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("email", email).Msg("failed to get user by email")
		return nil, err
	}

	return domainToDTO(found), nil
}

// Update updates an existing user.
func (s *Service) Update(ctx context.Context, id string, req *UpdateUserInput) (*Response, error) {
	if id == "" {
		return nil, ErrUserNotFound
	}
	if req == nil {
		return nil, errors.New("update request is required")
	}

	userID := ID(id)

	existing, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to get user for update")
		return nil, err
	}

	if req.Email != nil && *req.Email != existing.Email {
		exists, err := s.repo.Exists(ctx, *req.Email)
		if err != nil {
			s.logger.Error().Err(err).Str("email", *req.Email).Msg("failed to check email existence")
			return nil, err
		}
		if exists {
			return nil, ErrDuplicateEmail
		}
	}

	updated := &User{
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
		updated.Status = Status(*req.Status)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to update user")
		return nil, err
	}

	s.logger.Info().Str("user_id", id).Msg("user updated")

	return domainToDTO(result), nil
}

// Delete removes a user by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrUserNotFound
	}

	userID := ID(id)

	exists, err := s.repo.ExistsByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to check user existence")
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return err
		}
		s.logger.Error().Err(err).Str("user_id", id).Msg("failed to delete user")
		return err
	}

	s.logger.Info().Str("user_id", id).Msg("user deleted")

	return nil
}

// List retrieves a paginated list of users.
func (s *Service) List(ctx context.Context, params *ListQuery) (*ListResponse, error) {
	if params == nil {
		params = &ListQuery{}
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	repoParams := listParamsToRepo(*params)

	result, err := s.repo.List(ctx, &repoParams)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list users")
		return nil, err
	}

	return listResultToDTO(result), nil
}

func domainToDTO(u *User) *Response {
	if u == nil {
		return nil
	}
	return &Response{
		ID:        string(u.ID),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func createDTOToDomain(dto *CreateUserInput) (*User, error) {
	if dto == nil {
		return nil, errors.New("input is nil")
	}
	return New(dto.Email, dto.Password, dto.FirstName, dto.LastName)
}

func listParamsToRepo(dto ListQuery) ListParams {
	params := ListParams{
		Email:  dto.Email,
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}

	if dto.Status != nil {
		params.Status = dto.Status
	}

	return params
}

func listResultToDTO(result *ListResult) *ListResponse {
	if result == nil {
		return nil
	}

	users := make([]*Response, 0, len(result.Users))
	for _, u := range result.Users {
		users = append(users, domainToDTO(u))
	}

	return &ListResponse{
		Users:  users,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}
}
