package main

import (
	"fmt"
	"os"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
)

// @title           Zercle Go Template API
// @version         1.0
// @description     A production-ready RESTful API template built with Go Echo framework, featuring clean architecture, JWT authentication, and PostgreSQL database.

// @contact.name   API Support
// @contact.url    https://github.com/zercle/zercle-go-template
// @contact.email  support@zercle.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:3000
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	env := getEnv()

	cfg, err := config.Load("./configs/" + env + ".yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	log := logger.NewLogger(&cfg.Logging)
	log.Info("Starting zercle-go-template server", "env", cfg.Server.Env)

	application, err := app.NewApp(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize application", "error", err)
	}
	defer application.Close()

	if err := application.Start(); err != nil {
		log.Fatal("Application error", "error", err)
	}
}

func getEnv() string {
	env := os.Getenv("SERVER_ENV")
	if env == "" {
		env = "local"
	}
	return env
}
