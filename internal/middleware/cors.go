package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// CORSMiddleware handles CORS requests
func CORSMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		allowed := isOriginAllowed(origin, cfg.CORS.AllowedOrigins)

		if allowed {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Methods", strings.Join(cfg.CORS.AllowedMethods, ", "))
			c.Set("Access-Control-Allow-Headers", strings.Join(cfg.CORS.AllowedHeaders, ", "))

			if cfg.CORS.AllowedCredentials {
				c.Set("Access-Control-Allow-Credentials", "true")
			}

			c.Set("Access-Control-Max-Age", "3600")
		}

		// Handle preflight requests
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// isOriginAllowed checks if origin is in allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}

	return false
}
