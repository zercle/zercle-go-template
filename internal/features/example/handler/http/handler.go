// STUB FEATURE — delete internal/features/example to start your project.

package httphandler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/dto"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// Handler exposes the example domain service over HTTP.
type Handler struct {
	service domain.Service
}

// New returns an HTTP handler for the example feature.
func New(service domain.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts the example routes on the provided echo group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/items", h.Create)
	g.GET("/items", h.List)
	g.GET("/items/:id", h.Get)
}

// Create handles POST /items.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Create(c *echo.Context) error {
	var req dto.CreateItemRequest
	if err := c.Bind(&req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	item, err := h.service.Create(c.Request().Context(), req.Name)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusCreated, mapItemToResponse(item))
}

// Get handles GET /items/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Get(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, body := sharederrors.HTTPError(domain.ErrInvalidID)
		return c.JSON(status, body)
	}

	item, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, mapItemToResponse(item))
}

// List handles GET /items.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) List(c *echo.Context) error {
	var req dto.ListItemsRequest
	if err := c.Bind(&req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}
	if err := c.Validate(req); err != nil {
		status, body := sharederrors.HTTPError(sharederrors.ErrInvalidInput)
		return c.JSON(status, body)
	}

	items, err := h.service.List(c.Request().Context(), req.Limit, req.Offset)
	if err != nil {
		status, body := sharederrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, mapItemsToResponse(items))
}

func mapItemToResponse(item *domain.Item) dto.ItemResponse {
	if item == nil {
		return dto.ItemResponse{}
	}
	return dto.ItemResponse{
		ID:        item.ID.String(),
		Name:      item.Name,
		CreatedAt: item.CreatedAt.Format(timeFormat),
		UpdatedAt: item.UpdatedAt.Format(timeFormat),
	}
}

func mapItemsToResponse(items []domain.Item) dto.ListItemsResponse {
	resp := dto.ListItemsResponse{Items: make([]dto.ItemResponse, len(items))}
	for i, item := range items {
		resp.Items[i] = mapItemToResponse(&item)
	}
	return resp
}
