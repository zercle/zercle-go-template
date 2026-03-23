package user

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// ListParams contains filtering and pagination options for listing users.
type ListParams struct {
	// Email filters by email address (partial match supported by implementation).
	Email string
	// Status filters by user status.
	Status *UserStatus
	// Limit is the maximum number of results to return.
	Limit int32
	// Offset is the number of results to skip.
	Offset int32
}

// ListResult contains paginated user results.
type ListResult struct {
	// Users is the list of users matching the query.
	Users []*User
	// Total is the total count of matching records (before pagination).
	Total int64
	// Limit is the limit used for this query.
	Limit int32
	// Offset is the offset used for this query.
	Offset int32
}

// Repository defines the contract for user data persistence.
// This interface follows the repository pattern for clean architecture,
// allowing the domain layer to define data access contracts that
// infrastructure components implement.
type Repository interface {
	// Create inserts a new user into the data store.
	// Returns the created user with generated ID and timestamps.
	Create(ctx context.Context, user *User) (*User, error)

	// GetByID retrieves a user by their unique identifier.
	// Returns ErrUserNotFound if no user exists with the given ID.
	GetByID(ctx context.Context, id UserID) (*User, error)

	// GetByEmail retrieves a user by their email address.
	// Returns ErrUserNotFound if no user exists with the given email.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update modifies an existing user in the data store.
	// Returns the updated user.
	// Returns ErrUserNotFound if no user exists with the given ID.
	Update(ctx context.Context, user *User) (*User, error)

	// Delete removes a user from the data store.
	// Returns ErrUserNotFound if no user exists with the given ID.
	Delete(ctx context.Context, id UserID) error

	// List returns a paginated list of users matching the given parameters.
	List(ctx context.Context, params *ListParams) (*ListResult, error)

	// Exists checks if a user with the given email exists.
	Exists(ctx context.Context, email string) (bool, error)

	// ExistsByID checks if a user with the given ID exists.
	ExistsByID(ctx context.Context, id UserID) (bool, error)
}
