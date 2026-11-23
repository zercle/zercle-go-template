package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

// Handler exposes the reporting usecase service over HTTP. It owns the wire
// mapping: it binds and validates the pkg/api/v1 request, translates it into
// the usecase's domain inputs, and maps the domain result back into the
// pkg/api/v1 response.
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
	var req apiv1.SummaryRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	summary, err := h.service.Summary(c.Request().Context(), req.TopMachines)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, newSummaryResponse(summary))
}

// newSummaryResponse maps a domain Summary into the published wire response.
func newSummaryResponse(s *domain.Summary) *apiv1.SummaryResponse {
	resp := &apiv1.SummaryResponse{
		Catalog: apiv1.CatalogStats{
			ProductCount: s.Overview.ProductCount,
			TotalStock:   s.Overview.TotalStock,
		},
		Machines: apiv1.MachineStats{
			MachineCount:       s.Overview.MachineCount,
			TotalCoinBankCents: s.Overview.TotalBankCents,
		},
		Sales: apiv1.SalesStats{
			PurchaseCount: s.Overview.PurchaseCount,
			RevenueCents:  s.Overview.RevenueCents,
		},
		TopMachines: make([]apiv1.MachineSales, len(s.TopMachines)),
	}
	for i := range s.TopMachines {
		resp.TopMachines[i] = apiv1.MachineSales{
			MachineID:     s.TopMachines[i].MachineID.String(),
			Label:         s.TopMachines[i].Label,
			PurchaseCount: s.TopMachines[i].PurchaseCount,
			RevenueCents:  s.TopMachines[i].RevenueCents,
		}
	}
	return resp
}
