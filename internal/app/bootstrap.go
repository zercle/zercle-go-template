package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zercle/zercle-go-template/internal/feature/auth/handlers/grpc"
	httpHandler "github.com/zercle/zercle-go-template/internal/feature/auth/handlers/http"
	authRepos "github.com/zercle/zercle-go-template/internal/feature/auth/repositories/postgres"
	authUsecases "github.com/zercle/zercle-go-template/internal/feature/auth/usecases"
	chatHandler "github.com/zercle/zercle-go-template/internal/feature/chat/handlers/http"
	chatRepos "github.com/zercle/zercle-go-template/internal/feature/chat/repositories/postgres"
	chatUsecases "github.com/zercle/zercle-go-template/internal/feature/chat/usecases"
	"github.com/zercle/zercle-go-template/internal/infrastructure/databases/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/loggers/zerolog"

	coreconfig "github.com/zercle/zercle-go-template/internal/infrastructure/configs"
)

// App represents the main application instance.
type App struct {
	Config *coreconfig.Config
	DB     *postgres.DB

	AuthHandler *httpHandler.Handler
	AuthGrpc    *grpc.Server
	ChatHandler *chatHandler.Handler
}

// New creates a new application instance.
func New() (*App, error) {
	cfg, err := coreconfig.Load("./configs")
	if err != nil {
		return nil, err
	}

	if err := zerolog.Init(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		return nil, err
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	return &App{
		Config: cfg,
		DB:     db,
	}, nil
}

// InitAuth initializes the authentication feature.
func (a *App) InitAuth() {
	userRepo := authRepos.NewUserRepository(a.DB.Pool)
	sessionRepo := authRepos.NewSessionRepository(a.DB.Pool)

	authService := authUsecases.NewAuthService(
		userRepo,
		sessionRepo,
		a.Config.Auth.JWTSecret,
		a.Config.Auth.JWTExpiry,
		a.Config.Auth.RefreshExpiry,
	)

	a.AuthHandler = httpHandler.NewHandler(authService)
	a.AuthGrpc = grpc.NewServer(authService)
}

// InitChat initializes the chat feature.
func (a *App) InitChat() {
	roomRepo := chatRepos.NewRoomRepository(a.DB.Pool)
	messageRepo := chatRepos.NewMessageRepository(a.DB.Pool)

	chatService := chatUsecases.NewService(roomRepo, messageRepo)

	a.ChatHandler = chatHandler.NewHandler(chatService)
}

// Close closes all application resources.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}

// GetPostgresPool returns the PostgreSQL connection pool.
func (a *App) GetPostgresPool() *pgxpool.Pool {
	return a.DB.Pool
}

// GetJWTExpiry returns the JWT expiration duration.
func (a *App) GetJWTExpiry() time.Duration {
	return a.Config.Auth.JWTExpiry
}

// GetJWTSecret returns the JWT secret key.
func (a *App) GetJWTSecret() string {
	return a.Config.Auth.JWTSecret
}

// GetServerConfig returns the server configuration.
func (a *App) GetServerConfig() coreconfig.ServerConfig {
	return a.Config.Server
}
