package main

import (
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/infrastructure/loggers/zerolog"
	"github.com/zercle/zercle-go-template/internal/infrastructure/middlewares"

	"github.com/labstack/echo/v5"

	_ "github.com/zercle/zercle-go-template/docs"
)

// swaggerHandlerWrapper wraps the v4 swagger handler to work with v5
func swaggerHandlerWrapper(c echo.Context) error {
	// TODO: echo-swagger v1.5.2 does not support Echo v5 yet.
	// Once a v5-compatible version is released, update this code.
	// See: https://github.com/swaggo/echo-swagger/issues
	return echo.NewHTTPError(501, "swagger documentation not available - echo-swagger v5 support pending")
}

func main() {
	application, err := app.New()
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to initialize app")
		os.Exit(1)
	}
	defer application.Close()

	application.InitAuth()
	application.InitChat()

	e := echo.New()

	// Swagger endpoint temporarily disabled - echo-swagger v1.5.2 does not support Echo v5
	// e.GET("/swagger/*", swaggerHandlerWrapper())

	v1 := e.Group("/api/v1")

	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", application.AuthHandler.Register)
	authRoutes.POST("/login", application.AuthHandler.Login)
	authRoutes.POST("/refresh", application.AuthHandler.RefreshToken)

	authProtected := authRoutes.Group("", middlewares.AuthMiddleware([]byte(application.GetJWTSecret())))
	authProtected.GET("/me", application.AuthHandler.GetCurrentUser)
	authProtected.POST("/logout", application.AuthHandler.Logout)

	chatRoutes := v1.Group("/chat", middlewares.AuthMiddleware([]byte(application.GetJWTSecret())))
	chatRoutes.POST("/rooms", application.ChatHandler.CreateRoom)
	chatRoutes.GET("/rooms", application.ChatHandler.ListRooms)
	chatRoutes.GET("/rooms/:id", application.ChatHandler.GetRoom)
	chatRoutes.PUT("/rooms/:id", application.ChatHandler.UpdateRoom)
	chatRoutes.DELETE("/rooms/:id", application.ChatHandler.DeleteRoom)
	chatRoutes.POST("/rooms/:id/join", application.ChatHandler.JoinRoom)
	chatRoutes.POST("/rooms/:id/leave", application.ChatHandler.LeaveRoom)
	chatRoutes.GET("/rooms/:id/members", application.ChatHandler.GetRoomMembers)
	chatRoutes.POST("/rooms/:id/messages", application.ChatHandler.SendMessage)
	chatRoutes.GET("/rooms/:id/messages", application.ChatHandler.GetMessageHistory)

	serverConfig := application.GetServerConfig()

	go func() {
		zerolog.Info().Str("addr", serverConfig.HTTP.Addr()).Msg("HTTP server listening")
		if err := e.Start(serverConfig.HTTP.Addr()); err != nil {
			zerolog.Error().Err(err).Msg("HTTP server error")
		}
	}()

	zerolog.Info().Msg("Starting gRPC server")

	lis, err := net.Listen("tcp", serverConfig.GRPC.Addr())
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to listen")
		application.Close()
		os.Exit(1) //nolint:gocritic
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	zerolog.Info().Str("addr", lis.Addr().String()).Msg("gRPC server listening")

	if err := grpcServer.Serve(lis); err != nil {
		zerolog.Error().Err(err).Msg("gRPC server error")
	}
}
