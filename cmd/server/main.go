package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/swagger"
	"github.com/samber/do/v2"
	httpAdapter "github.com/zercle/zercle-go-template/internal/adapter/handler/http"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/container"
	"github.com/zercle/zercle-go-template/internal/infrastructure/server"
	"github.com/zercle/zercle-go-template/pkg/logger"

	_ "github.com/zercle/zercle-go-template/docs"
)

func main() {
	// 1. Setup global logger with zerolog
	logger.Setup(logger.Config{
		Level:  "debug",
		Pretty: false, // Set to true for development, false for production JSON logs
	})

	// 2. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// 3. Initialize DI with logger
	container, err := container.NewProdContainer(cfg, logger.Log)
	if err != nil {
		log.Fatalf("failed to initialize DI: %v", err)
	}

	// 4. Initialize Server (Infrastructure) with logger
	app := server.New(cfg, logger.Log)

	// 5. Register Routes
	userHandler := do.MustInvoke[*httpAdapter.UserHandler](container)
	postHandler := do.MustInvoke[*httpAdapter.PostHandler](container)

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1")

	// Register Routes
	userHandler.RegisterRoutes(api)
	postHandler.RegisterRoutes(api)

	// Health
	// Health & Liveness
	dbConn := do.MustInvoke[*sql.DB](container)
	healthHandler := httpAdapter.NewHealthHandler(dbConn)
	app.Get("/health", healthHandler.HealthCheck)
	app.Get("/health/live", healthHandler.Liveness)

	// Start
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Printf("Starting server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server exiting")
}
