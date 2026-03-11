package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/zercle/zercle-go-template/api/pb"
	authgrpc "github.com/zercle/zercle-go-template/internal/features/auth/handler/grpc"
	authhttp "github.com/zercle/zercle-go-template/internal/features/auth/handler/http"
	authservice "github.com/zercle/zercle-go-template/internal/features/auth/service"
	chatgrpc "github.com/zercle/zercle-go-template/internal/features/chat/handler/grpc"
	chathttp "github.com/zercle/zercle-go-template/internal/features/chat/handler/http"
	chatservice "github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"

	"github.com/labstack/echo/v5"
	echoswagger "github.com/swaggo/echo-swagger"

	_ "github.com/zercle/zercle-go-template/docs" // swagger docs
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

	chatSvc := chatservice.NewChatService(roomRepo, messageRepo)

	authServer := authgrpc.NewAuthServer(authSvc)
	chatServer := chatgrpc.NewChatServer(chatSvc)

	authHttpHandler := authhttp.NewAuthHandler(authSvc)
	chatHttpHandler := chathttp.NewChatHandler(chatSvc)

	// Start HTTP server with Swagger
	e := echo.New()

	// Swagger endpoint - available at /swagger/index.html
	e.GET("/swagger/*", echoswagger.WrapHandler)

	// Setup API routes
	v1 := e.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	auth.POST("/register", authHttpHandler.Register)
	auth.POST("/login", authHttpHandler.Login)
	auth.POST("/refresh", authHttpHandler.RefreshToken)

	// Protected auth routes
	authProtected := auth.Group("", middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	authProtected.GET("/me", authHttpHandler.GetCurrentUser)
	authProtected.POST("/logout", authHttpHandler.Logout)

	// Chat routes (all protected)
	chat := v1.Group("/chat", middleware.AuthMiddleware([]byte(cfg.Auth.JWTSecret)))
	chat.POST("/rooms", chatHttpHandler.CreateRoom)
	chat.GET("/rooms", chatHttpHandler.ListRooms)
	chat.GET("/rooms/:id", chatHttpHandler.GetRoom)
	chat.PUT("/rooms/:id", chatHttpHandler.UpdateRoom)
	chat.DELETE("/rooms/:id", chatHttpHandler.DeleteRoom)
	chat.POST("/rooms/:id/join", chatHttpHandler.JoinRoom)
	chat.POST("/rooms/:id/leave", chatHttpHandler.LeaveRoom)
	chat.GET("/rooms/:id/members", chatHttpHandler.GetRoomMembers)
	chat.POST("/rooms/:id/messages", chatHttpHandler.SendMessage)
	chat.GET("/rooms/:id/messages", chatHttpHandler.GetMessageHistory)

	// Start HTTP server
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

	RegisterServers(grpcServer, authServer, chatServer)

	logger.Info().Str("addr", lis.Addr().String()).Msg("gRPC server listening")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error().Err(err).Msg("gRPC server error")
		os.Exit(1)
	}
}

func RegisterServers(grpcServer *grpc.Server, authServer *authgrpc.AuthServer, chatServer *chatgrpc.ChatServer) {
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterChatServiceServer(grpcServer, chatServer)
}
