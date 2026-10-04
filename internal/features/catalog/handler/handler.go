package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	"github.com/zercle/zercle-go-template/internal/features/catalog/usecase"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Handler exposes the catalog usecase service over HTTP.
type Handler struct {
	service usecase.Service
}

// New returns an HTTP handler for the catalog feature.
func New(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts the catalog routes on the provided echo group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/products", h.Create)
	g.GET("/products", h.List)
	g.GET("/products/:id", h.Get)
}

// NOTE: Echo v5 changed echo.Context from an interface (v4) to a struct, and
// echo.HandlerFunc is now `func(c *Context) error`. Handlers therefore take
// *echo.Context — this is correct for v5, not a mistake.

// Create handles POST /products.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Create(c *echo.Context) error {
	var req contract.CreateProductRequest
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

// Get handles GET /products/:id.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) Get(c *echo.Context) error {
	resp, err := h.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		status, body := apperrors.HTTPError(err)
		return c.JSON(status, body)
	}

	return c.JSON(http.StatusOK, resp)
}

// List handles GET /products.
// nolint:wrapcheck // echo handlers return the JSON write error directly.
func (h *Handler) List(c *echo.Context) error {
	var req contract.ListProductsRequest
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
