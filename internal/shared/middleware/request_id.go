// Echo middleware for request identification.
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// requestIDHeader is the header used to propagate or generate a request id.
const requestIDHeader = "X-Request-ID"

// contextKey is the internal echo-context key for the request id.
type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID returns echo middleware that reads or generates an X-Request-ID
// header, stores it in the echo context, and echoes it back in the response.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			id := req.Header.Get(requestIDHeader)
			if id == "" {
				id = uuid.NewString()
			}

			c.Set(string(requestIDKey), id)
			c.Response().Header().Set(requestIDHeader, id)

			return next(c)
		}
	}
}

// RequestIDFromContext extracts the request id added by RequestID middleware.
func RequestIDFromContext(c *echo.Context) string {
	if id, ok := c.Get(string(requestIDKey)).(string); ok {
		return id
	}
	return ""
}
