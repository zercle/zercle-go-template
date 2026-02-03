package middleware

import (
	"github.com/labstack/echo/v5"
)

// ZerologMiddleware returns an Echo middleware for Zerolog logging.
func ZerologMiddleware(_ any) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return next(c)
		}
	}
}
