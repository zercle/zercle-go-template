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
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
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
	httpStopped     chan struct{}
	httpStartErr    error
	grpcServer      *grpc.Server
	grpcListener    net.Listener
	injector        do.Injector
	startMu         sync.Mutex
	httpStarted     chan struct{}
}

// NewApplication builds the runtime orchestrator from a populated DI
// container. It reads optional shutdown callbacks from the container and will
// look up infrastructure dependencies at runtime inside Run.
func NewApplication(injector do.Injector, cfg *config.Config, logger *zerolog.Logger) *Application {
	return &Application{
		cfg:         cfg,
		logger:      logger,
		injector:    injector,
		httpStarted: make(chan struct{}),
		httpStopped: make(chan struct{}),
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
		defer close(a.httpStopped)
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
			a.startMu.Lock()
			if a.httpListener == nil {
				a.httpStartErr = err
				close(a.httpStarted)
			}
			a.startMu.Unlock()
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

	var grpcErr error
	a.grpcServer, grpcErr = do.Invoke[*grpc.Server](a.injector)
	if grpcErr != nil {
		_ = listener.Close()
		return fmt.Errorf("resolve grpc server: %w", grpcErr)
	}
	a.grpcListener = listener

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

// runHTTPServer blocks until the HTTP server stops. If the server failed to
// bind (so the listener address was never produced), the start error is
// returned so Run can surface it instead of blocking forever.
func (a *Application) runHTTPServer() error {
	<-a.httpStarted
	if a.httpStartErr != nil {
		return a.httpStartErr
	}
	<-a.httpStartCtx.Done()
	return nil
}

// shutdown performs the ordered graceful shutdown sequence using a fresh,
// timeout-bounded context derived from ctx so an already-cancelled signal
// context does not starve the per-component shutdown calls.
func (a *Application) shutdown(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.App.ShutdownTimeout)
	defer cancel()

	if err := a.shutdownHTTP(shutdownCtx); err != nil {
		a.logger.Error().Err(err).Msg("http shutdown error")
	}

	done := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-shutdownCtx.Done():
		a.grpcServer.Stop()
	}

	if pool, ok := a.invokePool(); ok {
		pool.Close()
	}

	if client, ok := a.invokeValkey(); ok {
		client.Close()
	}

	if tp, ok := a.invokeTracerProvider(); ok {
		if err := tp.Shutdown(shutdownCtx); err != nil {
			a.logger.Error().Err(err).Msg("trace shutdown error")
		}
	}

	if mp, ok := a.invokeMeterProvider(); ok {
		if err := mp.Shutdown(shutdownCtx); err != nil {
			a.logger.Error().Err(err).Msg("meter shutdown error")
		}
	}

	a.logger.Info().Msg("shutdown complete")
}

// shutdownHTTP stops the echo HTTP server gracefully. It cancels the start
// context, which signals echo to begin its internal graceful drain, and then
// waits for the HTTP goroutine to actually finish (bounded by ctx) so that
// the database pool and Valkey are not closed underneath in-flight requests.
func (a *Application) shutdownHTTP(ctx context.Context) error {
	if a.httpStartCancel != nil {
		a.httpStartCancel()
	}
	select {
	case <-a.httpStopped:
	case <-ctx.Done():
		return fmt.Errorf("http shutdown timed out: %w", ctx.Err())
	}
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

// invokeTracerProvider looks up the OTel tracer provider from the DI container
// and reports whether it was found. A missing provider is treated as "not
// configured" and is skipped silently.
func (a *Application) invokeTracerProvider() (*trace.TracerProvider, bool) {
	tp, err := do.Invoke[*trace.TracerProvider](a.injector)
	if err == nil {
		return tp, true
	}
	if !errors.Is(err, do.ErrServiceNotFound) {
		a.logger.Warn().Err(err).Msg("optional tracer provider not available")
	}
	return nil, false
}

// invokeMeterProvider looks up the OTel meter provider from the DI container
// and reports whether it was found. A missing provider is treated as "not
// configured" and is skipped silently.
func (a *Application) invokeMeterProvider() (*metric.MeterProvider, bool) {
	mp, err := do.Invoke[*metric.MeterProvider](a.injector)
	if err == nil {
		return mp, true
	}
	if !errors.Is(err, do.ErrServiceNotFound) {
		a.logger.Warn().Err(err).Msg("optional meter provider not available")
	}
	return nil, false
}
