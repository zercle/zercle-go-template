package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

func TestGenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-123",
		Expiration: 3600,
	}

	idValue := "test-user-id"
	email := "test@example.com"

	token, err := GenerateToken(idValue, email, cfg)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGetUserID(t *testing.T) {
	ec := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := ec.NewContext(req, rec)

	idResult := GetUserID(c)

	assert.Empty(t, idResult)
}

func TestGetUserID_WithValidContext(t *testing.T) {
	ec := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := ec.NewContext(req, rec)

	c.Set(userContextKey, "test-user-123")

	idResult := GetUserID(c)

	assert.Equal(t, "test-user-123", idResult)
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-123",
		Expiration: 3600,
	}

	middleware := JWTAuth(cfg)
	ec := echo.New()
	ec.Use(middleware)
	ec.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	ec.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing authorization header")
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-123",
		Expiration: 3600,
	}

	middleware := JWTAuth(cfg)
	ec := echo.New()
	ec.Use(middleware)
	ec.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "InvalidFormat token123")
	rec := httptest.NewRecorder()

	ec.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid authorization format")
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-123",
		Expiration: 3600,
	}

	middleware := JWTAuth(cfg)
	ec := echo.New()
	ec.Use(middleware)
	ec.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token-xyz")
	rec := httptest.NewRecorder()

	ec.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid or expired token")
}

func TestJWTAuth_ValidToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-123",
		Expiration: 3600,
	}

	token, _ := GenerateToken("user-123", "test@example.com", cfg)

	middleware := JWTAuth(cfg)
	ec := echo.New()
	ec.Use(middleware)

	ec.GET("/", func(c echo.Context) error {
		idValue := GetUserID(c)
		assert.Equal(t, "user-123", idValue)
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	ec.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	cfg1 := &config.JWTConfig{
		Secret:     "secret-key-1",
		Expiration: 3600,
	}
	cfg2 := &config.JWTConfig{
		Secret:     "secret-key-2",
		Expiration: 3600,
	}

	token, _ := GenerateToken("user-123", "test@example.com", cfg1)

	middleware := JWTAuth(cfg2)
	ec := echo.New()
	ec.Use(middleware)
	ec.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	ec.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
