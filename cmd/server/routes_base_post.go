//go:build !post && !all
// +build !post,!all

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
)

// RegisterPostRoutes registers post-related routes
func RegisterPostRoutes(app *fiber.App, container do.Injector) {
	// Empty implementation - post not included in this build
}
