//go:build !health && !all
// +build !health,!all

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
)

// RegisterHealthRoutes registers health endpoints
func RegisterHealthRoutes(app *fiber.App, container do.Injector) {
	// Empty implementation - health not included in this build
}
