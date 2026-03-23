// Package response provides JSend-compliant response helpers for HTTP handlers.
package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Status represents the JSend response status.
type Status string

// JSend status constants.
const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

// Response represents a JSend-formatted response.
type Response struct {
	Status  Status `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// Success sends a success response with status "success" and data.
func Success(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, Response{
		Status: StatusSuccess,
		Data:   data,
	})
}

// Created sends a 201 Created response with status "success" and data.
func Created(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Response{
		Status: StatusSuccess,
		Data:   data,
	})
}

// NoContent sends a 204 No Content response.
func NoContent(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Fail sends a fail response (client error) with status "fail" and data.
func Fail(c *echo.Context, statusCode int, data any) error {
	return c.JSON(statusCode, Response{
		Status: StatusFail,
		Data:   data,
	})
}

// BadRequest sends a 400 Bad Request response.
func BadRequest(c *echo.Context, message string, err error) error {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return Fail(c, http.StatusBadRequest, map[string]string{
		"message": message,
		"error":   errMsg,
	})
}

// ValidationError sends a 400 Bad Request response for validation errors.
func ValidationError(c *echo.Context, err error) error {
	return Fail(c, http.StatusBadRequest, map[string]any{
		"validation": parseValidationErrors(err),
	})
}

// NotFound sends a 404 Not Found response.
func NotFound(c *echo.Context, resource string) error {
	return Fail(c, http.StatusNotFound, map[string]string{
		"message": resource + " not found",
	})
}

// Unauthorized sends a 401 Unauthorized response.
func Unauthorized(c *echo.Context, message string) error {
	return Fail(c, http.StatusUnauthorized, map[string]string{
		"message": message,
	})
}

// Forbidden sends a 403 Forbidden response.
func Forbidden(c *echo.Context, message string) error {
	return Fail(c, http.StatusForbidden, map[string]string{
		"message": message,
	})
}

// Conflict sends a 409 Conflict response.
func Conflict(c *echo.Context, message string) error {
	return Fail(c, http.StatusConflict, map[string]string{
		"message": message,
	})
}

// Error sends an error response (server error) with status "error".
func Error(c *echo.Context, statusCode int, message string, code int) error {
	return c.JSON(statusCode, Response{
		Status:  StatusError,
		Message: message,
		Code:    code,
	})
}

// InternalError sends a 500 Internal Server Error response.
func InternalError(c *echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message, CodeInternalError)
}

// parseValidationErrors extracts validation errors from error types.
// This is a placeholder that can be extended based on the validation library used.
func parseValidationErrors(err error) map[string]string {
	// For now, return a simple map with the error message
	// In production, this would parse specific validation error types
	return map[string]string{
		"error": err.Error(),
	}
}
