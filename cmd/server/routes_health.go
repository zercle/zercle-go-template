//go:build health || all

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	healthHandler "github.com/zercle/zercle-go-template/internal/features/health/handler"
	"github.com/zercle/zercle-go-template/internal/core/port"
)

// RegisterHealthRoutes registers health endpoints
func RegisterHealthRoutes(app *fiber.App, container do.Injector) {
	service := do.MustInvoke[port.HealthService](container)
	handler := healthHandler.NewHealthHandler(service)
	app.Get("/health", handler.HealthCheck)
	app.Get("/health/live", handler.Liveness)
}
