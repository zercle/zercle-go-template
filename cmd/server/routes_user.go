//go:build user || all

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	userHandler "github.com/zercle/zercle-go-template/internal/features/user/handler"
)

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(app *fiber.App, container do.Injector) {
	handler := do.MustInvoke[*userHandler.UserHandler](container)
	api := app.Group("/api/v1")
	handler.RegisterRoutes(api)
}
