// Package container provides dependency injection for the application.
// It manages the lifecycle of all components using samber/do v2.
package app

import (
	"github.com/samber/do/v2"
)

// Container wraps the DI injector with type-safe accessors.
type Container struct {
	injector *do.RootScope
}

// New creates a new DI container with all services registered.
func New() *Container {
	injector := do.New()

	// Register all providers in dependency order
	RegisterConfig(injector)
	RegisterLogger(injector)
	RegisterDatabase(injector)
	RegisterRepositories(injector)
	RegisterServices(injector)
	RegisterHandlers(injector)

	return &Container{injector: injector}
}

// Shutdown gracefully stops all services.
func (c *Container) Shutdown() *do.ShutdownReport {
	return c.injector.Shutdown()
}

// Injector returns the underlying do.RootScope for advanced use cases.
func (c *Container) Injector() *do.RootScope {
	return c.injector
}
