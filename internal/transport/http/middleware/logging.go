// Package middleware provides Echo middleware implementations.
package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"

	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
)

// ZerologMiddleware provides structured HTTP logging using zerolog.
func ZerologMiddleware(logger *logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			// Add logger to context
			ctx := logger.ToContext(c.Request().Context())
			c.SetRequest(c.Request().WithContext(ctx))

			err := next(c)

			// Build log event based on status code
			var event *zerolog.Event
			status := getStatusCode(c.Response())
			if status >= 500 {
				event = logger.Error()
			} else if status >= 400 {
				event = logger.Warn()
			} else {
				event = logger.Info()
			}

			event.
				Str("method", c.Request().Method).
				Str("path", c.Path()).
				Str("url", c.Request().URL.String()).
				Int("status", status).
				Dur("latency_ms", time.Since(start)).
				Str("remote_ip", c.RealIP()).
				Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
				Str("user_agent", c.Request().UserAgent()).
				Msg("HTTP request")

			return err
		}
	}
}

// getStatusCode extracts the status code from the response writer.
// Echo v5 uses a custom response writer that stores the status code internally.
func getStatusCode(rw http.ResponseWriter) int {
	if wr, ok := rw.(interface{ Status() int }); ok {
		return wr.Status()
	}
	return http.StatusOK
}
