//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

const maxRequestIDLen = 128

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

func TestRequestID_RejectsBlankHeader(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Use(middleware.RequestID())
			e.GET("/", func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.input != "" {
				req.Header.Set("X-Request-ID", tc.input)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			got := rec.Header().Get("X-Request-ID")
			require.NotEmpty(t, got, "expected generated request id in response")
			require.NotEqual(t, tc.input, got, "blank header must not be echoed back")
			_, err := uuid.Parse(got)
			require.NoError(t, err, "generated id must be a valid UUID")
		})
	}
}

func TestRequestID_RejectsForbiddenCharacters(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "plus", input: "abc+def"},
		{name: "slash", input: "ab/cd"},
		{name: "equals", input: "ab=cd"},
		{name: "space", input: "ab cd"},
		{name: "control_char", input: "ab\x00cd"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Use(middleware.RequestID())
			e.GET("/", func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Request-ID", tc.input)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			got := rec.Header().Get("X-Request-ID")
			require.NotEmpty(t, got, "expected generated request id in response")
			require.NotEqual(t, tc.input, got, "invalid header must not be echoed back")
			_, err := uuid.Parse(got)
			require.NoError(t, err, "generated id must be a valid UUID")
		})
	}
}

func TestRequestID_LengthBoundary(t *testing.T) {
	t.Run("at_max_accepted", func(t *testing.T) {
		input := strings.Repeat("a", maxRequestIDLen)

		e := echo.New()
		e.Use(middleware.RequestID())
		e.GET("/", func(c *echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", input)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, input, rec.Header().Get("X-Request-ID"),
			"value at max length must be echoed back unchanged")
	})

	t.Run("over_max_rejected", func(t *testing.T) {
		input := strings.Repeat("a", maxRequestIDLen+1)

		e := echo.New()
		e.Use(middleware.RequestID())
		e.GET("/", func(c *echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", input)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := rec.Header().Get("X-Request-ID")
		require.NotEqual(t, input, got, "over-length header must not be echoed back")
		_, err := uuid.Parse(got)
		require.NoError(t, err, "generated id must be a valid UUID")
	})
}

func TestRequestID_GeneratesValidUUIDOnInvalid(t *testing.T) {
	const invalid = "bad value!"

	e := echo.New()
	e.Use(middleware.RequestID())
	e.GET("/", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", invalid)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	got := rec.Header().Get("X-Request-ID")
	require.NotEqual(t, invalid, got, "invalid header must not be echoed back")
	parsed, err := uuid.Parse(got)
	require.NoError(t, err, "generated id must be a valid UUID")
	require.NotEqual(t, uuid.Nil, parsed, "generated id must not be the nil UUID")
}
