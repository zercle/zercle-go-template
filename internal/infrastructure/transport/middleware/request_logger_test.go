//go:build unit

package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/transport/middleware"
)

func TestRequestLogger_WritesOneLine(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger(&logger))
	e.Use(echomw.RecoverWithConfig(echomw.DefaultRecoverConfig))
	e.GET("/hello", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello?x=1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	out := buf.String()
	require.Contains(t, out, "http request")
	require.Contains(t, out, `"method":"GET"`)
	require.Contains(t, out, `"path":"/hello"`)
	require.Contains(t, out, `"status":204`)
	require.Contains(t, out, `"request_id"`)
	require.NotContains(t, out, "x=1", "path must not include query string")
}

func TestRequestLogger_PanickingHandlerLogs500(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger(&logger))
	e.Use(echomw.RecoverWithConfig(echomw.DefaultRecoverConfig))
	e.GET("/panic", func(c *echo.Context) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	out := buf.String()
	require.Contains(t, out, "http request", "panicked request must produce a log line")
	require.Contains(t, out, `"status":500`)
}
