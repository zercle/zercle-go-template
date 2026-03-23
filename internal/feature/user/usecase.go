package user

import (
	"context"
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// Usecase defines the contract for user business operations.
// This interface is used by handlers to invoke business logic,
// following the hexagonal architecture pattern where the application
// layer defines ports (interfaces) that adapters (handlers) use.
type Usecase interface {
	// Create creates a new user with the given parameters.
	// Returns the created user response or an error if creation fails.
	// Returns ErrDuplicateEmail if the email is already registered.
	Create(ctx context.Context, req *CreateUserDTO) (*UserDTO, error)

	// GetByID retrieves a user by their ID.
	// Returns the user response or ErrUserNotFound if not found.
	GetByID(ctx context.Context, id string) (*UserDTO, error)

	// GetByEmail retrieves a user by their email address.
	// Returns the user response or ErrUserNotFound if not found.
	GetByEmail(ctx context.Context, email string) (*UserDTO, error)

	// Update modifies an existing user.
	// Returns the updated user response or an error if update fails.
	// Returns ErrUserNotFound if the user does not exist.
	// Returns ErrDuplicateEmail if the new email is already registered.
	Update(ctx context.Context, id string, req *UpdateUserDTO) (*UserDTO, error)

	// Delete removes a user.
	// Returns ErrUserNotFound if the user does not exist.
	Delete(ctx context.Context, id string) error

	// List returns paginated users with optional filtering.
	List(ctx context.Context, params *ListParamsDTO) (*ListResultDTO, error)
}

// CreateUserDTO contains data for creating a new user.
type CreateUserDTO struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// UpdateUserDTO contains optional fields for updating a user.
type UpdateUserDTO struct {
	Email     *string `json:"email" validate:"omitempty,email"`
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	Status    *string `json:"status" validate:"omitempty,oneof=active inactive suspended"`
}

// UserDTO represents a user in API responses.
type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListParamsDTO contains query parameters for listing users.
type ListParamsDTO struct {
	Email  string  `query:"email"`
	Status *string `query:"status"`
	Limit  int32   `query:"limit" validate:"min=1,max=100"`
	Offset int32   `query:"offset" validate:"min=0"`
}

// ListResultDTO contains paginated user results for API responses.
type ListResultDTO struct {
	Users  []*UserDTO `json:"users"`
	Total  int64      `json:"total"`
	Limit  int32      `json:"limit"`
	Offset int32      `json:"offset"`
}
