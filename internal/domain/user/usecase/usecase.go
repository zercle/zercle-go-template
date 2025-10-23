package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/domain/user"
	"github.com/zercle/zercle-go-template/internal/domain/user/entity"
	"github.com/zercle/zercle-go-template/internal/domain/user/repository"
	"github.com/zercle/zercle-go-template/internal/domain/user/request"
	userResponse "github.com/zercle/zercle-go-template/internal/domain/user/response"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
	"github.com/zercle/zercle-go-template/internal/infrastructure/password"
	"github.com/zercle/zercle-go-template/pkg/middleware"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type userUseCase struct {
	repo       user.UserRepository
	jwt        *config.JWTConfig
	log        *logger.Logger
	passworder *password.Hasher
}

func NewUserUseCase(repo user.UserRepository, jwt *config.JWTConfig, argon2idCfg *config.Argon2idConfig, log *logger.Logger) user.UserService {
	hasher := password.NewHasher(
		argon2idCfg.Memory,
		argon2idCfg.Iterations,
		argon2idCfg.SaltLength,
		argon2idCfg.KeyLength,
		argon2idCfg.Parallelism,
	)

	return &userUseCase{
		repo:       repo,
		jwt:        jwt,
		log:        log,
		passworder: hasher,
	}
}

func (uc *userUseCase) Register(ctx context.Context, req request.RegisterUser) (*userResponse.LoginResponse, error) {
	_, err := uc.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	hashedPassword, err := uc.passworder.HashPassword(req.Password)
	if err != nil {
		uc.log.Error("Failed to hash password", "error", err)
		return nil, err
	}

	user := &entity.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Phone:    req.Phone,
	}

	created, err := uc.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

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
	userModel, err := uc.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if _, err := uc.passworder.VerifyPassword(req.Password, userModel.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

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
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
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
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

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
