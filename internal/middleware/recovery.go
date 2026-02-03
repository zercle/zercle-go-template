package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v5"

	appErr "zercle-go-template/internal/errors"
	"zercle-go-template/internal/logger"
)

// Recovery returns a middleware that recovers from panics.
// It logs the panic details and returns a 500 Internal Server Error response.
func Recovery(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Log the panic with stack trace
					log.Error("panic recovered",
						logger.String("error", fmt.Sprintf("%v", r)),
						logger.String("stack", string(debug.Stack())),
						logger.String("path", c.Path()),
						logger.String("method", c.Request().Method),
					)

					// Return error response
					c.Response().WriteHeader(http.StatusInternalServerError)
					_ = c.JSON(http.StatusInternalServerError, map[string]any{
						"success": false,
						"error": map[string]string{
							"code":    string(appErr.ErrCodeInternal),
							"message": "An unexpected error occurred",
						},
					})
				}
			}()

			return next(c)
		}
	}
}

// ErrorHandler returns a middleware that handles application errors.
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)

			// Check if there are any errors
			if err != nil {
				status := appErr.GetStatusCode(err)
				var errorInfo map[string]string

				if appError, ok := err.(*appErr.AppError); ok {
					errorInfo = map[string]string{
						"code":    string(appError.Code),
						"message": appError.Message,
						"details": appError.Details,
					}
				} else {
					errorInfo = map[string]string{
						"code":    string(appErr.ErrCodeInternal),
						"message": "An unexpected error occurred",
						"details": err.Error(),
					}
				}

				return c.JSON(status, map[string]any{
					"success": false,
					"error":   errorInfo,
				})
			}

			return nil
		}
	}
}

// NotFoundHandler handles requests to undefined routes.
func NotFoundHandler() echo.HandlerFunc {
	return func(c *echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error": map[string]string{
				"code":    string(appErr.ErrCodeNotFound),
				"message": "Resource not found",
			},
		})
	}
}

// MethodNotAllowedHandler handles requests with unsupported HTTP methods.
func MethodNotAllowedHandler() echo.HandlerFunc {
	return func(c *echo.Context) error {
		return c.JSON(http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error": map[string]string{
				"code":    string(appErr.ErrCodeBadRequest),
				"message": "Method not allowed",
			},
		})
	}
}
