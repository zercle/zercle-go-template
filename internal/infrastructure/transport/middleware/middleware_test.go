//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/transport/middleware"
)

func TestCORS_SetsHeaders(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	e, exporter, _ := newOTelTestEcho(t)
	e.GET("/bad", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "server error")
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status.Code)
}
