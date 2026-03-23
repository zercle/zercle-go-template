// Package router provides HTTP routing setup using Echo v5.
package router

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
	httpmiddleware "github.com/zercle/zercle-go-template/internal/transport/http/middleware"
)

// RequestTimeout creates a timeout middleware for Echo v5.
func RequestTimeout(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// Setup configures the HTTP router with all routes and middleware.
// It returns a configured echo.Echo with user and task routes mounted.
func Setup(logger *logging.Logger, userHandler user_entity.Handler, taskHandler task_entity.Handler) *echo.Echo {
	e := echo.New()

	// Core middleware from Echo
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	// Custom zerolog middleware
	e.Use(httpmiddleware.ZerologMiddleware(logger))

	// Request timeout middleware (60 seconds)
	e.Use(RequestTimeout(60 * time.Second))

	// Health check endpoint
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// API v1 routes
	api := e.Group("/api/v1")
	{
		// Mount user routes
		users := api.Group("/users")
		userHandler.RegisterRoutes(users)

		// Mount task routes
		tasks := api.Group("/tasks")
		taskHandler.RegisterRoutes(tasks)
	}

	return e
}
