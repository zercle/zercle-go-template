package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Logging returns an Echo middleware that logs HTTP request details (status, method, URI).
func Logging() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogMethod: true,
		LogURI:    true,
	})
}

// Recover returns an Echo middleware that recovers from panics and returns an internal server error.
func Recover() echo.MiddlewareFunc {
	return middleware.Recover()
}
