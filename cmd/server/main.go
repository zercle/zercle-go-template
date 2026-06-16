// Command server is the composition root and runtime entry point. It loads
// config and delegates to package app for DI wiring, then starts and gracefully
// shuts down the HTTP and gRPC servers.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/config"
)

var (
	// Version is set at build time via -ldflags "-X main.Version=...".
	Version = "dev"
	// CommitSHA is set at build time via -ldflags "-X main.CommitSHA=...".
	CommitSHA = "unknown"
	// BuildTime is set at build time via -ldflags "-X main.BuildTime=...".
	BuildTime = "unknown"
)

func main() {
	os.Exit(run())
}

// run loads configuration and starts the application. It returns the process
// exit code so main can exit in one place, allowing defers in app.Run to run.
func run() (exitCode int) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		return 1
	}

	app.Version = Version
	app.CommitSHA = CommitSHA
	app.BuildTime = BuildTime

	if err := app.Run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "server stopped with error: %v\n", err)
		return 1
	}

	return 0
}
