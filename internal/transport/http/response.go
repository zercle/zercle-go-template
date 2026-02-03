package http

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Response is a standardized HTTP response envelope.
type Response struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// JSONSuccess returns a successful JSON response.
func JSONSuccess(c *echo.Context, data any) error {
	return (*c).JSON(http.StatusOK, Response{Status: "success", Data: data})
}

// JSONCreated returns a 201 Created JSON response.
func JSONCreated(c *echo.Context, data any) error {
	return (*c).JSON(http.StatusCreated, Response{Status: "success", Data: data})
}

// JSONNoContent returns a 204 No Content response.
func JSONNoContent(c *echo.Context) error {
	return (*c).NoContent(http.StatusNoContent)
}

// JSONFail returns a failed JSON response with the given code.
func JSONFail(c *echo.Context, code int, data any) error {
	return (*c).JSON(code, Response{Status: "fail", Data: data})
}

// JSONError returns an error JSON response.
func JSONError(c *echo.Context, code int, message string) error {
	return (*c).JSON(code, Response{Status: "error", Message: message, Code: code})
}

// JSONBadRequest returns a 400 Bad Request JSON response.
func JSONBadRequest(c *echo.Context, message string, err error) error {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return JSONFail(c, http.StatusBadRequest, map[string]string{"message": message, "error": errMsg})
}

// JSONValidationError returns a 400 validation error JSON response.
func JSONValidationError(c *echo.Context, err error) error {
	return JSONFail(c, http.StatusBadRequest, map[string]any{"validation": map[string]string{"error": err.Error()}})
}

// JSONNotFound returns a 404 Not Found JSON response.
func JSONNotFound(c *echo.Context, resource string) error {
	return JSONFail(c, http.StatusNotFound, map[string]string{"message": resource + " not found"})
}

// JSONUnauthorized returns a 401 Unauthorized JSON response.
func JSONUnauthorized(c *echo.Context, message string) error {
	return JSONFail(c, http.StatusUnauthorized, map[string]string{"message": message})
}

// JSONForbidden returns a 403 Forbidden JSON response.
func JSONForbidden(c *echo.Context, message string) error {
	return JSONFail(c, http.StatusForbidden, map[string]string{"message": message})
}

// JSONConflict returns a 409 Conflict JSON response.
func JSONConflict(c *echo.Context, message string) error {
	return JSONFail(c, http.StatusConflict, map[string]string{"message": message})
}

// JSONInternalError returns a 500 Internal Server Error JSON response.
func JSONInternalError(c *echo.Context, message string) error {
	return JSONError(c, http.StatusInternalServerError, message)
}
