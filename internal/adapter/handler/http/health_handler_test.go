package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	h "github.com/zercle/zercle-go-template/internal/adapter/handler/http"
)

func TestHealthHandler_HealthCheck(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := h.NewHealthHandler(db)
	app := fiber.New()
	app.Get("/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestHealthHandler_Liveness_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	handler := h.NewHealthHandler(db)
	app := fiber.New()
	app.Get("/health/live", handler.Liveness)

	// Mock Ping
	mock.ExpectPing()

	req := httptest.NewRequest("GET", "/health/live", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthHandler_Liveness_Failure(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	handler := h.NewHealthHandler(db)
	app := fiber.New()
	app.Get("/health/live", handler.Liveness)

	// Mock Ping Failure
	mock.ExpectPing().WillReturnError(errors.New("db down"))

	req := httptest.NewRequest("GET", "/health/live", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
