package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/zercle/zercle-go-template/api/pb"
	authgrpc "github.com/zercle/zercle-go-template/internal/features/auth/handler/grpc"
	authservice "github.com/zercle/zercle-go-template/internal/features/auth/service"
	chatgrpc "github.com/zercle/zercle-go-template/internal/features/chat/handler/grpc"
	chatservice "github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
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
