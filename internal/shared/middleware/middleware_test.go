//go:build unit
// +build unit

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

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
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
	e := echo.New()
	e.Use(middleware.OTel())
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestOTel_RecordsError(t *testing.T) {
	e := echo.New()
	e.Use(middleware.OTel())
	e.GET("/bad", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
