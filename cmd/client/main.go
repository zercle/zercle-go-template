package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/infrastructure/loggers/zerolog"
	"github.com/zercle/zercle-go-template/internal/infrastructure/middlewares"
)

func main() {
	app, err := app.New()
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to initialize app")
		os.Exit(1)
	}
	defer app.Close()

	app.InitAuth()
	app.InitChat()

	e := echo.New()

	e.Use(echomiddleware.RequestLogger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")

	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", app.AuthHandler.Register)
	authRoutes.POST("/login", app.AuthHandler.Login)
	authRoutes.POST("/refresh", app.AuthHandler.RefreshToken)

	authProtected := authRoutes.Group("", middlewares.AuthMiddleware([]byte(app.GetJWTSecret())))
	authProtected.GET("/me", app.AuthHandler.GetCurrentUser)
	authProtected.POST("/logout", app.AuthHandler.Logout)

	chatRoutes := v1.Group("/chat", middlewares.AuthMiddleware([]byte(app.GetJWTSecret())))
	chatRoutes.POST("/rooms", app.ChatHandler.CreateRoom)
	chatRoutes.GET("/rooms", app.ChatHandler.ListRooms)
	chatRoutes.GET("/rooms/:id", app.ChatHandler.GetRoom)
	chatRoutes.PUT("/rooms/:id", app.ChatHandler.UpdateRoom)
	chatRoutes.DELETE("/rooms/:id", app.ChatHandler.DeleteRoom)
	chatRoutes.POST("/rooms/:id/join", app.ChatHandler.JoinRoom)
	chatRoutes.POST("/rooms/:id/leave", app.ChatHandler.LeaveRoom)
	chatRoutes.GET("/rooms/:id/members", app.ChatHandler.GetRoomMembers)
	chatRoutes.POST("/rooms/:id/messages", app.ChatHandler.SendMessage)
	chatRoutes.GET("/rooms/:id/messages", app.ChatHandler.GetMessageHistory)

	addr := fmt.Sprintf("%s:%d", app.GetServerConfig().HTTP.Host, app.GetServerConfig().HTTP.Port)
	zerolog.Info().Str("addr", addr).Msg("Starting HTTP server")

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			zerolog.Error().Err(err).Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zerolog.Info().Msg("Shutting down server")
}
