// Package middleware provides HTTP middleware for the auth feature.
package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	appErr "zercle-go-template/internal/errors"
	"zercle-go-template/internal/feature/auth/domain"
	"zercle-go-template/internal/feature/auth/usecase"
	"zercle-go-template/internal/logger"
)

const (
	// AuthorizationHeader is the HTTP header key for authorization.
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for Bearer token authentication.
	BearerPrefix = "Bearer "
)

// JWTAuth returns a middleware that validates JWT tokens.
func JWTAuth(jwtUsecase usecase.JWTUsecase, log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get(AuthorizationHeader)
			if authHeader == "" {
				log.Warn("missing authorization header",
					logger.String("path", c.Path()),
					logger.String("client_ip", c.RealIP()),
				)
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"success": false,
					"error": map[string]any{
						"code":    string(appErr.ErrCodeUnauthorized),
						"message": "Authorization header is required",
					},
				})
			}

			// Check for Bearer prefix
			if !strings.HasPrefix(authHeader, BearerPrefix) {
				log.Warn("invalid authorization header format",
					logger.String("path", c.Path()),
					logger.String("client_ip", c.RealIP()),
				)
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"success": false,
					"error": map[string]any{
						"code":    string(appErr.ErrCodeUnauthorized),
						"message": "Invalid authorization header format. Expected 'Bearer <token>'",
					},
				})
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
			if tokenString == "" {
				log.Warn("empty token",
					logger.String("path", c.Path()),
					logger.String("client_ip", c.RealIP()),
				)
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"success": false,
					"error": map[string]any{
						"code":    string(appErr.ErrCodeUnauthorized),
						"message": "Token is required",
					},
				})
			}

			// Validate token
			claims, err := jwtUsecase.ValidateToken(tokenString)
			if err != nil {
				log.Warn("token validation failed",
					logger.String("path", c.Path()),
					logger.String("client_ip", c.RealIP()),
					logger.Error(err),
				)
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"success": false,
					"error": map[string]any{
						"code":    string(appErr.ErrCodeUnauthorized),
						"message": "Invalid or expired token",
					},
				})
			}

			// Set user context
			ctx := usecase.WithUserContext(c.Request().Context(), claims)
			c.SetRequest(c.Request().WithContext(ctx))

			// Set user info in echo context for easy access
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("claims", claims)

			log.Debug("token validated successfully",
				logger.String("user_id", claims.UserID),
				logger.String("email", claims.Email),
				logger.String("path", c.Path()),
			)

			return next(c)
		}
	}
}

// GetUserID retrieves the user ID from the echo context.
// Returns empty string if not found.
func GetUserID(c *echo.Context) string {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetEmail retrieves the email from the echo context.
// Returns empty string if not found.
func GetEmail(c *echo.Context) string {
	if email, ok := c.Get("email").(string); ok {
		return email
	}
	return ""
}

// GetClaims retrieves the JWT claims from the echo context.
// Returns nil if not found.
func GetClaims(c *echo.Context) *domain.JWTClaims {
	if claims, ok := c.Get("claims").(*domain.JWTClaims); ok {
		return claims
	}
	return nil
}

// RequireAuth is a convenience function that combines JWTAuth middleware.
// It can be used directly in route definitions.
func RequireAuth(jwtUsecase usecase.JWTUsecase, log logger.Logger) echo.MiddlewareFunc {
	return JWTAuth(jwtUsecase, log)
}

// OptionalAuth returns a middleware that validates JWT tokens if present,
// but allows the request to continue even without a token.
// This is useful for endpoints that have different behavior for authenticated users.
func OptionalAuth(jwtUsecase usecase.JWTUsecase, log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get(AuthorizationHeader)
			if authHeader == "" {
				return next(c)
			}

			if !strings.HasPrefix(authHeader, BearerPrefix) {
				return next(c)
			}

			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
			if tokenString == "" {
				return next(c)
			}

			claims, err := jwtUsecase.ValidateToken(tokenString)
			if err != nil {
				// Token is invalid but we allow the request to continue
				log.Debug("optional auth: invalid token", logger.Error(err))
				return next(c)
			}

			// Set user context
			ctx := usecase.WithUserContext(c.Request().Context(), claims)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("claims", claims)

			return next(c)
		}
	}
}
