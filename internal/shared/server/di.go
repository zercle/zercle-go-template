// DI registration for shared servers and the application orchestrator.
package server

import (
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"google.golang.org/grpc"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register wires *echo.Echo, *grpc.Server, and the Application orchestrator
// into the DI container. It depends on config, logger, telemetry providers,
// and the health registry already being registered.
//
// Note: samber/do v2's Provide signature is `func Provide[T any](i Injector,
// provider Provider[T])` and returns no error. Any construction failure
// surfaces later via do.Invoke. We rely on the provider functions to
// surface their own errors via Invoke.
func Register(c do.Injector) error {
	do.Provide(c, func(i do.Injector) (*echo.Echo, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*zerolog.Logger](i)
		registry := do.MustInvoke[*telemetry.Registry](i)
		return NewHTTP(cfg, logger, registry), nil
	})

	do.Provide(c, func(i do.Injector) (*grpc.Server, error) {
		logger := do.MustInvoke[*zerolog.Logger](i)
		return NewGRPC(logger), nil
	})

	do.Provide(c, func(i do.Injector) (*Application, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*zerolog.Logger](i)
		return NewApplication(i, cfg, logger), nil
	})

	return nil
}
