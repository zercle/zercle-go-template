package handler

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/domain/service/repository"
	"github.com/zercle/zercle-go-template/domain/service/request"
	"github.com/zercle/zercle-go-template/domain/service/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type serviceHandler struct {
	usecase service.Usecase
	log     *logger.Logger
}

// NewServiceHandler creates a new service HTTP handler
func NewServiceHandler(usecase service.Usecase, log *logger.Logger) service.Handler {
	return &serviceHandler{
		usecase: usecase,
		log:     log,
	}
}

// CreateService handles service creation (protected)
// @Summary      Create a new service
// @Description  Create a new service with name, description, duration, price, and capacity
// @Tags         services
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body request.CreateService true "Service details"
// @Success      201  {object}  map[string]interface{} "Service created"
// @Failure      400  {object} map[string]interface{} "Validation error"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /services [post]
func (h *serviceHandler) CreateService(c echo.Context) error {
	var req request.CreateService
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.CreateService(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to create service", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrInvalidServicePrice) {
			return response.BadRequest(c, "Price must be greater than 0", nil)
		}
		if errors.Is(err, usecase.ErrInvalidDuration) {
			return response.BadRequest(c, "Duration must be between 1 and 480 minutes", nil)
		}
		if errors.Is(err, usecase.ErrInvalidCapacity) {
			return response.BadRequest(c, "Capacity must be between 1 and 50", nil)
		}
		return response.InternalError(c, "Failed to create service")
	}

	return response.Created(c, result)
}

// GetService handles get service by ID (public)
// @Summary      Get service by ID
// @Description  Get a single service by its ID
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Service ID"
// @Success      200  {object}  map[string]interface{} "Service retrieved"
// @Failure      400  {object} map[string]interface{} "Invalid service ID"
// @Failure      404  {object} map[string]interface{} "Service not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /services/{id} [get]
func (h *serviceHandler) GetService(c echo.Context) error {
	if h.usecase == nil {
		return response.InternalError(c, "Service usecase not initialized")
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid service ID", nil)
	}

	result, err := h.usecase.GetService(c.Request().Context(), id)
	if err != nil {
		h.log.Error("Failed to get service", "error", err, "request_id", middleware.GetRequestID(c), "service_id", id)
		if errors.Is(err, usecase.ErrServiceNotFound) || errors.Is(err, repository.ErrServiceNotFound) {
			return response.NotFound(c, "Service not found")
		}
		return response.InternalError(c, "Failed to get service")
	}

	return response.OK(c, result)
}

// UpdateService handles service update (protected)
// @Summary      Update a service
// @Description  Update an existing service by ID
// @Tags         services
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path     string                true  "Service ID"
// @Param        request  body     request.UpdateService  true  "Service update details"
// @Success      200      {object}  map[string]interface{}  "Service updated"
// @Failure      400      {object} map[string]interface{} "Validation error"
// @Failure      401      {object} map[string]interface{} "Unauthorized"
// @Failure      404      {object} map[string]interface{} "Service not found"
// @Failure      500      {object} map[string]interface{} "Internal server error"
// @Router       /services/{id} [put]
func (h *serviceHandler) UpdateService(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid service ID", nil)
	}

	var req request.UpdateService
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.UpdateService(c.Request().Context(), id, req)
	if err != nil {
		h.log.Error("Failed to update service", "error", err, "request_id", middleware.GetRequestID(c), "service_id", id)
		if errors.Is(err, usecase.ErrServiceNotFound) || errors.Is(err, repository.ErrServiceNotFound) {
			return response.NotFound(c, "Service not found")
		}
		if errors.Is(err, usecase.ErrInvalidDuration) {
			return response.BadRequest(c, "Duration must be between 1 and 480 minutes", nil)
		}
		if errors.Is(err, usecase.ErrInvalidServicePrice) {
			return response.BadRequest(c, "Price must be greater than 0", nil)
		}
		if errors.Is(err, usecase.ErrInvalidCapacity) {
			return response.BadRequest(c, "Capacity must be between 1 and 50", nil)
		}
		return response.InternalError(c, "Failed to update service")
	}

	return response.OK(c, result)
}

// DeleteService handles service deletion (protected)
// @Summary      Delete a service
// @Description  Delete an existing service by ID
// @Tags         services
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Service ID"
// @Success      204  "Service deleted"
// @Failure      400  {object} map[string]interface{} "Invalid service ID"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "Service not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /services/{id} [delete]
func (h *serviceHandler) DeleteService(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid service ID", nil)
	}

	if err := h.usecase.DeleteService(c.Request().Context(), id); err != nil {
		h.log.Error("Failed to delete service", "error", err, "request_id", middleware.GetRequestID(c), "service_id", id)
		if errors.Is(err, usecase.ErrServiceNotFound) || errors.Is(err, repository.ErrServiceNotFound) {
			return response.NotFound(c, "Service not found")
		}
		return response.InternalError(c, "Failed to delete service")
	}

	return response.NoContent(c)
}

// ListServices handles list services with pagination (public)
// @Summary      List services
// @Description  Get a paginated list of services
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        is_active  query     string  false  "Filter by active status"  Enums(true, false)
// @Param        limit      query     int     false  "Number of items (max 100)"  default(20)
// @Param        offset     query     int     false  "Number of items to skip"   default(0)
// @Success      200        {object}  map[string]interface{} "Services retrieved"
// @Failure      500        {object} map[string]interface{} "Internal server error"
// @Router       /services [get]
func (h *serviceHandler) ListServices(c echo.Context) error {
	// Parse query parameters
	isActive := c.QueryParam("is_active") == "true"

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	req := request.ListServices{
		IsActive: isActive,
		Limit:    limit,
		Offset:   offset,
	}

	if h.usecase == nil {
		return response.InternalError(c, "Service usecase not initialized")
	}

	result, err := h.usecase.ListServices(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to list services", "error", err, "request_id", middleware.GetRequestID(c))
		return response.InternalError(c, "Failed to list services")
	}

	return response.OK(c, result)
}

// SearchServices handles search services by name (public)
// @Summary      Search services
// @Description  Search for services by name
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        name      query     string  true   "Search term for service name"
// @Param        is_active  query     string  false  "Filter by active status"  Enums(true, false)
// @Param        limit      query     int     false  "Number of items (max 100)"  default(20)
// @Success      200        {object} map[string]interface{} "Search results"
// @Failure      400        {object} map[string]interface{} "Missing name parameter"
// @Failure      500        {object} map[string]interface{} "Internal server error"
// @Router       /services/search [get]
func (h *serviceHandler) SearchServices(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return response.BadRequest(c, "Name parameter is required", nil)
	}

	isActive := c.QueryParam("is_active") != "false" // default true

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	result, err := h.usecase.SearchServices(c.Request().Context(), name, isActive, limit)
	if err != nil {
		h.log.Error("Failed to search services", "error", err, "request_id", middleware.GetRequestID(c), "name", name)
		return response.InternalError(c, "Failed to search services")
	}

	return response.OK(c, map[string]interface{}{
		"services": result,
		"count":    len(result),
	})
}
