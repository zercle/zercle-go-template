package handler

import (
	"context"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
)

// Handler implements task_entity.Handler interface for HTTP transport.
type handler struct {
	usecase task_entity.Usecase
	logger  *logging.Logger
}

// NewHandler creates a new task handler with the given injector.
// The handler implements the task_entity.Handler interface for handling
// HTTP requests related to task operations.
// Uses samber/do/v2 for dependency injection.
func NewHandler(i do.Injector) (task_entity.Handler, error) {
	uc := do.MustInvoke[task_entity.Usecase](i)
	return &handler{
		usecase: uc,
		logger:  logging.FromContext(context.TODO()),
	}, nil
}

// Compile-time interface check - Handler implements task_entity.Handler
var _ task_entity.Handler = (*handler)(nil)

// getLogger returns the logger with request context if available.
func (h *handler) getLogger(c *echo.Context) *logging.Logger {
	return h.logger.WithContext(c.Request().Context())
}
