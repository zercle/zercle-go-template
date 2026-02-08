// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"zercle-go-template/internal/logger"
)

// RequestLogger returns a middleware that logs HTTP requests.
func RequestLogger(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			path := c.Path()
			query := c.QueryString()

			// Add logger to context
			ctx := logger.WithContext(c.Request().Context(), log)
			c.Set("logger", log)
			c.SetRequest(c.Request().WithContext(ctx))

			// Process request
			err := next(c)

			// Log request details
			duration := time.Since(start)
			clientIP := c.RealIP()
			method := c.Request().Method
			statusCode := c.Response().Status

			if len(query) > 0 {
				path = path + "?" + query
			}

			fields := []logger.Field{
				logger.String("client_ip", clientIP),
				logger.String("method", method),
				logger.String("path", path),
				logger.Int("status", statusCode),
				logger.Duration("duration", duration),
				logger.String("user_agent", c.Request().Header.Get("User-Agent")),
			}

			// Add error field if present
			if err != nil {
				fields = append(fields, logger.String("error", err.Error()))
			}

			// Log based on status code
			switch {
			case statusCode >= 500:
				log.Error("server error", fields...)
			case statusCode >= 400:
				log.Warn("client error", fields...)
			default:
				log.Info("request completed", fields...)
			}

			return err
		}
	}
}

// LoggerContext adds the logger to the request context.
func LoggerContext(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := logger.WithContext(c.Request().Context(), log)
			c.Set("logger", log)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
