//go:build unit

package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/server"
)

// TestApplication_RunReturnsOnBindFailure pins that Run surfaces a bind failure
// instead of blocking forever. The listener address is unresolvable, so the
// serve goroutine's Start fails before binding and closes the stopped channel;
// Run must observe it and return an error well before the context deadline.
func TestApplication_RunReturnsOnBindFailure(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		App:  config.AppConfig{ShutdownTimeout: 2 * time.Second},
		HTTP: config.HTTPConfig{Host: "invalid.invalid", Port: 0},
	}
	logger := zerolog.New(nil)

	injector := do.New()
	do.ProvideValue(injector, echo.New())
	app := server.NewApplication(injector, cfg, &logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()

	select {
	case err := <-done:
		require.Error(t, err, "bind failure must be surfaced, not swallowed")
	case <-time.After(5 * time.Second):
		t.Fatal("Run blocked after bind failure instead of returning")
	}
}
