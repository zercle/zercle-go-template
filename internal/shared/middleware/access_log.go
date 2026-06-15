// Echo access logging middleware.
package middleware

import (
	"errors"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// AccessLog returns echo middleware that emits one structured log line per
// HTTP request with method, path, status, latency, and request id.
func AccessLog(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			err := next(c)

			status := responseStatus(c, err)

			logger.Info().
				Str("request_id", RequestIDFromContext(c)).
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status", status).
				Dur("latency", time.Since(start)).
				Msg("http request")

			return err
		}
	}
}

// responseStatus returns the HTTP status for the current echo context. It
// prefers an explicit echo.HTTPError from the handler chain and falls back to
// the response status already recorded on the echo Response.
func responseStatus(c *echo.Context, err error) int {
	if err != nil {
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return httpErr.Code
		}
	}

	if resp, ok := c.Response().(*echo.Response); ok {
		return resp.Status
	}

	return 0
}
