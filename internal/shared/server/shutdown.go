// Application orchestrates starting and gracefully shutting down HTTP, gRPC,
// database, Valkey, and telemetry providers.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/valkey-io/valkey-go"
	"google.golang.org/grpc"

	"github.com/zercle/zercle-go-template/internal/config"
)

// Application holds the runtime components required to start and stop the
// service. It is constructed from a populated DI container and keeps the
// orchestration logic separate from the wiring code.
type Application struct {
	cfg             *config.Config
	logger          *zerolog.Logger
	httpServer      *echo.Echo
	httpListener    net.Addr
	httpStartCtx    context.Context
	httpStartCancel context.CancelFunc
	grpcServer      *grpc.Server
	grpcListener    net.Listener
	injector        do.Injector
	traceShutdown   func(context.Context) error
	meterShutdown   func(context.Context) error
	shutdownTimeout int64
	startMu         sync.Mutex
	httpStarted     chan struct{}
}

// NewApplication builds the runtime orchestrator from a populated DI
// container. It reads optional shutdown callbacks from the container and will
// look up infrastructure dependencies at runtime inside Run.
func NewApplication(injector do.Injector, cfg *config.Config, logger *zerolog.Logger) *Application {
	return &Application{
		cfg:             cfg,
		logger:          logger,
		injector:        injector,
		shutdownTimeout: int64(cfg.App.ShutdownTimeout.Seconds()),
		httpStarted:     make(chan struct{}),
	}
}

// Echo returns the underlying HTTP server. Tests and callers outside the
// package may use it to mount routes or drive httptest servers. The server is
// resolved lazily on the first call so that routes are available before Run().
func (a *Application) Echo() *echo.Echo {
	a.startMu.Lock()
	defer a.startMu.Unlock()

	if a.httpServer == nil {
		var err error
		a.httpServer, err = do.Invoke[*echo.Echo](a.injector)
		if err != nil {
			a.logger.Error().Err(err).Msg("resolve http server")
			return nil
		}
	}
	return a.httpServer
}

// HTTPAddr returns the bound listener address after Run() has started the HTTP
// server. It returns an empty string before the server has started.
func (a *Application) HTTPAddr() string {
	a.startMu.Lock()
	defer a.startMu.Unlock()

	if a.httpListener == nil {
		return ""
	}
	return a.httpListener.String()
}

// HasHTTPStarted returns a channel that is closed once the HTTP listener has
// been bound. Callers can use it to wait for the server to be ready.
func (a *Application) HasHTTPStarted() <-chan struct{} {
	return a.httpStarted
}

// Logger returns the application logger.
func (a *Application) Logger() *zerolog.Logger {
	return a.logger
}

// Run starts the HTTP and gRPC servers and blocks until a signal or a server
// error occurs, then performs an ordered graceful shutdown.
func (a *Application) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.StartHTTP(ctx); err != nil {
		return fmt.Errorf("start http: %w", err)
	}

	if err := a.startGRPC(); err != nil {
		return fmt.Errorf("start grpc: %w", err)
	}

	var runErr error
	select {
	case <-ctx.Done():
		a.logger.Info().Msg("shutdown signal received")
	case err := <-a.serverErrorChannel():
		a.logger.Error().Err(err).Msg("server error")
		runErr = err
	}

	a.shutdown(ctx)

	return runErr
}

// StartHTTP eagerly resolves and starts the HTTP server. It is safe to call
// multiple times; subsequent calls are no-ops. The gRPC server is still started
// inside Run().
func (a *Application) StartHTTP(ctx context.Context) error {
	a.startMu.Lock()
	if a.httpStartCtx != nil {
		a.startMu.Unlock()
		return nil
	}

	if a.httpServer == nil {
		var err error
		a.httpServer, err = do.Invoke[*echo.Echo](a.injector)
		if err != nil {
			a.startMu.Unlock()
			return fmt.Errorf("resolve http server: %w", err)
		}
	}
	a.httpStartCtx, a.httpStartCancel = context.WithCancel(ctx)
	a.startMu.Unlock()

	go func() {
		sc := echo.StartConfig{
			Address:         a.cfg.HTTPAddr(),
			HideBanner:      true,
			HidePort:        true,
			ListenerNetwork: "tcp",
			GracefulTimeout: a.cfg.App.ShutdownTimeout,
			ListenerAddrFunc: func(addr net.Addr) {
				a.startMu.Lock()
				a.httpListener = addr
				a.startMu.Unlock()
				close(a.httpStarted)
			},
		}
		if err := sc.Start(a.httpStartCtx, a.httpServer); err != nil {
			a.logger.Error().Err(err).Msg("http server stopped")
		}
	}()

	return nil
}

