package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	sharedHandler "github.com/zercle/zercle-go-template/internal/shared/handler/response"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return sharedHandler.Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": "missing authorization header"})
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return sharedHandler.Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": "invalid authorization header format"})
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})
		if err != nil {
			return sharedHandler.Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": "invalid or expired token"})
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			return sharedHandler.Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": "invalid token claims"})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens but doesn't fail if missing
func OptionalAuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})
		if err != nil {
			return c.Next()
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			return c.Next()
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}
