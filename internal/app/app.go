package app

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
	_ "github.com/zercle/zercle-go-template/docs"
	"github.com/zercle/zercle-go-template/internal/domain/task"
	taskDomainHandler "github.com/zercle/zercle-go-template/internal/domain/task/handler"
	"github.com/zercle/zercle-go-template/internal/domain/task/repository"
	taskUsecase "github.com/zercle/zercle-go-template/internal/domain/task/usecase"
	"github.com/zercle/zercle-go-template/internal/domain/user"
	userHandler "github.com/zercle/zercle-go-template/internal/domain/user/handler"
	userRepo "github.com/zercle/zercle-go-template/internal/domain/user/repository"
	userUsecase "github.com/zercle/zercle-go-template/internal/domain/user/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
	sqlcDb "github.com/zercle/zercle-go-template/internal/infrastructure/sqlc/db"
	"github.com/zercle/zercle-go-template/pkg/health"
	"github.com/zercle/zercle-go-template/pkg/middleware"
)

type App struct {
	cfg           *config.Config
	log           *logger.Logger
	db            db.Database
	echo          *echo.Echo
	userHandler   user.UserHandler
	taskHandler   task.TaskHandler
	healthHandler *health.Handler
}

func NewApp(cfg *config.Config, log *logger.Logger) (*App, error) {
	database, err := db.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	queries, ok := database.Queries().(*sqlcDb.Queries)
	if !ok || queries == nil {
		return nil, fmt.Errorf("unsupported database type for user repository")
	}

	userRepository := userRepo.NewUserRepository(queries, log)
	userUseCase := userUsecase.NewUserUseCase(userRepository, &cfg.JWT, &cfg.Argon2id, log)
	userHandler := userHandler.NewUserHandler(userUseCase, log)

	var taskHandler task.TaskHandler
	if pgxDB, ok := database.(*db.PostgresDatabase); ok {
		taskRepository := repository.NewTaskRepository(pgxDB.Pool)
		taskUseCase := taskUsecase.NewTaskUseCase(taskRepository, log)
		taskHandler = taskDomainHandler.NewTaskHandler(taskUseCase, log)
	} else {
		log.Warn("Task functionality requires PostgreSQL database")
	}

	healthHandler := health.NewHandler(database)

	echo := echo.New()
	echo.HideBanner = true
	echo.HidePort = true
	echo.Validator = middleware.NewCustomValidator()

	setupMiddleware(echo, cfg, log)
	registerRoutes(echo, cfg, log, userHandler, taskHandler, healthHandler)

	return &App{
		cfg:           cfg,
		log:           log,
		db:            database,
		echo:          echo,
		userHandler:   userHandler,
		taskHandler:   taskHandler,
		healthHandler: healthHandler,
	}, nil
}

func setupMiddleware(e *echo.Echo, cfg *config.Config, log *logger.Logger) {
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(log))
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORS(&cfg.CORS))
	e.Use(middleware.RateLimit(&cfg.RateLimit))
}

func registerRoutes(e *echo.Echo, cfg *config.Config, log *logger.Logger, userHandler user.UserHandler, taskHandler task.TaskHandler, healthHandler *health.Handler) {
	e.GET("/health", healthHandler.Check)
	e.GET("/readiness", healthHandler.Check)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")

	protected := api.Group("")
	protected.Use(middleware.JWTAuth(&cfg.JWT))

	userHandler.RegisterRoutes(api, protected)
	taskHandler.RegisterRoutes(protected)
}

func (a *App) Start() error {
	go func() {
		addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
		a.log.Info("Server starting", "address", addr)

		if err := a.echo.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	a.log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		a.log.Error("Server shutdown error", "error", err)
	}

	a.log.Info("Server stopped")
	return nil
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}

func (a *App) GetEcho() *echo.Echo {
	return a.echo
}
