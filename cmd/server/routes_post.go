//go:build post || all

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	postHandler "github.com/zercle/zercle-go-template/internal/features/post/handler"
)

// RegisterPostRoutes registers post-related routes
func RegisterPostRoutes(app *fiber.App, container do.Injector) {
	handler := do.MustInvoke[*postHandler.PostHandler](container)
	api := app.Group("/api/v1")
	handler.RegisterRoutes(api)
}
