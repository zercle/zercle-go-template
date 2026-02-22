// Package usecase provides JWT authentication business logic for the auth feature.
package usecase

//go:generate mockgen -source=$GOFILE -destination=./mocks/$GOFILE -package=mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/feature/auth/domain"
	"zercle-go-template/internal/logger"
)

// TokenCache stores validated JWT tokens with TTL-based expiration.
// It uses sync.Map for thread-safe concurrent access.
// It runs a background cleanup goroutine to periodically remove expired entries.
type TokenCache struct {
	tokens    sync.Map // map[string]*cachedToken
	ctx       context.Context
	cancel    context.CancelFunc
	ticker    *time.Ticker
}

// cachedToken represents a cached token with its claims and expiration time.
type cachedToken struct {
	claims    *domain.JWTClaims
	expiresAt time.Time
}

// Get retrieves a cached token if it exists and hasn't expired.
// Returns the claims and true if found and valid, otherwise nil and false.
func (c *TokenCache) Get(token string) (*domain.JWTClaims, bool) {
	if val, ok := c.tokens.Load(token); ok {
		ct := val.(*cachedToken)
		if time.Now().Before(ct.expiresAt) {
			return ct.claims, true
		}
		// Token expired, remove from cache
		c.tokens.Delete(token)
	}
	return nil, false
}

// Set stores a token in the cache with the specified TTL.
func (c *TokenCache) Set(token string, claims *domain.JWTClaims, ttl time.Duration) {
	c.tokens.Store(token, &cachedToken{
		claims:    claims,
		expiresAt: time.Now().Add(ttl),
	})
}

// Delete removes a token from the cache.
func (c *TokenCache) Delete(token string) {
	c.tokens.Delete(token)
}

// Size returns the approximate number of cached tokens.
func (c *TokenCache) Size() int {
	var count int
	c.tokens.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// Clear removes all tokens from the cache.
func (c *TokenCache) Clear() {
	c.tokens = sync.Map{}
}

// cleanupLoop periodically removes expired tokens from the cache.
// It runs every 5 minutes and stops when the context is cancelled.
func (c *TokenCache) cleanupLoop() {
	defer c.ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.ticker.C:
			c.removeExpiredTokens()
		}
	}
}

// removeExpiredTokens iterates through all cached tokens and removes expired ones.
func (c *TokenCache) removeExpiredTokens() {
	now := time.Now()
	var deletedCount int
	c.tokens.Range(func(key, value any) bool {
		token := key.(string)
		ct := value.(*cachedToken)
		if now.After(ct.expiresAt) {
			c.tokens.Delete(token)
			deletedCount++
		}
		return true
	})
	if deletedCount > 0 {
		// Optional: Log cleanup activity
	}
}

// Stop stops the background cleanup goroutine gracefully.
// It should be called when the TokenCache is no longer needed.
func (c *TokenCache) Stop() {
	c.cancel()
}

// NewTokenCache creates a new TokenCache with a background cleanup goroutine.
// The cleanup runs every 5 minutes to remove expired tokens.
func NewTokenCache() *TokenCache {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(5 * time.Minute)
	cache := &TokenCache{
		ctx:    ctx,
		cancel: cancel,
		ticker: ticker,
	}
	go cache.cleanupLoop()
	return cache
}

// jwtClaimsPool is a sync.Pool for reusing JWTClaims objects to reduce GC pressure.
// This reduces allocations during token generation by reusing claim structures.
var jwtClaimsPool = sync.Pool{
	New: func() any {
		return &domain.JWTClaims{}
	},
}

// getJWTClaims retrieves a JWTClaims object from the pool.
func getJWTClaims() *domain.JWTClaims {
	return jwtClaimsPool.Get().(*domain.JWTClaims)
}

// putJWTClaims returns a JWTClaims object to the pool after resetting its fields.
func putJWTClaims(c *domain.JWTClaims) {
	// Reset all fields to prevent data leakage between uses
	c.UserID = ""
	c.Email = ""
	c.RegisteredClaims = jwt.RegisteredClaims{}
	jwtClaimsPool.Put(c)
}

// JWTUsecase defines the interface for JWT token operations.
type JWTUsecase interface {
	// GenerateTokenPair generates a new access and refresh token pair for a user.
	GenerateTokenPair(user domain.TokenUser) (*domain.TokenPair, error)
	// ValidateToken validates a JWT token and returns the claims.
	ValidateToken(tokenString string) (*domain.JWTClaims, error)
	// GenerateAccessToken generates a new access token from claims.
	GenerateAccessToken(claims *domain.JWTClaims) (string, error)
	// Stop stops the token cache cleanup goroutine.
	Stop()
}

// jwtUsecase implements JWTUsecase.
type jwtUsecase struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logger          logger.Logger
	// cache is the token validation cache for performance optimization
	cache        *TokenCache
	cacheEnabled bool
	cacheTTL     time.Duration
}

// NewJWTUsecase creates a new JWT usecase instance.
func NewJWTUsecase(cfg *config.JWTConfig, log logger.Logger) JWTUsecase {
	return &jwtUsecase{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		logger:          log,
		cache:           NewTokenCache(),
		cacheEnabled:    cfg.CacheEnabled,
		cacheTTL:        cfg.CacheTTL,
	}
}

// GenerateTokenPair generates a new access and refresh token pair for a user.
func (s *jwtUsecase) GenerateTokenPair(user domain.TokenUser) (*domain.TokenPair, error) {
	now := time.Now()

	// Get pooled access token claims
	accessClaims := getJWTClaims()
	accessClaims.UserID = user.GetID()
	accessClaims.Email = user.GetEmail()
	accessClaims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "zercle-go-template",
		Subject:   user.GetID(),
		ID:        uuid.New().String(),
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to sign access token", logger.Error(err))
		putJWTClaims(accessClaims)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Get pooled refresh token claims
	refreshClaims := getJWTClaims()
	refreshClaims.UserID = user.GetID()
	refreshClaims.Email = user.GetEmail()
	refreshClaims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "zercle-go-template",
		Subject:   user.GetID(),
		ID:        uuid.New().String(),
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secret)
	if err != nil {
		s.logger.Error("failed to sign refresh token", logger.Error(err))
		putJWTClaims(accessClaims)
		putJWTClaims(refreshClaims)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := accessClaims.ExpiresAt.Time

	// Return claims to pool after successful token generation
	putJWTClaims(accessClaims)
	putJWTClaims(refreshClaims)

	s.logger.Info("token pair generated successfully",
		logger.String("user_id", user.GetID()),
		logger.Time("expires_at", expiresAt),
	)

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims.
// It checks the cache first if caching is enabled.
func (s *jwtUsecase) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	// Check cache first if enabled
	if s.cacheEnabled {
		if claims, ok := s.cache.Get(tokenString); ok {
			return claims, nil
		}
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (any, error) {
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

	// Cache the validated token if enabled
	if s.cacheEnabled {
		s.cache.Set(tokenString, claims, s.cacheTTL)
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

// Stop stops the token cache and its background cleanup goroutine.
func (s *jwtUsecase) Stop() {
	if s.cache != nil {
		s.cache.Stop()
	}
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
