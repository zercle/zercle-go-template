package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Logger creates an Echo middleware that logs HTTP request details including method, path, status, latency, and errors.
// Each log entry includes the request ID and contextual fields for structured logging.
func Logger(l *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			reqID := GetRequestID(c)

			err := next(c)

			latency := time.Since(start)

			fields := []interface{}{
				"request_id", reqID,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"query", c.Request().URL.RawQuery,
				"status", c.Response().Status,
				"ip", c.RealIP(),
				"user_agent", c.Request().UserAgent(),
				"latency", latency,
			}

			if err != nil {
				l.Error("HTTP request completed with error", fields...)
			} else {
				l.Info("HTTP request completed", fields...)
			}

			return err
		}
	}
}
