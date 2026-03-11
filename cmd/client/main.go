package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"

	authhttp "github.com/zercle/zercle-go-template/internal/features/auth/handler/http"
	authservice "github.com/zercle/zercle-go-template/internal/features/auth/service"
	chathttp "github.com/zercle/zercle-go-template/internal/features/chat/handler/http"
	ssehandler "github.com/zercle/zercle-go-template/internal/features/chat/handler/sse"
	"github.com/zercle/zercle-go-template/internal/features/chat/messaging"
	chatservice "github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

func main() {
	cfg, err := config.Load("./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		os.Exit(1)
	}
	defer db.Close()

	valkeyClient, err := valkey.New(cfg.Valkey)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to Valkey")
		os.Exit(1)
	}
	defer func() { _ = valkeyClient.Close() }()

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	roomRepo := postgres.NewRoomRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	authSvc := authservice.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpiry,
		cfg.Auth.RefreshExpiry,
	)

	chatPubsub := messaging.New(valkeyClient)
	chatSvc := chatservice.NewChatServiceWithPubSub(roomRepo, messageRepo, chatPubsub)

	authHTTPHandler := authhttp.NewAuthHandler(authSvc)
	chatHTTPHandler := chathttp.NewChatHandler(chatSvc)
	sseHandler := ssehandler.NewHandler(valkeyClient)

	e := echo.New()

	e.Use(echomiddleware.RequestLogger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.GET("/ready", func(c *echo.Context) error {
		ctx := c.Request().Context()
		if err := db.Ping(ctx); err != nil {
			return c.JSON(503, map[string]string{"status": "not ready", "database": "unavailable"})
		}
		if err := valkeyClient.Ping(ctx); err != nil {
			return c.JSON(503, map[string]string{"status": "not ready", "valkey": "unavailable"})
		}
		return c.JSON(200, map[string]string{"status": "ready", "database": "ok", "valkey": "ok"})
	})

	e.POST("/api/v1/auth/register", authHTTPHandler.Register)
	e.POST("/api/v1/auth/login", authHTTPHandler.Login)
	e.POST("/api/v1/auth/refresh", authHTTPHandler.RefreshToken)
	e.POST("/api/v1/auth/logout", authHTTPHandler.Logout, middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	e.GET("/api/v1/auth/me", authHTTPHandler.GetCurrentUser, middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))

	rooms := e.Group("/api/v1/rooms", middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	rooms.POST("", chatHTTPHandler.CreateRoom)
	rooms.GET("", chatHTTPHandler.ListRooms)
	rooms.GET("/:id", chatHTTPHandler.GetRoom)
	rooms.PUT("/:id", chatHTTPHandler.UpdateRoom)
	rooms.DELETE("/:id", chatHTTPHandler.DeleteRoom)
	rooms.POST("/:id/join", chatHTTPHandler.JoinRoom)
	rooms.POST("/:id/leave", chatHTTPHandler.LeaveRoom)
	rooms.GET("/:id/members", chatHTTPHandler.GetRoomMembers)

	messages := e.Group("/api/v1/rooms/:id/messages", middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	messages.POST("", chatHTTPHandler.SendMessage)
	messages.GET("", chatHTTPHandler.GetMessageHistory)

	e.GET("/api/v1/rooms/:id/events", sseHandler.HandleSSE, middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))

	addr := fmt.Sprintf("%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
	logger.Info().Str("addr", addr).Msg("Starting HTTP server")

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server")
}
