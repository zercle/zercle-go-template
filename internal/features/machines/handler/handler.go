package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

const timeFormat = time.RFC3339

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
	var req apiv1.CreateMachineRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	machine, err := h.service.Create(c.Request().Context(), req.Label, req.InitialCoins)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusCreated, toMachineResponse(machine))
}

// Get handles GET /machines/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Get(c *echo.Context) error {
	machine, err := h.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, toMachineResponse(machine))
}

// List handles GET /machines.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) List(c *echo.Context) error {
	var req apiv1.ListMachinesRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	machines, err := h.service.List(c.Request().Context(), req.Limit, req.Offset)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	resp := &apiv1.ListMachinesResponse{Machines: make([]apiv1.MachineResponse, len(machines))}
	for i := range machines {
		resp.Machines[i] = toMachineResponse(&machines[i])
	}
	return c.JSON(http.StatusOK, resp)
}

// RestockBank handles POST /machines/:id/bank.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) RestockBank(c *echo.Context) error {
	var req apiv1.RestockBankRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	machine, err := h.service.RestockBank(c.Request().Context(), c.Param("id"), req.Coins)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, toMachineResponse(machine))
}

func toMachineResponse(m *domain.Machine) apiv1.MachineResponse {
	if m == nil {
		return apiv1.MachineResponse{}
	}
	bank := m.CoinBank
	if bank == nil {
		bank = domain.CoinBank{}
	}
	return apiv1.MachineResponse{
		ID:        m.ID.String(),
		Label:     m.Label,
		CoinBank:  bank,
		CreatedAt: m.CreatedAt.Format(timeFormat),
		UpdatedAt: m.UpdatedAt.Format(timeFormat),
	}
}
