package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const requestIDKey = "request_id"

const headerXRequestID = "X-Request-ID"

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

func GetRequestID(c echo.Context) string {
	if reqID, ok := c.Get(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}
