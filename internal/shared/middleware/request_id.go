// Echo middleware for request identification.
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// requestIDHeader is the header used to propagate or generate a request id.
const requestIDHeader = "X-Request-ID"

// maxRequestIDLen caps the length of an accepted client-supplied request id
// to prevent log/header injection and DoS via huge values.
const maxRequestIDLen = 128

// contextKey is the internal echo-context key for the request id.
type contextKey string

const requestIDKey contextKey = "request_id"

// isValidRequestID reports whether id is acceptable as a client-supplied
// request id. The allowed charset is URL-safe base64 / UUID-like characters
// ([A-Za-z0-9_-]) with a length cap of maxRequestIDLen.
func isValidRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLen {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

// RequestID returns echo middleware that reads or generates an X-Request-ID
// header, stores it in the echo context, and echoes it back in the response.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			id := req.Header.Get(requestIDHeader)
			if !isValidRequestID(id) {
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
