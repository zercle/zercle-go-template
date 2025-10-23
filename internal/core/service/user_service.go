package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/domain"
	domerrors "github.com/zercle/zercle-go-template/internal/core/domain/errors"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/pkg/utils/password"
)

type userService struct {
	repo      port.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo port.UserRepository, jwtSecret string, jwtExpiry time.Duration) port.UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Register registers a new user with the given details.
// It checks for duplicate email and hashes the password before saving.
func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	// Check if email exists
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return nil, domerrors.ErrDuplicate
	}

	// Hash password
	hashed, err := password.Hash(req.Password)
	if err != nil {
		return nil, domerrors.ErrInternalServer
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, domerrors.ErrInternalServer
	}

	user := &domain.User{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.mapToDTO(user), nil
}

// Login authenticates a user and returns a JWT token.
func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (string, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", domerrors.ErrInvalidCreds
	}

	match, err := password.Verify(req.Password, user.Password)
	if err != nil || !match {
		return "", domerrors.ErrInvalidCreds
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(s.jwtExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GetProfile retrieves the user profile by ID.
func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, domerrors.ErrNotFound
	}

	return s.mapToDTO(user), nil
}

func (s *userService) mapToDTO(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
