// Package usecase provides business logic implementations for the user feature.
// It orchestrates use cases and coordinates between repositories.
package usecase

//go:generate mockgen -source=$GOFILE -destination=./mocks/$GOFILE -package=mocks

import (
	"context"

	appErr "zercle-go-template/internal/errors"
	"zercle-go-template/internal/feature/user/domain"
	"zercle-go-template/internal/feature/user/dto"
	"zercle-go-template/internal/feature/user/repository"
	"zercle-go-template/internal/logger"
)

// UserUsecase defines the interface for user business logic.
type UserUsecase interface {
	// CreateUser creates a new user with the given request data.
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domain.User, error)

	// GetUser retrieves a user by their ID.
	GetUser(ctx context.Context, id string) (*domain.User, error)

	// GetUserByEmail retrieves a user by their email.
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

	// ListUsers retrieves a paginated list of users.
	ListUsers(ctx context.Context, page, limit int) ([]*domain.User, int, error)

	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*domain.User, error)

	// DeleteUser deletes a user by their ID.
	DeleteUser(ctx context.Context, id string) error

	// UpdatePassword updates a user's password.
	UpdatePassword(ctx context.Context, id string, req dto.UpdatePasswordRequest) error

	// Authenticate validates user credentials and returns the user.
	Authenticate(ctx context.Context, email, password string) (*domain.User, error)
}

// userUsecase implements UserUsecase.
type userUsecase struct {
	repo   repository.UserRepository
	logger logger.Logger
}

// NewUserUsecase creates a new user usecase instance.
func NewUserUsecase(repo repository.UserRepository, log logger.Logger) UserUsecase {
	return &userUsecase{
		repo:   repo,
		logger: log,
	}
}

// CreateUser implements UserUsecase.CreateUser.
func (u *userUsecase) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domain.User, error) {
	log := u.logger.WithContext(ctx).WithFields(
		logger.String("email", req.Email),
		logger.String("name", req.Name),
	)

	// Check if email already exists
	exists, err := u.repo.Exists(ctx, req.Email)
	if err != nil {
		log.Error("failed to check email existence", logger.Error(err))
		return nil, appErr.InternalError("failed to check email availability").WithCause(err)
	}
	if exists {
		log.Warn("attempt to create user with existing email")
		return nil, appErr.ConflictError("email already exists")
	}

	// Create the user
	user, err := domain.NewUser(req.Email, req.Name, req.Password)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			log.Warn("validation failed", logger.String("field", domainErr.Code))
			return nil, appErr.ValidationError(domainErr.Message)
		}
		log.Error("failed to create user", logger.Error(err))
		return nil, appErr.InternalError("failed to create user").WithCause(err)
	}

	// Save to repository
	if err := u.repo.Create(ctx, user); err != nil {
		log.Error("failed to save user", logger.Error(err))
		return nil, err
	}

	log.Info("user created successfully", logger.String("user_id", user.ID))
	return user, nil
}

// GetUser implements UserUsecase.GetUser.
func (u *userUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	log := u.logger.WithContext(ctx).WithFields(logger.String("user_id", id))

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if appErr.IsNotFoundError(err) {
			log.Warn("user not found")
			return nil, err
		}
		if appErr.IsValidationError(err) {
			log.Warn("invalid user ID format")
			return nil, err
		}
		log.Error("failed to get user", logger.Error(err))
		return nil, appErr.InternalError("failed to retrieve user").WithCause(err)
	}

	return user, nil
}

// GetUserByEmail implements UserUsecase.GetUserByEmail.
func (u *userUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	log := u.logger.WithContext(ctx).WithFields(logger.String("email", email))

	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if appErr.IsNotFoundError(err) {
			log.Warn("user not found")
			return nil, err
		}
		log.Error("failed to get user by email", logger.Error(err))
		return nil, appErr.InternalError("failed to retrieve user").WithCause(err)
	}

	return user, nil
}

