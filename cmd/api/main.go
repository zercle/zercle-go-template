// Main package for the API server.
// It initializes the application, sets up dependencies, and starts the HTTP server.
//
//	@title			Zercle Go Template API
//	@version		1.0.0
//	@description	Production-ready Go web application template with JWT authentication
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.email	support@zercle.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
//	@description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "zercle-go-template/api/docs"
	"zercle-go-template/internal/config"
	"zercle-go-template/internal/container"
	userhandler "zercle-go-template/internal/feature/user/handler"
	"zercle-go-template/internal/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	log := logger.New(cfg.App.Name, cfg.App.Environment)
	log.Info("starting application",
		logger.String("name", cfg.App.Name),
		logger.String("version", cfg.App.Version),
		logger.String("environment", cfg.App.Environment),
	)

	// Initialize dependency container
	container, err := container.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize container: %w", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Error("failed to close container", logger.Error(err))
		}
	}()

	// Initialize Echo app
	e := echo.New()
	e.HideBanner = cfg.IsProduction()
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	// Setup routes
	setupRouter(e, container, log)

	// Start server in a goroutine
	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		log.Info("HTTP server starting", logger.String("address", address))
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server failed", logger.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", logger.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Info("server exited gracefully")
	return nil
}

// setupRouter initializes the HTTP router with middleware and routes.
func setupRouter(e *echo.Echo, container *container.Container, log logger.Logger) {
	// Add Echo built-in middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogMethod: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("HTTP request",
				logger.String("method", v.Method),
				logger.String("uri", v.URI),
				logger.Int("status", v.Status),
				logger.Duration("latency", v.Latency),
			)
			return nil
		},
	}))
	e.Use(middleware.RequestID())

	// Health check endpoint
	e.GET("/health", healthCheck)

	// Swagger documentation endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API routes
	api := e.Group("/api/v1")
	{
		// Initialize handlers
		userHandler := userhandler.NewUserHandler(container.UserUsecase, container.JWTUsecase, log)
		userHandler.RegisterRoutes(api)
	}

	// 404 handler - Echo handles this with a custom route
	e.RouteNotFound("/*", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":    "NOT_FOUND",
				"message": "Resource not found",
			},
		})
	})
}

// healthCheck handles health check requests.
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		},
	})
}
