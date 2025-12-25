package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/pkg/response"
)

const userContextKey = "user_id"

// JWTClaims represents JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTAuth creates a JWT authentication middleware
func JWTAuth(cfg *config.JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Unauthorized(c, "Missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return response.Unauthorized(c, "Invalid authorization format")
			}

			tokenString := parts[1]

			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Secret), nil
			})

			if err != nil || !token.Valid {
				return response.Unauthorized(c, "Invalid or expired token")
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				return response.Unauthorized(c, "Invalid token claims")
			}

			c.Set(userContextKey, claims.UserID)

			return next(c)
		}
	}
}

// GetUserID retrieves the authenticated user ID from the Echo context.
// Returns an empty string if no user ID is present or authentication has not occurred.
func GetUserID(c echo.Context) string {
	if userID, ok := c.Get(userContextKey).(string); ok {
		return userID
	}
	return ""
}

// GenerateToken creates a signed JWT token for the specified user credentials.
// The token contains the user ID and email in the claims.
func GenerateToken(userID, email string, cfg *config.JWTConfig) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
