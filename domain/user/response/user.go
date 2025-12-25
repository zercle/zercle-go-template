package response

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse represents a user response
type UserResponse struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone,omitempty"`
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// ListUsersResponse represents a list of users response
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}
