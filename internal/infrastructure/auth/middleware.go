package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// ContextKeyUserClaims is the context key for storing user claims.
const ContextKeyUserClaims = "user_claims"

// JWTMiddleware returns middleware that validates JWT access tokens.
func JWTMiddleware(tokenService TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			tokenString := extractToken(c.Request())
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Authorization header required"})
			}

			claims, err := tokenService.ValidateAccessTokenSimple(c.Request().Context(), tokenString)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token expired"})
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid or expired token"})
			}

			c.Set(ContextKeyUserClaims, claims)

			return next(c)
		}
	}
}

// OptionalJWTMiddleware returns middleware that validates JWT tokens if present.
func OptionalJWTMiddleware(tokenService TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			tokenString := extractToken(c.Request())
			if tokenString != "" {
				claims, err := tokenService.ValidateAccessTokenSimple(c.Request().Context(), tokenString)
				if err == nil {
					c.Set(ContextKeyUserClaims, claims)
				}
			}
			return next(c)
		}
	}
}

// GetUserClaims retrieves user claims from the Echo context.
func GetUserClaims(c *echo.Context) (TokenClaims, bool) {
	claimsVal := c.Get(ContextKeyUserClaims)
	if claimsVal == nil {
		return TokenClaims{}, false
	}
	claims, ok := claimsVal.(TokenClaims)
	if !ok {
		return TokenClaims{}, false
	}
	return claims, true
}

// GetUserIDFromContext retrieves the user ID from the Echo context.
func GetUserIDFromContext(c *echo.Context) string {
	claims, ok := GetUserClaims(c)
	if !ok {
		return ""
	}
	return claims.UserID
}

// GetUserEmailFromContext retrieves the user email from the Echo context.
func GetUserEmailFromContext(c *echo.Context) string {
	claims, ok := GetUserClaims(c)
	if !ok {
		return ""
	}
	return claims.Email
}

// IsAuthenticated checks if the request context contains valid user claims.
func IsAuthenticated(c *echo.Context) bool {
	_, ok := GetUserClaims(c)
	return ok
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
