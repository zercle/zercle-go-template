// Package domain contains JWT-related domain types for the auth feature.
package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT token claims.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GetAccessToken returns the access token for interface compatibility.
func (tp *TokenPair) GetAccessToken() string {
	return tp.AccessToken
}

// ContextKey is the key type for storing values in context.
type ContextKey string

// Context keys for authentication.
const (
	ContextKeyUserID ContextKey = "user_id"
	ContextKeyEmail  ContextKey = "email"
	ContextKeyClaims ContextKey = "claims"
)

// TokenUser represents a user that can be used to generate JWT tokens.
// This interface is typically satisfied by user domain objects.
type TokenUser interface {
	GetID() string
	GetEmail() string
}
