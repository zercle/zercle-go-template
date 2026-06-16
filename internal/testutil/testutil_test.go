//go:build unit

package testutil_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/testutil"
	"github.com/zercle/zercle-go-template/internal/testutil/fixtures"
)

func TestNewRequest_JSON(t *testing.T) {
	req := testutil.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte(`{}`)))
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestDoJSON(t *testing.T) {
	e := echo.New()
	e.POST("/echo", func(c *echo.Context) error {
		var body map[string]any
		if err := c.Bind(&body); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, body)
	})

	rec := testutil.DoJSON(t, e, http.MethodPost, "/echo", map[string]any{"key": "value"})
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "value")
}

func TestDecodeJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	_, err := rec.Body.WriteString(`{"key":"value"}`)
	require.NoError(t, err)

	var dst map[string]string
	testutil.DecodeJSON(t, rec, &dst)
	assert.Equal(t, "value", dst["key"])
}

func TestFixtures_NewItem(t *testing.T) {
	item := fixtures.NewItem("fixture")
	assert.Equal(t, "fixture", item.Name)
	assert.NotZero(t, item.ID)
	assert.NotZero(t, item.CreatedAt)
}

func TestNewRequest_NilBody(t *testing.T) {
	req := testutil.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, http.MethodGet, req.Method)
	body, _ := io.ReadAll(req.Body)
	assert.Empty(t, body)
}
