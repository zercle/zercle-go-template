package handler

import (
	"context"

	"github.com/labstack/echo/v5"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
)

// Handler implements user_entity.Handler interface for HTTP transport.
type Handler struct {
	usecase user_entity.Usecase
	logger  *logging.Logger
}

// NewHandler creates a new user handler with the given usecase dependency.
// The handler implements the user_entity.Handler interface for handling
// HTTP requests related to user operations.
func NewHandler(usecase user_entity.Usecase) user_entity.Handler {
	return &Handler{
		usecase: usecase,
		logger:  logging.FromContext(context.TODO()),
	}
}

// NewHandlerWithLogger creates a new user handler with the given usecase
// and logger dependencies.
func NewHandlerWithLogger(usecase user_entity.Usecase, logger *logging.Logger) user_entity.Handler {
	return &Handler{
		usecase: usecase,
		logger:  logger,
	}
}

// Compile-time interface check - Handler implements user_entity.Handler
var _ user_entity.Handler = (*Handler)(nil)

// getLogger returns the logger with request context if available.
func (h *Handler) getLogger(c *echo.Context) *logging.Logger {
	return h.logger.WithContext(c.Request().Context())
}
