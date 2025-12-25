package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/user"
	"github.com/zercle/zercle-go-template/domain/user/model"
	"github.com/zercle/zercle-go-template/domain/user/repository"
	"github.com/zercle/zercle-go-template/domain/user/request"
	userResponse "github.com/zercle/zercle-go-template/domain/user/response"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserAlreadyExists is returned when user already exists
	ErrUserAlreadyExists = errors.New("user already exists")
)

type userUseCase struct {
	repo user.Repository
	jwt  *config.JWTConfig
	log  *logger.Logger
}

// NewUserUseCase creates a new user use case with dependencies
func NewUserUseCase(repo user.Repository, jwt *config.JWTConfig, log *logger.Logger) user.Usecase {
	return &userUseCase{
		repo: repo,
		jwt:  jwt,
		log:  log,
	}
}

func (uc *userUseCase) Register(ctx context.Context, req request.RegisterUser) (*userResponse.LoginResponse, error) {
	// Check if email exists
	_, err := uc.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.log.Error("Failed to hash password", "error", err)
		return nil, err
	}

	// Create user
	user := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Phone:    req.Phone,
	}

	created, err := uc.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate token
	token, err := middleware.GenerateToken(created.ID.String(), created.Email, uc.jwt)
	if err != nil {
		uc.log.Error("Failed to generate token", "error", err)
		return nil, err
	}

	return &userResponse.LoginResponse{
		Token: token,
		User: userResponse.UserResponse{
			ID:        created.ID,
			Email:     created.Email,
			FullName:  created.FullName,
			Phone:     created.Phone,
			CreatedAt: created.CreatedAt,
			UpdatedAt: created.UpdatedAt,
		},
	}, nil
}

func (uc *userUseCase) Login(ctx context.Context, req request.LoginUser) (*userResponse.LoginResponse, error) {
	// Get user by email
	userModel, err := uc.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := middleware.GenerateToken(userModel.ID.String(), userModel.Email, uc.jwt)
	if err != nil {
		uc.log.Error("Failed to generate token", "error", err)
		return nil, err
	}

	return &userResponse.LoginResponse{
		Token: token,
		User: userResponse.UserResponse{
			ID:        userModel.ID,
			Email:     userModel.Email,
			FullName:  userModel.FullName,
			Phone:     userModel.Phone,
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		},
	}, nil
}

func (uc *userUseCase) GetProfile(ctx context.Context, id uuid.UUID) (*userResponse.UserResponse, error) {
	userModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &userResponse.UserResponse{
		ID:        userModel.ID,
		Email:     userModel.Email,
		FullName:  userModel.FullName,
		Phone:     userModel.Phone,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	}, nil
}

func (uc *userUseCase) UpdateProfile(ctx context.Context, id uuid.UUID, req request.UpdateUser) (*userResponse.UserResponse, error) {
	// Get existing user
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate and update fields
	if req.FullName != "" {
		if len(req.FullName) < 2 {
			return nil, errors.New("full_name must be at least 2 characters")
		}
		existing.FullName = req.FullName
	}
	if req.Phone != "" {
		existing.Phone = req.Phone
	}
	existing.UpdatedAt = time.Now()

	// Save
	updated, err := uc.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return &userResponse.UserResponse{
		ID:        updated.ID,
		Email:     updated.Email,
		FullName:  updated.FullName,
		Phone:     updated.Phone,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

func (uc *userUseCase) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *userUseCase) ListUsers(ctx context.Context, limit, offset int) (*userResponse.ListUsersResponse, error) {
	users, total, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	userResponses := make([]userResponse.UserResponse, len(users))
	for i, userModel := range users {
		userResponses[i] = userResponse.UserResponse{
			ID:        userModel.ID,
			Email:     userModel.Email,
			FullName:  userModel.FullName,
			Phone:     userModel.Phone,
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		}
	}

	return &userResponse.ListUsersResponse{
		Users: userResponses,
		Total: total,
	}, nil
}
