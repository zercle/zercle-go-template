// Package usecase provides JWT authentication business logic for the auth feature.
package usecase

//go:generate mockgen -source=$GOFILE -destination=./mocks/$GOFILE -package=mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/feature/auth/domain"
	"zercle-go-template/internal/logger"
)

// JWTUsecase defines the interface for JWT token operations.
type JWTUsecase interface {
	// GenerateTokenPair generates a new access and refresh token pair for a user.
	GenerateTokenPair(user domain.TokenUser) (*domain.TokenPair, error)
	// ValidateToken validates a JWT token and returns the claims.
	ValidateToken(tokenString string) (*domain.JWTClaims, error)
	// GenerateAccessToken generates a new access token from claims.
	GenerateAccessToken(claims *domain.JWTClaims) (string, error)
}

// jwtUsecase implements JWTUsecase.
type jwtUsecase struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logger          logger.Logger
}

// NewJWTUsecase creates a new JWT usecase instance.
func NewJWTUsecase(cfg *config.JWTConfig, log logger.Logger) JWTUsecase {
	return &jwtUsecase{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		logger:          log,
	}
}

// GenerateTokenPair generates a new access and refresh token pair for a user.
func (s *jwtUsecase) GenerateTokenPair(user domain.TokenUser) (*domain.TokenPair, error) {
	now := time.Now()

	// Create access token claims
	accessClaims := &domain.JWTClaims{
		UserID: user.GetID(),
		Email:  user.GetEmail(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "zercle-go-template",
			Subject:   user.GetID(),
			ID:        uuid.New().String(),
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to sign access token", logger.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Create refresh token claims
	refreshClaims := &domain.JWTClaims{
		UserID: user.GetID(),
		Email:  user.GetEmail(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "zercle-go-template",
			Subject:   user.GetID(),
			ID:        uuid.New().String(),
		},
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to sign refresh token", logger.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.logger.Info("token pair generated successfully",
		logger.String("user_id", user.GetID()),
		logger.Time("expires_at", accessClaims.ExpiresAt.Time),
	)

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessClaims.ExpiresAt.Time,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (s *jwtUsecase) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		s.logger.Warn("token validation failed", logger.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		s.logger.Warn("invalid token claims")
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateAccessToken generates a new access token from claims.
func (s *jwtUsecase) GenerateAccessToken(claims *domain.JWTClaims) (string, error) {
	now := time.Now()

	// Update expiration time
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(s.accessTokenTTL))
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)
	claims.ID = uuid.New().String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to sign access token", logger.Error(err))
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return tokenString, nil
}

// WithUserContext adds user information to the context.
func WithUserContext(ctx context.Context, claims *domain.JWTClaims) context.Context {
	ctx = context.WithValue(ctx, domain.ContextKeyUserID, claims.UserID)
	ctx = context.WithValue(ctx, domain.ContextKeyEmail, claims.Email)
	ctx = context.WithValue(ctx, domain.ContextKeyClaims, claims)
	return ctx
}

// GetUserIDFromContext retrieves the user ID from the context.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(domain.ContextKeyUserID).(string)
	return userID, ok
}

// GetEmailFromContext retrieves the email from the context.
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(domain.ContextKeyEmail).(string)
	return email, ok
}

// GetClaimsFromContext retrieves the JWT claims from the context.
func GetClaimsFromContext(ctx context.Context) (*domain.JWTClaims, bool) {
	claims, ok := ctx.Value(domain.ContextKeyClaims).(*domain.JWTClaims)
	return claims, ok
}
