//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

func TestCORS_NilConfigAppliesPackageDefaults(t *testing.T) {
	e := echo.New()
	e.Use(middleware.CORS(nil))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/ok", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"),
		"nil cfg should default to allowing all origins")

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	require.Contains(t, allowHeaders, "Authorization",
		"nil cfg branch must permit Authorization (comment 14)")

	require.NotEmpty(t, rec.Header().Get("Access-Control-Max-Age"),
		"nil cfg branch must set Access-Control-Max-Age")
}

func TestCORS_ConfigAppliesExposeAndMaxAge(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			CORSAllowOrigins: []string{"https://app.example.com"},
		},
	}

	e := echo.New()
	e.Use(middleware.CORS(cfg))
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// Preflight: assert Allow-Origin and Max-Age (echo sets Expose-Headers
	// on non-preflight actual CORS requests only).
	preflightReq := httptest.NewRequest(http.MethodOptions, "/ok", nil)
	preflightReq.Header.Set("Origin", "https://app.example.com")
	preflightReq.Header.Set("Access-Control-Request-Method", "GET")
	preflightRec := httptest.NewRecorder()
	e.ServeHTTP(preflightRec, preflightReq)

	require.Equal(t, http.StatusNoContent, preflightRec.Code)
	require.Equal(t, "https://app.example.com", preflightRec.Header().Get("Access-Control-Allow-Origin"))
	require.NotEmpty(t, preflightRec.Header().Get("Access-Control-Max-Age"),
		"preflight response must include Access-Control-Max-Age")

	// Actual CORS request: assert Expose-Headers contains Content-Length.
	actualReq := httptest.NewRequest(http.MethodGet, "/ok", nil)
	actualReq.Header.Set("Origin", "https://app.example.com")
	actualRec := httptest.NewRecorder()
	e.ServeHTTP(actualRec, actualReq)

	require.Equal(t, http.StatusNoContent, actualRec.Code)
	exposed := actualRec.Header().Get("Access-Control-Expose-Headers")
	require.True(t,
		strings.Contains(exposed, "Content-Length"),
		"expected Access-Control-Expose-Headers to contain Content-Length, got %q", exposed)
}
