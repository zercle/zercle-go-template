// Package testutil provides small helpers for HTTP handler and end-to-end
// tests.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

// NewRequest builds an HTTP request with JSON content-type for use with echo.
func NewRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// DoJSON performs an HTTP request against the echo server with a JSON body.
// The body value is marshaled to JSON; pass nil for an empty body.
func DoJSON(t *testing.T, e *echo.Echo, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}

	rec := httptest.NewRecorder()
	req := NewRequest(method, path, reader)
	e.ServeHTTP(rec, req)
	return rec
}

// DecodeJSON unmarshals the response body into dst.
func DecodeJSON(t *testing.T, rr *httptest.ResponseRecorder, dst any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), dst))
}
