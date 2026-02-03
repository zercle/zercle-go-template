package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/zercle/zercle-go-template/internal/feature/auth"
	"github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/feature/user"
	infra_auth "github.com/zercle/zercle-go-template/internal/infrastructure/auth"
	"github.com/zercle/zercle-go-template/internal/infrastructure/database"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
	"github.com/zercle/zercle-go-template/internal/infrastructure/observability"
	"github.com/zercle/zercle-go-template/pkg/config"
)

// Container holds all application dependencies.
type Container struct {
	Config         *config.Config
	DB             *pgxpool.Pool
	Logger         *logging.Logger
	UserRepo       user.Repository
	UserSvc        *user.Service
	UserHandler    *user.Handler
	CredRepo       auth.CredentialRepository
	TokenRepo      auth.RefreshTokenRepository
	AuthSvc        *auth.Service
	AuthHandler    *auth.Handler
	TaskRepo       task.Repository
	TaskSvc        *task.Service
	TaskHandler    *task.Handler
	TokenSvc       infra_auth.TokenService
	PwdHasher      infra_auth.PasswordHasher
	HealthAgg      *observability.HealthAggregator
	Registry       *prometheus.Registry
	TracerShutdown func()
}

// New creates and initializes a new application Container.
func New() *Container {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	logger := logging.New(cfg.Log)

	db, err := database.New(cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	healthAgg := observability.NewHealthAggregator()
	healthAgg.Register(observability.NewDatabaseHealthChecker(db.Pool))

	tracerProvider, shutdown, err := observability.InitTracer(observability.TracingConfig{
		Enabled:     cfg.Observability.TracingEnabled,
		Endpoint:    cfg.Observability.TracingEndpoint,
		ServiceName: cfg.Observability.ServiceName,
	})
	if err != nil {
		logger.Warn().Err(err).Msg("failed to initialize tracer")
	}
	_ = tracerProvider

	tokenSvc := infra_auth.NewTokenService(cfg.Auth)
	pwdHasher := infra_auth.NewPasswordHasher()

	userRepo := user.NewPostgresRepository(db.Pool)
	credRepo := auth.NewCredentialRepository(db.Pool)
	tokenRepo := auth.NewRefreshTokenRepository(db.Pool)
	taskRepo := task.NewPostgresRepository(db.Pool)

	userSvc := user.NewServiceWithLogger(userRepo, logger.With().Str("component", "user-service").Logger())
	authSvc := auth.NewServiceWithLogger(userSvc, credRepo, tokenRepo, tokenSvc, pwdHasher, logger.With().Str("component", "auth-service").Logger())
	taskSvc := task.NewServiceWithLogger(taskRepo, logger.With().Str("component", "task-service").Logger())

	userHandler := user.NewHandler(userSvc, logger.With().Str("component", "user-handler").Logger())
	authHandler := auth.NewHandler(authSvc, logger.With().Str("component", "auth-handler").Logger())
	taskHandler := task.NewHandler(taskSvc, logger.With().Str("component", "task-handler").Logger())

	return &Container{
		Config:         cfg,
		DB:             db.Pool,
		Logger:         logger,
		UserRepo:       userRepo,
		UserSvc:        userSvc,
		UserHandler:    userHandler,
		CredRepo:       credRepo,
		TokenRepo:      tokenRepo,
		AuthSvc:        authSvc,
		AuthHandler:    authHandler,
		TaskRepo:       taskRepo,
		TaskSvc:        taskSvc,
		TaskHandler:    taskHandler,
		TokenSvc:       tokenSvc,
		PwdHasher:      pwdHasher,
		HealthAgg:      healthAgg,
		Registry:       registry,
		TracerShutdown: func() { _ = shutdown(context.Background()) },
	}
}

// Shutdown gracefully shuts down the application container.
func (c *Container) Shutdown() {
	c.TracerShutdown()
	c.DB.Close()
}
