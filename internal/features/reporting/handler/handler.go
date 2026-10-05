package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/reporting/contract"
	"github.com/zercle/zercle-go-template/internal/features/reporting/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Handler exposes the reporting usecase service over HTTP.
type Handler struct {
	service usecase.Service
}

// New returns an HTTP handler for the reporting feature.
func New(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts the reporting routes on the provided echo group.
func (h *Handler) Register(g *echo.Group) {
	g.GET("/reports/summary", h.Summary)
}

// NOTE: Echo v5 changed echo.Context from an interface (v4) to a struct, and
// echo.HandlerFunc is now `func(c *Context) error`. Handlers therefore take
// *echo.Context — this is correct for v5, not a mistake.

// Summary handles GET /reports/summary.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Summary(c *echo.Context) error {
	var req contract.SummaryRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	resp, err := h.service.Summary(c.Request().Context(), &req)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, resp)
}
