// Package main provides the entry point for the HTTP server.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/app"
	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
	"github.com/zercle/zercle-go-template/internal/transport/http/router"
)

func main() {
	// Initialize app container with all dependencies (loads config internally)
	c := app.New()

	// Get logger from container
	logger := do.MustInvoke[*logging.Logger](c.Injector())

	logger.Info().Str("host", "0.0.0.0").Int("port", 8080).Msg("starting server")

	// Setup router with handlers from container
	userHandler := do.MustInvoke[user_entity.Handler](c.Injector())
	taskHandler := do.MustInvoke[task_entity.Handler](c.Injector())
	e := router.Setup(logger, userHandler, taskHandler)

	// Create HTTP server using Echo's server
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           e,
		ReadHeaderTimeout: 5 * 1e9, // 5 seconds - protects against Slowloris attacks
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("addr", ":8080").Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("server error")
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("server forced to shutdown")
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error().Msg("shutdown timeout exceeded, forcing exit")
			os.Exit(1)
		}
	}

	// Shutdown container
	report := c.Shutdown()
	if len(report.Errors) > 0 {
		logger.Error().Interface("errors", report.Errors).Msg("container shutdown errors")
	}

	logger.Info().Msg("server exited gracefully")
}
