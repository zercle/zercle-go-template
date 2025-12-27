package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func setupEchoContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestJSend_Success(t *testing.T) {
	c, rec := setupEchoContext()

	data := map[string]interface{}{"key": "value"}
	err := Created(c, data)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusSuccess, resp.Status)
	assert.Equal(t, data, resp.Data)
}

func TestJSend_OK(t *testing.T) {
	c, rec := setupEchoContext()

	data := map[string]interface{}{"key": "value"}
	err := OK(c, data)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusSuccess, resp.Status)
	assert.Equal(t, data, resp.Data)
}

func TestJSend_Fail(t *testing.T) {
	c, rec := setupEchoContext()

	errors := []FieldError{
		{Field: "email", Message: "Invalid email format"},
		{Field: "password", Message: "Password too short"},
	}
	err := Fail(c, http.StatusBadRequest, "Validation failed", errors)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Len(t, resp.Errors, 2)
	assert.Equal(t, "email", resp.Errors[0].Field)
}

func TestJSend_Error(t *testing.T) {
	c, rec := setupEchoContext()

	err := Error(c, http.StatusInternalServerError, "Something went wrong")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Something went wrong", resp.Message)
}

func TestJSend_BadRequest(t *testing.T) {
	c, rec := setupEchoContext()

	err := BadRequest(c, "Invalid input", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusFail, resp.Status)
	assert.Equal(t, "Invalid input", resp.Message)
}

func TestJSend_Unauthorized(t *testing.T) {
	c, rec := setupEchoContext()

	err := Unauthorized(c, "Token expired")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Token expired", resp.Message)
}

func TestJSend_Forbidden(t *testing.T) {
	c, rec := setupEchoContext()

	err := Forbidden(c, "Access denied")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Access denied", resp.Message)
}

func TestJSend_NotFound(t *testing.T) {
	c, rec := setupEchoContext()

	err := NotFound(c, "Resource not found")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Resource not found", resp.Message)
}

func TestJSend_InternalError(t *testing.T) {
	c, rec := setupEchoContext()

	err := InternalError(c, "Internal server error")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp JSend
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, "Internal server error", resp.Message)
}
