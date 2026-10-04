// Application orchestrates starting and gracefully shutting down the HTTP
// server, database (gorm), Valkey, and telemetry providers.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
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
	injector        do.Injector
	startMu         sync.Mutex
	httpStarted     chan struct{}
}

// NewApplication builds the runtime orchestrator from a populated DI
// container. It resolves nothing eagerly: infrastructure dependencies are
// looked up inside Run and Shutdown, so a partially-wired container still
// yields a usable Application.
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

// Run starts the HTTP server and blocks until the context is cancelled or the
// server stops on its own, then performs an ordered graceful shutdown. It never
// installs process-level signal handlers; the caller supplies a context that is
// cancelled on SIGINT/SIGTERM (see cmd/server).
func (a *Application) Run(ctx context.Context) error {
	if err := a.StartHTTP(ctx); err != nil {
		return fmt.Errorf("start http: %w", err)
	}

	var runErr error
	select {
	case <-ctx.Done():
		a.logger.Info().Msg("shutdown signal received")
	case <-a.httpStopped:
		// The server stopped without the context being cancelled: a bind/serve
		// failure (or an externally closed listener). Surface the recorded error
		// instead of blocking forever.
		if err := a.httpStartError(); err != nil {
			a.logger.Error().Err(err).Msg("server error")
			runErr = err
		}
	}

	a.shutdown(ctx)

	// A signal-triggered shutdown and a simultaneous start failure can race the
	// select above; report a recorded error that was not observed there.
	if runErr == nil {
		runErr = a.httpStartError()
	}

	return runErr
}

// StartHTTP eagerly resolves and starts the HTTP server. It is safe to call
// multiple times; subsequent calls are no-ops.
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
			BeforeServeFunc: func(s *http.Server) error {
				s.ReadTimeout = a.cfg.HTTP.ReadTimeout
				s.WriteTimeout = a.cfg.HTTP.WriteTimeout
				s.IdleTimeout = a.cfg.HTTP.IdleTimeout
				return nil
			},
			ListenerAddrFunc: func(addr net.Addr) {
				a.startMu.Lock()
				a.httpListener = addr
				a.startMu.Unlock()
				close(a.httpStarted)
			},
		}
		if err := sc.Start(a.httpStartCtx, a.httpServer); err != nil {
			a.logger.Error().Err(err).Msg("http server stopped")
			// A graceful stop (context cancelled) surfaces as ErrServerClosed or
			// context.Canceled; that is a normal shutdown, not a startup failure.
			// Anything else — including a bind failure after ListenerAddrFunc
			// already closed httpStarted — is recorded so Run can surface it
			// instead of blocking until an external signal.
			if !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
				a.startMu.Lock()
				a.httpStartErr = err
				a.startMu.Unlock()
			}
		}
	}()

	return nil
}

// httpStartError returns the HTTP start/serve failure, if any. It reads under
// the start mutex so callers never race with the serve goroutine recording an
// error after the listener has already bound.
func (a *Application) httpStartError() error {
	a.startMu.Lock()
	defer a.startMu.Unlock()
	return a.httpStartErr
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

	if db, ok := a.invokeDB(); ok {
		a.closeDB(db)
	}

	if client, ok := a.invokeValkey(); ok {
		client.Close()
	}

	// The OTel tracer and meter providers implement samber/do's shutdown
	// interface, so the injector shuts them down itself (see app.Run's defer).
	// Shutting them down here as well would be a second, redundant call that the
	// Prometheus reader rejects with "reader is shutdown".

	a.logger.Info().Msg("shutdown complete")
}

// shutdownHTTP stops the echo HTTP server gracefully. It cancels the start
// context, which signals echo to begin its internal graceful drain, and then
// waits for the HTTP goroutine to actually finish (bounded by ctx) so that
// the gorm db and Valkey are not closed underneath in-flight requests.
func (a *Application) shutdownHTTP(ctx context.Context) error {
	if a.httpStartCancel == nil {
		return nil
	}
	a.httpStartCancel()
	select {
	case <-a.httpStopped:
	case <-ctx.Done():
		return fmt.Errorf("http shutdown timed out: %w", ctx.Err())
	}
	return nil
}

// invokeDB looks up the *gorm.DB from the DI container and reports whether
// it was found. A missing provider is treated as "not configured" and is
// skipped silently.
func (a *Application) invokeDB() (*gorm.DB, bool) {
	db, err := do.Invoke[*gorm.DB](a.injector)
	if err == nil {
		return db, true
	}
	if !errors.Is(err, do.ErrServiceNotFound) {
		a.logger.Warn().Err(err).Msg("optional gorm db not available")
	}
	return nil, false
}

// closeDB closes the underlying *sql.DB of a *gorm.DB and logs any error.
// *gorm.DB itself does not expose Close; the database/sql handle does.
func (a *Application) closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		a.logger.Warn().Err(err).Msg("gorm db sql handle unavailable")
		return
	}
	if err := sqlDB.Close(); err != nil {
		a.logger.Warn().Err(err).Msg("gorm db close error")
	}
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
