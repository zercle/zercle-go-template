//go:build !user && !all
// +build !user,!all

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
)

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(app *fiber.App, container do.Injector) {
	// Empty implementation - user not included in this build
}
