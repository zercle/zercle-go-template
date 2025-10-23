package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// LoggerMiddleware logs incoming requests with structured logging using zerolog
func LoggerMiddleware(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Get request ID from context (set by RequestID middleware)
		requestID := getRequestID(c)

		// Log request details
		duration := time.Since(start)

		logEvent := log.Info()
		if err != nil {
			logEvent = log.Error().Err(err)
		}

		logEvent.
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Msg("HTTP Request")

		return err
	}
}
