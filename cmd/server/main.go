package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	logger, err := telemetry.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	if err := run(&cfg, logger); err != nil {
		logger.Error("server error", "error", err)
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

	chatSvc := chatservice.NewChatService(roomRepo, messageRepo)

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
		logger.Info("HTTP server listening", "addr", cfg.ServerAddr())
		if err := e.Start(cfg.ServerAddr()); err != nil {
			logger.Error("HTTP server error", "error", err)
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

	logger.Info("gRPC server listening", "addr", grpcLis.Addr().String())

	if err := grpcServer.Serve(grpcLis); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	return nil
}
