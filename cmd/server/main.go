// Package main provides the HTTP server entry point for the Zercle online booking system.
//
// The application follows clean architecture principles with dependency injection and graceful shutdown.
// Configuration is loaded from YAML files and can be overridden by environment variables.
//
// Startup sequence:
//  1. Load configuration based on SERVER_ENV (defaults to "local")
//  2. Initialize structured logger
//  3. Connect to PostgreSQL database
//  4. Initialize Echo web framework
//  5. Register middleware (request ID, logging, CORS, rate limiting)
//  6. Initialize domain layers (repositories, use cases, handlers)
//  7. Register API routes (public and protected)
//  8. Start HTTP server with graceful shutdown handling
//
// Graceful shutdown: The server listens for SIGINT/SIGTERM signals and allows
// up to 30 seconds for in-flight requests to complete before terminating.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "github.com/zercle/zercle-go-template/docs" // Import docs for swagger
	bookingHandler "github.com/zercle/zercle-go-template/domain/booking/handler"
	bookingRepository "github.com/zercle/zercle-go-template/domain/booking/repository"
	bookingUsecase "github.com/zercle/zercle-go-template/domain/booking/usecase"
	paymentHandler "github.com/zercle/zercle-go-template/domain/payment/handler"
	paymentRepository "github.com/zercle/zercle-go-template/domain/payment/repository"
	paymentUsecase "github.com/zercle/zercle-go-template/domain/payment/usecase"
	serviceHandler "github.com/zercle/zercle-go-template/domain/service/handler"
	serviceRepository "github.com/zercle/zercle-go-template/domain/service/repository"
	serviceUsecase "github.com/zercle/zercle-go-template/domain/service/usecase"
	userHandler "github.com/zercle/zercle-go-template/domain/user/handler"
	userRepository "github.com/zercle/zercle-go-template/domain/user/repository"
	userUsecase "github.com/zercle/zercle-go-template/domain/user/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/db"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/health"
	"github.com/zercle/zercle-go-template/pkg/middleware"
)

// @title           Zercle Go Template API
// @version         1.0
// @description     A production-ready RESTful API template built with Go Echo framework, featuring clean architecture, JWT authentication, and online booking capabilities.

// @contact.name   API Support
// @contact.url    https://github.com/zercle/zercle-go-template
// @contact.email  support@zercle.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:3000
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	env := getEnv()

	cfg, err := config.Load("./configs/" + env + ".yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	log := logger.NewLogger(&cfg.Logging)
	log.Info("Starting zercle-go-template server", "env", cfg.Server.Env)

	database, err := db.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	log.Info("Database connected successfully")

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Validator = middleware.NewCustomValidator()

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(log))
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORS(&cfg.CORS))
	e.Use(middleware.RateLimit(&cfg.RateLimit))

	registerRoutes(e, database, cfg, log)

	startServer(e, cfg, log)
}

// getEnv retrieves the deployment environment from the SERVER_ENV environment variable.
// Defaults to "local" if not set.
func getEnv() string {
	env := os.Getenv("SERVER_ENV")
	if env == "" {
		env = "local"
	}
	return env
}

// startServer starts the HTTP server in a goroutine and blocks until a shutdown signal is received.
// Performs graceful shutdown with a 30-second timeout for in-flight requests to complete.
func startServer(e *echo.Echo, cfg *config.Config, log *logger.Logger) {
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Info("Server starting", "address", addr)

		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", "error", err)
	}

	log.Info("Server stopped")
}

// registerRoutes initializes all domain layers and registers API routes with Echo.
// Initializes repositories, use cases, and handlers for user, service, booking, and payment domains.
// Registers health check endpoints, Swagger documentation, and API v1 routes (public and protected).
func registerRoutes(e *echo.Echo, database *db.Database, cfg *config.Config, log *logger.Logger) {
	userRepo := userRepository.Initialize(database.Queries(), log)
	userUseCase := userUsecase.Initialize(userRepo, cfg, log)
	userHdlr := userHandler.Initialize(userUseCase, log)

	serviceRepo := serviceRepository.Initialize(database.Queries(), log)
	serviceUseCase := serviceUsecase.Initialize(serviceRepo, log)
	serviceHdlr := serviceHandler.Initialize(serviceUseCase, log)

	bookingRepo := bookingRepository.Initialize(database.Queries(), log)
	bookingUseCase := bookingUsecase.Initialize(bookingRepo, serviceRepo, log)
	bookingHdlr := bookingHandler.Initialize(bookingUseCase, log)

	paymentRepo := paymentRepository.Initialize(database.Queries(), log)
	paymentUseCase := paymentUsecase.Initialize(paymentRepo, log)
	paymentHdlr := paymentHandler.Initialize(paymentUseCase, log)

	healthHandler := health.NewHandler(database)

	e.GET("/health", healthHandler.Check)
	e.GET("/readiness", healthHandler.Check)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")

	api.POST("/auth/register", userHdlr.Register)
	api.POST("/auth/login", userHdlr.Login)

	api.GET("/services", serviceHdlr.ListServices)
	api.GET("/services/search", serviceHdlr.SearchServices)
	api.GET("/services/:id", serviceHdlr.GetService)

	api.GET("/bookings/dates", bookingHdlr.ListBookingsByDateRange)
	api.GET("/bookings/services/:id", bookingHdlr.ListBookingsByService)

	protected := api.Group("")
	protected.Use(middleware.JWTAuth(&cfg.JWT))

	protected.GET("/users/profile", userHdlr.GetProfile)
	protected.PUT("/users/profile", userHdlr.UpdateProfile)
	protected.DELETE("/users/profile", userHdlr.DeleteAccount)
	protected.GET("/users", userHdlr.ListUsers)

	protected.POST("/services", serviceHdlr.CreateService)
	protected.PUT("/services/:id", serviceHdlr.UpdateService)
	protected.DELETE("/services/:id", serviceHdlr.DeleteService)

	protected.POST("/bookings", bookingHdlr.CreateBooking)
	protected.GET("/bookings", bookingHdlr.ListBookingsByUser)
	protected.GET("/bookings/:id", bookingHdlr.GetBooking)
	protected.PUT("/bookings/:id/status", bookingHdlr.UpdateBookingStatus)
	protected.PUT("/bookings/:id/cancel", bookingHdlr.CancelBooking)

	protected.POST("/payments", paymentHdlr.CreatePayment)
	protected.GET("/payments", paymentHdlr.ListPayments)
	protected.GET("/payments/:id", paymentHdlr.GetPayment)
	protected.GET("/bookings/:booking_id/payments", paymentHdlr.GetPaymentByBooking)
	protected.PUT("/payments/:id/confirm", paymentHdlr.ConfirmPayment)
	protected.PUT("/payments/:id/refund", paymentHdlr.RefundPayment)
}
