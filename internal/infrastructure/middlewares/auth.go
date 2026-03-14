package middlewares

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// AuthMiddleware creates JWT authentication middleware.
func AuthMiddleware(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(401, "invalid authorization header")
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(401, "invalid token")
				}
				return secret, nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(401, "invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(401, "invalid token")
			}

			userIDStr, ok := claims["sub"].(string)
			if !ok {
				return echo.NewHTTPError(401, "invalid token")
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return echo.NewHTTPError(401, "invalid token")
			}

			c.Set("user_id", userID)

			return next(c)
		}
	}
}

// GetUserID extracts user ID from context.
func GetUserID(c *echo.Context) (uuid.UUID, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return uuid.Nil, echo.NewHTTPError(401, "unauthorized")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(401, "invalid token")
	}

	return uid, nil
}
