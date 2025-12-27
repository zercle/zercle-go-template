package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Status represents JSend status types
type Status string

const (
	// StatusSuccess indicates successful request
	StatusSuccess Status = "success"
	// StatusFail indicates request with invalid data
	StatusFail Status = "fail"
	// StatusError indicates internal error
	StatusError Status = "error"
)

// JSend represents standardized API response format
type JSend struct {
	Status  Status       `json:"status"`
	Data    interface{}  `json:"data,omitempty"`
	Message string       `json:"message,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// FieldError represents a single field validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Success sends a successful response with data
func Success(c echo.Context, code int, data interface{}) error {
	return c.JSON(code, JSend{
		Status: StatusSuccess,
		Data:   data,
	})
}

// Fail sends a fail response with validation errors
func Fail(c echo.Context, code int, message string, errors []FieldError) error {
	return c.JSON(code, JSend{
		Status:  StatusFail,
		Message: message,
		Errors:  errors,
	})
}

// Error sends an error response
func Error(c echo.Context, code int, message string) error {
	return c.JSON(code, JSend{
		Status:  StatusError,
		Message: message,
	})
}

// Created sends a 201 Created response
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, JSend{
		Status: StatusSuccess,
		Data:   data,
	})
}

// OK sends a 200 OK response
func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, JSend{
		Status: StatusSuccess,
		Data:   data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c echo.Context, message string, errors []FieldError) error {
	return c.JSON(http.StatusBadRequest, JSend{
		Status:  StatusFail,
		Message: message,
		Errors:  errors,
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, JSend{
		Status:  StatusError,
		Message: message,
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, JSend{
		Status:  StatusError,
		Message: message,
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, JSend{
		Status:  StatusError,
		Message: message,
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, JSend{
		Status:  StatusError,
		Message: message,
	})
}
