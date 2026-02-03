// Package router provides HTTP routing setup using Echo v5.
package router

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	httpSwagger "github.com/swaggo/echo-swagger/v2"

	"github.com/zercle/zercle-go-template/internal/feature/auth"
	"github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/feature/user"
	infra_auth "github.com/zercle/zercle-go-template/internal/infrastructure/auth"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
	"github.com/zercle/zercle-go-template/internal/infrastructure/observability"
	"github.com/zercle/zercle-go-template/pkg/config"

	// Import swagger docs for side-effect registration.
	_ "github.com/zercle/zercle-go-template/internal/transport/http/swagger"
)

// Setup configures and returns the Echo router with all routes and middleware.
func Setup(
	logger *logging.Logger,
	cfg *config.Config,
	userHandler *user.Handler,
	authHandler *auth.Handler,
	taskHandler *task.Handler,
	tokenService infra_auth.TokenService,
	healthAgg *observability.HealthAggregator,
) *echo.Echo {
	e := echo.New()

	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())
	e.Use(RequestTimeout(60 * time.Second))

	e.GET("/swagger/*", httpSwagger.WrapHandler)

	e.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.GET("/readyz", func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		health := healthAgg.Check(ctx)
		if health.Status != "healthy" {
			return c.JSON(503, health)
		}
		return c.JSON(200, health)
	})

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	jwtMiddleware := infra_auth.JWTMiddleware(tokenService)

	api := e.Group("/api/v1")

	authPublic := api.Group("/auth")
	authHandler.RegisterPublicRoutes(authPublic)

	authProtected := api.Group("/auth", jwtMiddleware)
	authHandler.RegisterProtectedRoutes(authProtected)

	users := api.Group("/users", jwtMiddleware)
	userHandler.RegisterProtectedRoutes(users)

	tasks := api.Group("/tasks", jwtMiddleware)
	taskHandler.RegisterProtectedRoutes(tasks)

	return e
}

// RequestTimeout returns middleware that sets a request context timeout.
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
