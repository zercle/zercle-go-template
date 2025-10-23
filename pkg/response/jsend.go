package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

type JSend struct {
	Status  Status       `json:"status"`
	Data    any          `json:"data,omitempty"`
	Message string       `json:"message,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Success(c echo.Context, code int, data any) error {
	return c.JSON(code, JSend{
		Status: StatusSuccess,
		Data:   data,
	})
}

func Fail(c echo.Context, code int, message string, errors []FieldError) error {
	return c.JSON(code, JSend{
		Status:  StatusFail,
		Message: message,
		Errors:  errors,
	})
}

func Error(c echo.Context, code int, message string) error {
	return c.JSON(code, JSend{
		Status:  StatusError,
		Message: message,
	})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, JSend{
		Status: StatusSuccess,
		Data:   data,
	})
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, JSend{
		Status: StatusSuccess,
		Data:   data,
	})
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func BadRequest(c echo.Context, message string, errors []FieldError) error {
	return c.JSON(http.StatusBadRequest, JSend{
		Status:  StatusFail,
		Message: message,
		Errors:  errors,
	})
}

func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, JSend{
		Status:  StatusError,
		Message: message,
	})
}

func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, JSend{
		Status:  StatusError,
		Message: message,
	})
}

func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, JSend{
		Status:  StatusError,
		Message: message,
	})
}

func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, JSend{
		Status:  StatusError,
		Message: message,
	})
}
