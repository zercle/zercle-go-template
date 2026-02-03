// Package auth provides JWT token management and password hashing services.
package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/config"
)

var (
	// ErrInvalidToken indicates the token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired indicates the token has expired.
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidSignature indicates the token signature is invalid.
	ErrInvalidSignature = errors.New("invalid token signature")
)

// TokenClaims represents the JWT claims structure.
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// TokenService provides JWT token operations.
type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error)
	ValidateAccessToken(ctx context.Context, tokenString string) (*TokenClaims, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (*TokenClaims, error)
	ExtractUserID(ctx context.Context, tokenString string) (string, error)
	// Methods for feature auth TokenService interface
	GenerateAccessToken(ctx context.Context, userID, email string) (string, error)
	GenerateRefreshToken(ctx context.Context, userID, email string) (string, error)
	ValidateAccessTokenSimple(ctx context.Context, token string) (TokenClaims, error)
	ValidateRefreshTokenSimple(ctx context.Context, token string) (TokenClaims, error)
}

// jwtService implements TokenService using golang-jwt/jwt/v5.
type jwtService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

// NewTokenService creates a new JWT token service.
func NewTokenService(cfg config.AuthConfig) TokenService {
	return &jwtService{
		accessSecret:  []byte(cfg.AccessTokenSecret),
		refreshSecret: []byte(cfg.RefreshTokenSecret),
		accessTTL:     cfg.AccessTokenTTL,
		refreshTTL:    cfg.RefreshTokenTTL,
		issuer:        cfg.Issuer,
	}
}

// GenerateTokenPair generates a new access and refresh token pair.
func (s *jwtService) GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessClaims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			ID:        uuid.New().String(),
		},
		UserID: userID,
		Email:  email,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			ID:        uuid.New().String(),
		},
		UserID: userID,
		Email:  email,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(s.accessTTL),
	}, nil
}

// GenerateAccessToken generates just an access token.
func (s *jwtService) GenerateAccessToken(ctx context.Context, userID, email string) (string, error) {
	pair, err := s.GenerateTokenPair(ctx, userID, email)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}

// GenerateRefreshToken generates just a refresh token.
func (s *jwtService) GenerateRefreshToken(ctx context.Context, userID, email string) (string, error) {
	pair, err := s.GenerateTokenPair(ctx, userID, email)
	if err != nil {
		return "", err
	}
	return pair.RefreshToken, nil
}

// ValidateAccessToken validates and parses an access token.
func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessTokenSimple validates and returns claims by value.
func (s *jwtService) ValidateAccessTokenSimple(ctx context.Context, token string) (TokenClaims, error) {
	claims, err := s.ValidateAccessToken(ctx, token)
	if err != nil {
		return TokenClaims{}, err
	}
	return *claims, nil
}

// ValidateRefreshToken validates and parses a refresh token.
func (s *jwtService) ValidateRefreshToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.refreshSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshTokenSimple validates and returns claims by value.
func (s *jwtService) ValidateRefreshTokenSimple(ctx context.Context, token string) (TokenClaims, error) {
	claims, err := s.ValidateRefreshToken(ctx, token)
	if err != nil {
		return TokenClaims{}, err
	}
	return *claims, nil
}

// ExtractUserID extracts the user ID from a token without full validation.
func (s *jwtService) ExtractUserID(ctx context.Context, tokenString string) (string, error) {
	claims, err := s.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// HashToken creates a secure hash of a token for database storage.
// Uses constant-time comparison to prevent timing attacks.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// CompareToken compares a token with its hash using constant-time comparison.
func CompareToken(token, hash string) bool {
	return subtle.ConstantTimeCompare([]byte(HashToken(token)), []byte(hash)) == 1
}
