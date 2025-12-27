package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const requestIDKey = "request_id"

const headerXRequestID = "X-Request-ID"

// RequestID creates an Echo middleware that generates or propagates a unique request ID.
// If an X-Request-ID header is present in the request, it is reused; otherwise a new UUID is generated.
// The request ID is added to both the Echo context and the X-Request-ID response header.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(headerXRequestID)
			if reqID == "" {
				id, err := uuid.NewV7()
				if err != nil {
					c.Set(requestIDKey, "")
					c.Response().Header().Set(headerXRequestID, "")
					return next(c)
				}
				reqID = id.String()
			}

			c.Set(requestIDKey, reqID)
			c.Response().Header().Set(headerXRequestID, reqID)

			return next(c)
		}
	}
}

// GetRequestID retrieves the request ID from the Echo context.
// Returns an empty string if no request ID is present.
func GetRequestID(c echo.Context) string {
	if reqID, ok := c.Get(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}
