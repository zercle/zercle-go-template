package dto

import "github.com/google/uuid"

// RegisterRequest represents the user registration request body.
// swagger:model
type RegisterRequest struct {
	// The username for the new user
	// Required: true
	// Min Length: 3
	// Max Length: 50
	Username string `json:"username" validate:"required,min=3,max=50"`
	// The email address for the new user
	// Required: true
	// Format: email
	Email string `json:"email" validate:"required,email"`
	// The password for the new user
	// Required: true
	// Min Length: 8
	Password string `json:"password" validate:"required,min=8"`
	// The display name for the new user (optional)
	DisplayName string `json:"display_name"`
}

// LoginRequest represents the login request body.
// swagger:model
type LoginRequest struct {
	// The user's email address
	// Required: true
	// Format: email
	Email string `json:"email" validate:"required,email"`
	// The user's password
	// Required: true
	Password string `json:"password" validate:"required"`
}

// RefreshRequest represents the token refresh request body.
// swagger:model
type RefreshRequest struct {
	// The refresh token
	// Required: true
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents the authentication response.
// swagger:model
type AuthResponse struct {
	// The JWT access token
	AccessToken string `json:"access_token"`
	// The refresh token for obtaining new access tokens
	RefreshToken string `json:"refresh_token"`
	// The authenticated user's information
	User *UserDTO `json:"user"`
	// Unix timestamp when the access token expires
	// Format: int64
	ExpiresAt int64 `json:"expires_at"`
}

// UserDTO represents user information.
// swagger:model
type UserDTO struct {
	// The unique user identifier
	// Format: uuid
	ID uuid.UUID `json:"id"`
	// The username
	Username string `json:"username"`
	// The email address
	// Format: email
	Email string `json:"email"`
	// The display name
	DisplayName string `json:"display_name"`
	// URL to the user's avatar image
	// Format: uri
	AvatarURL string `json:"avatar_url"`
	// The user's status (e.g., online, offline, away)
	Status string `json:"status"`
}
