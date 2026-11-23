//go:build unit

package middleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/transport/middleware"
)

func TestSecure_SetsSecurityHeaders(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{HTTP: config.HTTPConfig{}}

	e := echo.New()
	e.Use(middleware.Secure(cfg))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
	require.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
}

func TestSecure_NoHSTSWhenMaxAgeZero(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{HTTP: config.HTTPConfig{HSTSMaxAge: 0}}

	e := echo.New()
	e.Use(middleware.Secure(cfg))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Empty(t, rec.Header().Get("Strict-Transport-Security"))
}

func TestSecure_HSTSWhenMaxAgeSet(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{HTTP: config.HTTPConfig{HSTSMaxAge: time.Hour}}

	e := echo.New()
	e.Use(middleware.Secure(cfg))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Contains(t, rec.Header().Get("Strict-Transport-Security"), "max-age=3600")
}

func TestSecure_NilConfigDoesNotPanic(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(middleware.Secure(nil))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Empty(t, rec.Header().Get("Strict-Transport-Security"))
}
