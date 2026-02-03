package user

import (
	"time"
)

// CreateUserInput contains data for creating a new user.
type CreateUserInput struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// UpdateUserInput contains optional fields for updating a user.
type UpdateUserInput struct {
	Email     *string `json:"email" validate:"omitempty,email"`
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	Status    *string `json:"status" validate:"omitempty,oneof=active inactive suspended"`
}

// Response represents a user in API responses.
type Response struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListQuery contains query parameters for listing users.
type ListQuery struct {
	Email  string  `query:"email"`
	Status *string `query:"status"`
	Limit  int32   `query:"limit" validate:"min=1,max=100"`
	Offset int32   `query:"offset" validate:"min=0"`
}

// ListResponse contains paginated user results for API responses.
type ListResponse struct {
	Users  []*Response `json:"users"`
	Total  int64       `json:"total"`
	Limit  int32       `json:"limit"`
	Offset int32       `json:"offset"`
}

// ListParams contains filtering and pagination options for listing users.
type ListParams struct {
	Email  string
	Status *string
	Limit  int32
	Offset int32
}

// ListResult contains paginated user results.
type ListResult struct {
	Users  []*User
	Total  int64
	Limit  int32
	Offset int32
}
