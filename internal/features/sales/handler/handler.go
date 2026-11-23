package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

const timeFormat = time.RFC3339

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
	var req apiv1.PurchaseRequest
	if err := c.Bind(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(&req); err != nil {
		status, body := apperrors.HTTPError(apperrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	record, err := h.service.Purchase(c.Request().Context(), req.MachineID, req.ProductID, req.Coins)
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusCreated, newPurchaseResponse(record))
}

// newPurchaseResponse maps a domain purchase record to its wire form.
func newPurchaseResponse(record *domain.PurchaseRecord) apiv1.PurchaseResponse {
	return apiv1.PurchaseResponse{
		ID:                 record.ID.String(),
		MachineID:          record.MachineID.String(),
		ProductID:          record.ProductID.String(),
		PriceCents:         record.PriceCents,
		TotalInsertedCents: record.TotalInsertedCents,
		ChangeCents:        record.ChangeCents,
		ChangeCoins:        record.ChangeCoins,
		PurchasedAt:        record.PurchasedAt.Format(timeFormat),
	}
}
