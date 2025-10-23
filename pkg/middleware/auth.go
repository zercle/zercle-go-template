package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/pkg/response"
)

const userContextKey = "user_id"

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

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

			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
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

func GetUserID(c echo.Context) string {
	if userID, ok := c.Get(userContextKey).(string); ok {
		return userID
	}
	return ""
}

func GenerateToken(userID, email string, cfg *config.JWTConfig) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
