package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/pkg/config"
)

func createTestTokenService() TokenService {
	cfg := config.AuthConfig{
		AccessTokenSecret:  "test-access-token-secret-key-32bytes",
		RefreshTokenSecret: "test-refresh-token-secret-key-32bytes",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	return NewTokenService(cfg)
}

func createExpiredTokenService() TokenService {
	cfg := config.AuthConfig{
		AccessTokenSecret:  "test-access-token-secret-key-32bytes",
		RefreshTokenSecret: "test-refresh-token-secret-key-32bytes",
		AccessTokenTTL:     -1 * time.Second, // Already expired
		RefreshTokenTTL:    -1 * time.Second,
		Issuer:             "test-issuer",
	}
	return NewTokenService(cfg)
}

func TestJWTService_GenerateTokenPair_Success(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.False(t, tokenPair.ExpiresAt.IsZero())
	assert.True(t, tokenPair.ExpiresAt.After(time.Now()))
}

func TestJWTService_GenerateTokenPair_DifferentUsersProduceDifferentTokens(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	tokenPair1, err1 := svc.GenerateTokenPair(ctx, "user-1", "user1@example.com")
	tokenPair2, err2 := svc.GenerateTokenPair(ctx, "user-2", "user2@example.com")

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, tokenPair1.AccessToken, tokenPair2.AccessToken)
	assert.NotEqual(t, tokenPair1.RefreshToken, tokenPair2.RefreshToken)
}

func TestJWTService_ValidateAccessToken_ValidToken(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(ctx, tokenPair.AccessToken)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestJWTService_ValidateAccessToken_ExpiredToken(t *testing.T) {
	svc := createExpiredTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(ctx, tokenPair.AccessToken)

	assert.Error(t, err)
	assert.Equal(t, ErrTokenExpired, err)
}

func TestJWTService_ValidateAccessToken_InvalidToken(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	_, err := svc.ValidateAccessToken(ctx, "invalid-token")

	assert.Error(t, err)
}

func TestJWTService_ValidateAccessToken_WrongSecret(t *testing.T) {
	// Create two services with different secrets
	cfg1 := config.AuthConfig{
		AccessTokenSecret:  "secret-key-one-32-characters!!!",
		RefreshTokenSecret: "refresh-secret-key-one-32!!!!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	cfg2 := config.AuthConfig{
		AccessTokenSecret:  "secret-key-two-32-characters!!!!",
		RefreshTokenSecret: "refresh-secret-key-two-32!!!!!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	svc1 := NewTokenService(cfg1)
	svc2 := NewTokenService(cfg2)
	ctx := context.Background()

	tokenPair, err := svc1.GenerateTokenPair(ctx, "user-123", "test@example.com")
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = svc2.ValidateAccessToken(ctx, tokenPair.AccessToken)

	assert.Error(t, err)
}

func TestJWTService_ValidateRefreshToken_ValidToken(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(ctx, tokenPair.RefreshToken)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestJWTService_ValidateRefreshToken_ExpiredToken(t *testing.T) {
	svc := createExpiredTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	_, err = svc.ValidateRefreshToken(ctx, tokenPair.RefreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrTokenExpired, err)
}

func TestJWTService_ValidateRefreshToken_InvalidToken(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	_, err := svc.ValidateRefreshToken(ctx, "invalid-token")

	assert.Error(t, err)
}

func TestJWTService_ValidateRefreshToken_WrongSecret(t *testing.T) {
	// Create two services with different secrets
	cfg1 := config.AuthConfig{
		AccessTokenSecret:  "secret-key-one-32-characters!!!",
		RefreshTokenSecret: "refresh-secret-key-one-32!!!!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	cfg2 := config.AuthConfig{
		AccessTokenSecret:  "secret-key-two-32-characters!!!!",
		RefreshTokenSecret: "refresh-secret-key-two-32!!!!!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	svc1 := NewTokenService(cfg1)
	svc2 := NewTokenService(cfg2)
	ctx := context.Background()

	tokenPair, err := svc1.GenerateTokenPair(ctx, "user-123", "test@example.com")
	require.NoError(t, err)

	// Try to validate refresh with different secret
	_, err = svc2.ValidateRefreshToken(ctx, tokenPair.RefreshToken)

	assert.Error(t, err)
}

func TestJWTService_ExtractUserID_ValidToken(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	extractedUserID, err := svc.ExtractUserID(ctx, tokenPair.AccessToken)

	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestJWTService_ExtractUserID_InvalidToken(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	_, err := svc.ExtractUserID(ctx, "invalid-token")

	assert.Error(t, err)
}

func TestJWTService_ExtractUserID_ExpiredToken(t *testing.T) {
	svc := createExpiredTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	_, err = svc.ExtractUserID(ctx, tokenPair.AccessToken)

	assert.Error(t, err)
}

func TestJWTService_TokenClaims_HasCorrectFields(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(ctx, tokenPair.AccessToken)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.NotEmpty(t, claims.ID) // JWT ID
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now().Add(time.Second)))
}

func TestJWTService_AccessAndRefreshTokensAreDifferent(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	assert.NotEqual(t, tokenPair.AccessToken, tokenPair.RefreshToken)

	// Validate access token with access secret
	accessClaims, err := svc.ValidateAccessToken(ctx, tokenPair.AccessToken)
	require.NoError(t, err)

	// Validate refresh token with refresh secret
	refreshClaims, err := svc.ValidateRefreshToken(ctx, tokenPair.RefreshToken)
	require.NoError(t, err)

	assert.Equal(t, accessClaims.UserID, refreshClaims.UserID)
	assert.Equal(t, accessClaims.Email, refreshClaims.Email)
}

func TestJWTService_RejectsTokenSignedWithRefreshSecret(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	// Try to validate refresh token with access secret (should fail)
	_, err = svc.ValidateAccessToken(ctx, tokenPair.RefreshToken)
	assert.Error(t, err)
}

func TestJWTService_RejectsTokenSignedWithAccessSecret(t *testing.T) {
	svc := createTestTokenService()
	ctx := context.Background()

	userID := "user-123"
	email := "test@example.com"

	tokenPair, err := svc.GenerateTokenPair(ctx, userID, email)
	require.NoError(t, err)

	// Try to validate access token with refresh secret (should fail)
	_, err = svc.ValidateRefreshToken(ctx, tokenPair.AccessToken)
	assert.Error(t, err)
}
