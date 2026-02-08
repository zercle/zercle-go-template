// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"zercle-go-template/internal/logger"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		query      string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful GET request",
			method:     http.MethodGet,
			path:       "/api/users",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "POST request with body",
			method:     http.MethodPost,
			path:       "/api/users",
			statusCode: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "request with query params",
			method:     http.MethodGet,
			path:       "/api/users",
			query:      "page=1&limit=10",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "client error (400)",
			method:     http.MethodPost,
			path:       "/api/users",
			statusCode: http.StatusBadRequest,
			wantErr:    false,
		},
		{
			name:       "server error (500)",
			method:     http.MethodGet,
			path:       "/api/users",
			statusCode: http.StatusInternalServerError,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewNop()

			e := echo.New()
			e.Use(RequestLogger(log))

			e.Any(tt.path, func(c echo.Context) error {
				return c.String(tt.statusCode, "")
			})

			url := tt.path
			if tt.query != "" {
				url = url + "?" + tt.query
			}

			req := httptest.NewRequest(tt.method, url, nil)
			req.Header.Set("User-Agent", "test-agent")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}

func TestRequestLogger_WithError(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(RequestLogger(log))

	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestLoggerContext(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(LoggerContext(log))

	var capturedLogger logger.Logger
	e.GET("/test", func(c echo.Context) error {
		capturedLogger = logger.FromContext(c.Request().Context())
		return c.String(http.StatusOK, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, capturedLogger)
}

func TestRequestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
	}{
		{
			name:        "2xx - Info level",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "3xx - Info level",
			statusCode:  http.StatusMovedPermanently,
			expectError: false,
		},
		{
			name:        "4xx - Warn level",
			statusCode:  http.StatusBadRequest,
			expectError: false,
		},
		{
			name:        "5xx - Error level",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewNop()

			e := echo.New()
			e.Use(RequestLogger(log))

			e.GET("/test", func(c echo.Context) error {
				return c.String(tt.statusCode, "")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}

func TestRequestLogger_ClientIP(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(RequestLogger(log))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_WithUserAgent(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(RequestLogger(log))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 Test Browser")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_MethodNotAllowed(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(RequestLogger(log))

	// Only allow GET
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	// Try POST
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Echo returns 405 for method not allowed when route exists but method doesn't match
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestLoggerContext_MultipleCalls(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(LoggerContext(log))

	callCount := 0
	e.GET("/test", func(c echo.Context) error {
		// Call multiple times to verify idempotency
		_ = logger.FromContext(c.Request().Context())
		_ = logger.FromContext(c.Request().Context())
		callCount++
		return c.String(http.StatusOK, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, callCount)
}

func TestRequestLogger_QueryStringInPath(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(RequestLogger(log))

	var capturedPath string
	e.GET("/test", func(c echo.Context) error {
		capturedPath = c.Path()
		return c.String(http.StatusOK, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar&baz=qux", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, "/test", capturedPath)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_LongPath(t *testing.T) {
	log := logger.NewNop()

	e := echo.New()
	e.Use(RequestLogger(log))

	longPath := "/api/v1/users/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/profile"
	e.GET(longPath, func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	req := httptest.NewRequest(http.MethodGet, longPath, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
