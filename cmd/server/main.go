package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/zercle/zercle-go-template/api/pb"
	"github.com/zercle/zercle-go-template/internal/config"
	authgrpc "github.com/zercle/zercle-go-template/internal/features/auth/handler/grpc"
	authhttp "github.com/zercle/zercle-go-template/internal/features/auth/handler/http"
	authservice "github.com/zercle/zercle-go-template/internal/features/auth/service"
	chatgrpc "github.com/zercle/zercle-go-template/internal/features/chat/handler/grpc"
	chathttp "github.com/zercle/zercle-go-template/internal/features/chat/handler/http"
	chatservice "github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"

	"github.com/labstack/echo/v5"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

func main() {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	logger := telemetry.New(cfg.LogLevel, cfg.LogFormat)

	// Set the global zerolog level for packages that use the global logger (e.g. uuidgen).
	zerolog.SetGlobalLevel(logger.GetLevel())

	if err := run(&cfg, logger); err != nil {
		logger.Err(err).Msg("server error")
		os.Exit(1)
	}
}

func run(cfg *config.Config, logger *telemetry.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	roomRepo := postgres.NewRoomRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	authSvc := authservice.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.AuthAccessTokenSecret,
		cfg.AuthAccessTokenTTL,
		cfg.AuthRefreshTokenTTL,
	)

	chatSvc := chatservice.NewChatService(roomRepo, messageRepo, &logger.Logger)

	authServer := authgrpc.NewAuthServer(authSvc)
	chatServer := chatgrpc.NewChatServer(chatSvc)

	authHTTPHandler := authhttp.NewAuthHandler(authSvc)
	chatHTTPHandler := chathttp.NewChatHandler(chatSvc)

	e := echo.New()

	v1 := e.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.POST("/register", authHTTPHandler.Register)
	auth.POST("/login", authHTTPHandler.Login)
	auth.POST("/refresh", authHTTPHandler.RefreshToken)

	authProtected := auth.Group("", middleware.AuthMiddleware([]byte(cfg.AuthAccessTokenSecret)))
	authProtected.GET("/me", authHTTPHandler.GetCurrentUser)
	authProtected.POST("/logout", authHTTPHandler.Logout)

	chat := v1.Group("/chat", middleware.AuthMiddleware([]byte(cfg.AuthAccessTokenSecret)))
	chat.POST("/rooms", chatHTTPHandler.CreateRoom)
	chat.GET("/rooms", chatHTTPHandler.ListRooms)
	chat.GET("/rooms/:id", chatHTTPHandler.GetRoom)
	chat.PUT("/rooms/:id", chatHTTPHandler.UpdateRoom)
	chat.DELETE("/rooms/:id", chatHTTPHandler.DeleteRoom)
	chat.POST("/rooms/:id/join", chatHTTPHandler.JoinRoom)
	chat.POST("/rooms/:id/leave", chatHTTPHandler.LeaveRoom)
	chat.GET("/rooms/:id/members", chatHTTPHandler.GetRoomMembers)
	chat.POST("/rooms/:id/messages", chatHTTPHandler.SendMessage)
	chat.GET("/rooms/:id/messages", chatHTTPHandler.GetMessageHistory)

	go func() {
		logger.Info().Str("addr", cfg.ServerAddr()).Msg("HTTP server listening")
		if err := e.Start(cfg.ServerAddr()); err != nil {
			logger.Err(err).Msg("HTTP server error")
		}
	}()

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		return fmt.Errorf("failed to listen for gRPC: %w", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterChatServiceServer(grpcServer, chatServer)

	logger.Info().Str("addr", grpcLis.Addr().String()).Msg("gRPC server listening")

	if err := grpcServer.Serve(grpcLis); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	return nil
}
