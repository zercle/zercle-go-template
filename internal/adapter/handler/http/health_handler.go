package handler

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zercle/zercle-go-template/pkg/utils/response"
)

// HealthHandler handles system health checks.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck godoc
// @Summary System Health Check
// @Description Checks if the application is running
// @Tags system
// @Produce json
// @Success 200 {object} response.Response{data=string}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, "OK")
}

// Liveness godoc
// @Summary System Liveness Check
// @Description Checks if the application is alive and can connect to the database
// @Tags system
// @Produce json
// @Success 200 {object} response.Response{data=string}
// @Failure 503 {object} response.Response
// @Router /health/live [get]
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 1*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		return response.Fail(c, fiber.StatusServiceUnavailable, fiber.Map{
			"status": "down",
			"error":  "database unreachable",
		})
	}

	return response.Success(c, fiber.StatusOK, "Alive")
}
