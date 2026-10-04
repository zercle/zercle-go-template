//go:build unit

package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/middleware"
)

func TestRecover_CatchesPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.Recover(&logger))
	e.GET("/panic", func(c *echo.Context) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, buf.String(), "panic")
}

func TestRecover_CatchesPanicError(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.Recover(&logger))
	e.GET("/panic", func(c *echo.Context) error {
		panic(errors.New("panic error"))
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, buf.String(), "panic error")
}

// TestRecover_RepanicsAbortHandler pins that http.ErrAbortHandler propagates
// instead of being converted into a logged 500: net/http uses that sentinel to
// abort a request silently, and echo's own recover middleware re-panics it.
func TestRecover_RepanicsAbortHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.Recover(&logger))
	e.GET("/abort", func(c *echo.Context) error {
		panic(http.ErrAbortHandler)
	})

	req := httptest.NewRequest(http.MethodGet, "/abort", nil)
	rec := httptest.NewRecorder()

	require.PanicsWithValue(t, http.ErrAbortHandler, func() {
		e.ServeHTTP(rec, req)
	})
	require.Empty(t, buf.String(), "an intentional abort must not be logged as a panic")
}

// TestAccessLog_WrapsRecover pins the middleware ordering contract: AccessLog
// must be registered outside Recover so a panicking request still yields one
// access-log line with status 500 rather than only a recovered-panic entry.
func TestAccessLog_WrapsRecover(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.AccessLog(&logger))
	e.Use(middleware.Recover(&logger))
	e.GET("/panic", func(c *echo.Context) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	out := buf.String()
	require.Contains(t, out, "http request", "panicked requests must still be access-logged")
	require.Contains(t, out, `"status":500`)
}

func TestAccessLog_WritesLogLine(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(middleware.AccessLog(&logger))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Contains(t, buf.String(), "http request")
	require.Contains(t, buf.String(), "204")
}

func TestCORS_SetsHeaders(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			CORSAllowOrigins: []string{"https://example.com"},
			CORSAllowMethods: []string{"GET", "POST"},
			CORSAllowHeaders: []string{"Content-Type"},
		},
	}

	e := echo.New()
	e.Use(middleware.CORS(cfg))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/ok", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func TestOTel_StartsSpan(t *testing.T) {
	e, _, _ := newOTelTestEcho(t)
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestOTel_RecordsError(t *testing.T) {
	e, exporter, _ := newOTelTestEcho(t)
	e.GET("/bad", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status.Code)
}