// ListUsers implements UserUsecase.ListUsers.
func (u *userUsecase) ListUsers(ctx context.Context, page, limit int) ([]*domain.User, int, error) {
	log := u.logger.WithContext(ctx).WithFields(
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Get total count
	total, err := u.repo.Count(ctx)
	if err != nil {
		log.Error("failed to count users", logger.Error(err))
		return nil, 0, appErr.InternalError("failed to retrieve users").WithCause(err)
	}

	// Get users
	users, err := u.repo.GetAll(ctx, offset, limit)
	if err != nil {
		log.Error("failed to list users", logger.Error(err))
		return nil, 0, appErr.InternalError("failed to retrieve users").WithCause(err)
	}

	log.Info("users listed successfully", logger.Int("count", len(users)), logger.Int("total", total))
	return users, total, nil
}

// UpdateUser implements UserUsecase.UpdateUser.
func (u *userUsecase) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*domain.User, error) {
	log := u.logger.WithContext(ctx).WithFields(logger.String("user_id", id))

	// Get existing user
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if appErr.IsNotFoundError(err) {
			log.Warn("user not found for update")
			return nil, err
		}
		if appErr.IsValidationError(err) {
			log.Warn("invalid user ID format")
			return nil, err
		}
		log.Error("failed to get user for update", logger.Error(err))
		return nil, appErr.InternalError("failed to retrieve user").WithCause(err)
	}

	// Update fields
	if req.Name != "" {
		user.Update(req.Name)
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, appErr.ValidationError(domainErr.Message)
		}
		return nil, err
	}

	// Save changes
	if err := u.repo.Update(ctx, user); err != nil {
		log.Error("failed to update user", logger.Error(err))
		return nil, err
	}

	log.Info("user updated successfully")
	return user, nil
}

// DeleteUser implements UserUsecase.DeleteUser.
func (u *userUsecase) DeleteUser(ctx context.Context, id string) error {
	log := u.logger.WithContext(ctx).WithFields(logger.String("user_id", id))

	// Check if user exists
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if appErr.IsNotFoundError(err) {
			log.Warn("user not found for deletion")
			return err
		}
		if appErr.IsValidationError(err) {
			log.Warn("invalid user ID format")
			return err
		}
		log.Error("failed to get user for deletion", logger.Error(err))
		return appErr.InternalError("failed to retrieve user").WithCause(err)
	}

	// Delete user
	if err := u.repo.Delete(ctx, id); err != nil {
		log.Error("failed to delete user", logger.Error(err))
		return err
	}

	log.Info("user deleted successfully")
	return nil
}

// UpdatePassword implements UserUsecase.UpdatePassword.
func (u *userUsecase) UpdatePassword(ctx context.Context, id string, req dto.UpdatePasswordRequest) error {
	log := u.logger.WithContext(ctx).WithFields(logger.String("user_id", id))

	// Get existing user
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if appErr.IsNotFoundError(err) {
			log.Warn("user not found for password update")
			return err
		}
		if appErr.IsValidationError(err) {
			log.Warn("invalid user ID format")
			return err
		}
		log.Error("failed to get user for password update", logger.Error(err))
		return appErr.InternalError("failed to retrieve user").WithCause(err)
	}

	// Verify old password
	if !user.VerifyPassword(req.OldPassword) {
		log.Warn("invalid old password provided")
		return appErr.UnauthorizedError("invalid current password")
	}

	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return appErr.ValidationError(domainErr.Message)
		}
		return err
	}

	// Save changes
	if err := u.repo.Update(ctx, user); err != nil {
		log.Error("failed to update password", logger.Error(err))
		return err
	}

	log.Info("password updated successfully")
	return nil
}

// Authenticate implements UserUsecase.Authenticate.
func (u *userUsecase) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	log := u.logger.WithContext(ctx).WithFields(logger.String("email", email))

	// Get user by email
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if appErr.IsNotFoundError(err) {
			log.Warn("authentication failed: user not found")
			return nil, appErr.UnauthorizedError("invalid credentials")
		}
		log.Error("failed to get user for authentication", logger.Error(err))
		return nil, appErr.InternalError("authentication failed").WithCause(err)
	}

	// Verify password
	if !user.VerifyPassword(password) {
		log.Warn("authentication failed: invalid password")
		return nil, appErr.UnauthorizedError("invalid credentials")
	}

	log.Info("user authenticated successfully")
	return user, nil
}
