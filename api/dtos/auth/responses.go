package auth

import "time"

// UserResponse represents user data in API responses.
type UserResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Response represents an authentication response with tokens.
type Response struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         *UserResponse `json:"user,omitempty"`
	ExpiresAt    int64         `json:"expires_at"`
}

// RefreshResponse represents a token refresh response.
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}
