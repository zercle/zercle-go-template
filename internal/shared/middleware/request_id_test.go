//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

func TestRequestID_GeneratesWhenAbsent(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.GET("/", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	id := rec.Header().Get("X-Request-ID")
	require.NotEmpty(t, id, "expected generated request id in response")
}

func TestRequestID_PropagatesWhenPresent(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.GET("/", func(c *echo.Context) error {
		require.Equal(t, "existing-id", middleware.RequestIDFromContext(c))
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "existing-id", rec.Header().Get("X-Request-ID"))
}
