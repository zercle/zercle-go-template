package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zercle/zercle-go-template/internal/core/port"
	sharedHandler "github.com/zercle/zercle-go-template/internal/shared/handler/response"
)

// HealthHandler handles system health checks.
type HealthHandler struct {
	service port.HealthService
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(service port.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

// HealthCheck godoc
// @Summary System Health Check
// @Description Checks if the application is running
// @Tags system
// @Produce json
// @Success 200 {object} sharedHandler.Response{data=string}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	ctx := c.Context()

	response, err := h.service.HealthCheck(ctx)
	if err != nil {
		return sharedHandler.Fail(c, fiber.StatusServiceUnavailable, response)
	}

	return sharedHandler.Success(c, fiber.StatusOK, response)
}

// Liveness godoc
// @Summary System Liveness Check
// @Description Checks if the application is alive and can connect to the database
// @Tags system
// @Produce json
// @Success 200 {object} sharedHandler.Response{data=string}
// @Failure 503 {object} sharedHandler.Response
// @Router /health/live [get]
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	ctx := c.Context()

	response, err := h.service.LivenessCheck(ctx)
	if err != nil {
		return sharedHandler.Fail(c, fiber.StatusServiceUnavailable, response)
	}

	if response.Status == "down" {
		return sharedHandler.Fail(c, fiber.StatusServiceUnavailable, response)
	}

	return sharedHandler.Success(c, fiber.StatusOK, response)
}