// startGRPC resolves the shared gRPC server from the DI container and binds the
// listener. HTTP must be started before or alongside this call.
func (a *Application) startGRPC() error {
	listener, err := net.Listen("tcp", a.cfg.GRPCAddr())
	if err != nil {
		return fmt.Errorf("listen grpc %s: %w", a.cfg.GRPCAddr(), err)
	}
	a.grpcListener = listener

	var grpcErr error
	a.grpcServer, grpcErr = do.Invoke[*grpc.Server](a.injector)
	if grpcErr != nil {
		return fmt.Errorf("resolve grpc server: %w", grpcErr)
	}

	return nil
}

// serverErrorChannel launches both servers and returns a channel that receives
// the first fatal error from either.
func (a *Application) serverErrorChannel() <-chan error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- a.runHTTPServer()
	}()
	go func() {
		errCh <- a.grpcServer.Serve(a.grpcListener)
	}()

	return errCh
}

// runHTTPServer blocks until the HTTP server stops.
func (a *Application) runHTTPServer() error {
	<-a.httpStarted
	<-a.httpStartCtx.Done()
	return nil
}

// shutdown performs the ordered graceful shutdown sequence.
func (a *Application) shutdown(ctx context.Context) {
	if err := a.shutdownHTTP(ctx); err != nil {
		a.logger.Error().Err(err).Msg("http shutdown error")
	}

	done := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		a.grpcServer.Stop()
	}

	if pool, ok := a.invokePool(); ok {
		pool.Close()
	}

	if client, ok := a.invokeValkey(); ok {
		client.Close()
	}

	if a.traceShutdown != nil {
		if err := a.traceShutdown(ctx); err != nil {
			a.logger.Error().Err(err).Msg("trace shutdown error")
		}
	}

	if a.meterShutdown != nil {
		if err := a.meterShutdown(ctx); err != nil {
			a.logger.Error().Err(err).Msg("meter shutdown error")
		}
	}

	a.logger.Info().Msg("shutdown complete")
}

// shutdownHTTP stops the echo HTTP server gracefully. Echo v5's StartConfig
// handles graceful shutdown internally when the context passed to Start is
// cancelled, because the listener created inside Start belongs to that
// goroutine. We cancel the context here; the listener rebind fallback is only
// used when the listener address was supplied externally.
func (a *Application) shutdownHTTP(ctx context.Context) error {
	if a.httpStartCancel != nil {
		a.httpStartCancel()
	}
	_ = ctx
	return nil
}

// invokePool looks up the pgx pool from the DI container and reports whether
// it was found. A missing provider is treated as "not configured" and is
// skipped silently.
func (a *Application) invokePool() (*pgxpool.Pool, bool) {
	pool, err := do.Invoke[*pgxpool.Pool](a.injector)
	if err == nil {
		return pool, true
	}
	if !errors.Is(err, do.ErrServiceNotFound) {
		a.logger.Warn().Err(err).Msg("optional pgx pool not available")
	}
	return nil, false
}

// invokeValkey looks up the valkey client from the DI container and reports
// whether it was found. A missing provider is treated as "not configured" and
// is skipped silently.
func (a *Application) invokeValkey() (valkey.Client, bool) {
	client, err := do.Invoke[valkey.Client](a.injector)
	if err == nil {
		return client, true
	}
	if !errors.Is(err, do.ErrServiceNotFound) {
		a.logger.Warn().Err(err).Msg("optional valkey client not available")
	}
	return nil, false
}
