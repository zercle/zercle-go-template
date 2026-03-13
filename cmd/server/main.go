package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/zercle/zercle-go-template/api/pb"
	"github.com/zercle/zercle-go-template/internal/config"
	authRepo "github.com/zercle/zercle-go-template/internal/features/auth"
	authHandler "github.com/zercle/zercle-go-template/internal/features/auth/handler"
	authService "github.com/zercle/zercle-go-template/internal/features/auth/service"
	chatRepo "github.com/zercle/zercle-go-template/internal/features/chat"
	chatGrpc "github.com/zercle/zercle-go-template/internal/features/chat/handler"
	chatHttp "github.com/zercle/zercle-go-template/internal/features/chat/handler"
	chatService "github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/logger"
	"github.com/zercle/zercle-go-template/internal/middleware"
	"github.com/zercle/zercle-go-template/internal/postgres"

	"github.com/labstack/echo/v5"
	echoswagger "github.com/swaggo/echo-swagger"

	_ "github.com/zercle/zercle-go-template/docs"
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

	userRepo := authRepo.NewUserRepository(db)
	sessionRepo := authRepo.NewSessionRepository(db)
	roomRepo := chatRepo.NewRoomRepository(db)
	messageRepo := chatRepo.NewMessageRepository(db)

	authSvc := authService.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpiry,
		cfg.Auth.RefreshExpiry,
	)

	chatSvc := chatService.NewChatService(roomRepo, messageRepo)

	authServer := authHandler.NewAuthServer(authSvc)
	chatServer := chatGrpc.NewChatServer(chatSvc)

	authHttpHandler := authHandler.NewAuthHandler(authSvc)
	chatHttpHandler := chatHttp.NewChatHandler(chatSvc)

	e := echo.New()

	e.GET("/swagger/*", echoswagger.WrapHandler)

	v1 := e.Group("/api/v1")

	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", authHttpHandler.Register)
	authRoutes.POST("/login", authHttpHandler.Login)
	authRoutes.POST("/refresh", authHttpHandler.RefreshToken)

	authProtected := authRoutes.Group("", middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	authProtected.GET("/me", authHttpHandler.GetCurrentUser)
	authProtected.POST("/logout", authHttpHandler.Logout)

	chatRoutes := v1.Group("/chat", middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	chatRoutes.POST("/rooms", chatHttpHandler.CreateRoom)
	chatRoutes.GET("/rooms", chatHttpHandler.ListRooms)
	chatRoutes.GET("/rooms/:id", chatHttpHandler.GetRoom)
	chatRoutes.PUT("/rooms/:id", chatHttpHandler.UpdateRoom)
	chatRoutes.DELETE("/rooms/:id", chatHttpHandler.DeleteRoom)
	chatRoutes.POST("/rooms/:id/join", chatHttpHandler.JoinRoom)
	chatRoutes.POST("/rooms/:id/leave", chatHttpHandler.LeaveRoom)
	chatRoutes.GET("/rooms/:id/members", chatHttpHandler.GetRoomMembers)
	chatRoutes.POST("/rooms/:id/messages", chatHttpHandler.SendMessage)
	chatRoutes.GET("/rooms/:id/messages", chatHttpHandler.GetMessageHistory)

	go func() {
		logger.Info().Str("addr", cfg.Server.HTTP.Addr()).Msg("HTTP server listening")
		if err := e.Start(cfg.Server.HTTP.Addr()); err != nil {
			logger.Error().Err(err).Msg("HTTP server error")
		}
	}()

	logger.Info().Msg("Starting gRPC server")

	lis, err := net.Listen("tcp", cfg.Server.GRPC.Addr())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to listen")
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterChatServiceServer(grpcServer, chatServer)

	logger.Info().Str("addr", lis.Addr().String()).Msg("gRPC server listening")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error().Err(err).Msg("gRPC server error")
		os.Exit(1)
	}
}
