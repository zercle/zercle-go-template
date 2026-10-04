package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/sales/contract"
	"github.com/zercle/zercle-go-template/internal/features/sales/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Handler exposes the sales usecase service over HTTP.
type Handler struct {
	service usecase.Service
}

// New returns an HTTP handler for the sales feature.
func New(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts the sales routes on the provided echo group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/purchases", h.Purchase)
}

// NOTE: Echo v5 changed echo.Context from an interface (v4) to a struct, and
// echo.HandlerFunc is now `func(c *Context) error`. Handlers therefore take
// *echo.Context — this is correct for v5, not a mistake.

// Purchase handles POST /purchases.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Purchase(c *echo.Context) error {
	var req contract.PurchaseRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	resp, err := h.service.Purchase(c.Request().Context(), &req)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusCreated, resp)
}
