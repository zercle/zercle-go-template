// Echo middleware for request identification, built on echo's RequestID
// middleware with a UUID v7 generator.
package middleware

import (
	"uuid"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

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
// A client-supplied id that fails validation is replaced with a freshly
// generated UUID v7.
func RequestID() echo.MiddlewareFunc {
	return echomw.RequestIDWithConfig(echomw.RequestIDConfig{
		Generator: func() string { return uuid.NewV7().String() },
		RequestIDHandler: func(c *echo.Context, id string) {
			if !isValidRequestID(id) {
				id = uuid.NewV7().String()
				c.Response().Header().Set(echo.HeaderXRequestID, id)
			}
			c.Set(string(requestIDKey), id)
		},
	})
}

// RequestIDFromContext extracts the request id added by RequestID middleware.
func RequestIDFromContext(c *echo.Context) string {
	if id, ok := c.Get(string(requestIDKey)).(string); ok {
		return id
	}
	return ""
}
