package response

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccess(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testData := map[string]string{"key": "value"}

	err := Success(c, testData)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, StatusSuccess, resp.Status)
	assert.NotNil(t, resp.Data)

	// Check that Data is the testData
	dataMap, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", dataMap["key"])
}

func TestFail(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testData := map[string]string{"error": "validation failed"}

	err := Fail(c, http.StatusBadRequest, testData)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, StatusFail, resp.Status)
	assert.NotNil(t, resp.Data)
}

func TestError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Error(c, http.StatusInternalServerError, "Something went wrong", 5000)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Something went wrong", resp.Message)
	assert.Equal(t, 5000, resp.Code)
}

func TestErrorWithCode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	testCases := []struct {
		name       string
		statusCode int
		message    string
		code       int
	}{
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			message:    "Internal server error",
			code:       CodeInternalError,
		},
		{
			name:       "database error",
			statusCode: http.StatusInternalServerError,
			message:    "Database operation failed",
			code:       CodeDatabaseError,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			message:    "Resource not found",
			code:       CodeNotFound,
		},
		{
			name:       "validation error",
			statusCode: http.StatusBadRequest,
			message:    "Validation failed",
			code:       CodeValidationError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := Error(c, tc.statusCode, tc.message, tc.code)

			require.NoError(t, err)
			assert.Equal(t, tc.statusCode, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, StatusError, resp.Status)
			assert.Equal(t, tc.message, resp.Message)
			assert.Equal(t, tc.code, resp.Code)
		})
	}
}

func TestWriteSuccess(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testData := map[string]int{"count": 42}

	err := Success(c, testData)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusSuccess, resp.Status)
}

func TestWriteFail(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testData := map[string]string{"field": "email", "message": "invalid format"}

	err := Fail(c, http.StatusUnprocessableEntity, testData)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestWriteError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Error(c, http.StatusServiceUnavailable, "Service temporarily unavailable", 5030)

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Service temporarily unavailable", resp.Message)
	assert.Equal(t, 5030, resp.Code)
}

func TestCreated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testData := map[string]string{"id": "123"}

	err := Created(c, testData)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusSuccess, resp.Status)
}

func TestNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := NoContent(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestBadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testErr := assert.AnError

	err := BadRequest(c, "invalid input", testErr)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := NotFound(c, "user")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestUnauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Unauthorized(c, "invalid token")

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestForbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Forbidden(c, "access denied")

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestConflict(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Conflict(c, "resource already exists")

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestInternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := InternalError(c, "database connection failed")

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "database connection failed", resp.Message)
	assert.Equal(t, CodeInternalError, resp.Code)
}

func TestValidationError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validationErr := assert.AnError

	err := ValidationError(c, validationErr)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
}

func TestResponseStructure(t *testing.T) {
	t.Run("success response has correct structure", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		data := map[string]string{"message": "hello"}
		err := Success(c, data)

		require.NoError(t, err)

		var resp Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, StatusSuccess, resp.Status)
		assert.NotNil(t, resp.Data)
		assert.Empty(t, resp.Message)
		assert.Zero(t, resp.Code)
	})

	t.Run("error response has correct structure", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := Error(c, http.StatusBadRequest, "validation failed", 4000)

		require.NoError(t, err)

		var resp Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, StatusError, resp.Status)
		assert.Nil(t, resp.Data)
		assert.Equal(t, "validation failed", resp.Message)
		assert.Equal(t, 4000, resp.Code)
	})
}
