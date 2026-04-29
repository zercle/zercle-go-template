package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// JWTClaims represents the claims stored in a JWT token for authenticated users.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}

// AuthMiddleware returns an Echo middleware that validates JWT tokens from the Authorization header.
func AuthMiddleware(jwtSecret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
			}

			tokenString := parts[1]

			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, apperrors.ErrTokenInvalid
				}
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)

			return next(c)
		}
	}
}

// GetUserID extracts the user ID from the Echo context, returning an error if not set.
func GetUserID(c *echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, apperrors.ErrUnauthorized
	}
	return userID, nil
}

// GetUsername extracts the username from the Echo context.
func GetUsername(c *echo.Context) string {
	username, _ := c.Get("username").(string)
	return username
}
