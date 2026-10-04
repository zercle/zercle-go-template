package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Handler exposes the machines usecase service over HTTP.
type Handler struct {
	service usecase.Service
}

// New returns an HTTP handler for the machines feature.
func New(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts the machines routes on the provided echo group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/machines", h.Create)
	g.GET("/machines", h.List)
	g.GET("/machines/:id", h.Get)
	g.POST("/machines/:id/bank", h.RestockBank)
}

// NOTE: Echo v5 changed echo.Context from an interface (v4) to a struct, and
// echo.HandlerFunc is now `func(c *Context) error`. Handlers therefore take
// *echo.Context — this is correct for v5, not a mistake.

// Create handles POST /machines.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Create(c *echo.Context) error {
	var req contract.CreateMachineRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	resp, err := h.service.Create(c.Request().Context(), &req)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusCreated, resp)
}

// Get handles GET /machines/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Get(c *echo.Context) error {
	resp, err := h.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, resp)
}

// List handles GET /machines.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) List(c *echo.Context) error {
	var req contract.ListMachinesRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	resp, err := h.service.List(c.Request().Context(), &req)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, resp)
}

// RestockBank handles POST /machines/:id/bank.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) RestockBank(c *echo.Context) error {
	var req contract.RestockBankRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	resp, err := h.service.RestockBank(c.Request().Context(), c.Param("id"), &req)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, resp)
}
