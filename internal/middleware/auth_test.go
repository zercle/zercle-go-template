package middleware_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/middleware"
)

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "testsecret",
			Expiration: time.Hour,
		},
	}

	app := fiber.New()
	app.Use(middleware.AuthMiddleware(cfg))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("protected")
	})

	t.Run("Missing Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("Invalid Header Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("Valid Token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.JWTClaims{
			UserID: "user123",
			Email:  "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})
		tokenString, _ := token.SignedString([]byte(cfg.JWT.Secret))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.JWTClaims{
			UserID: "user123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			},
		})
		tokenString, _ := token.SignedString([]byte(cfg.JWT.Secret))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		}
	})
}
