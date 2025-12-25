package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/booking"
	bookingRepository "github.com/zercle/zercle-go-template/domain/booking/repository"
	"github.com/zercle/zercle-go-template/domain/booking/request"
	"github.com/zercle/zercle-go-template/domain/booking/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type bookingHandler struct {
	usecase booking.Usecase
	log     *logger.Logger
}

// NewBookingHandler creates a new booking HTTP handler
func NewBookingHandler(usecase booking.Usecase, log *logger.Logger) booking.Handler {
	return &bookingHandler{
		usecase: usecase,
		log:     log,
	}
}

// CreateBooking handles booking creation
// @Summary      Create a new booking
// @Description  Create a new booking for a service
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body request.CreateBooking true "Booking details"
// @Success      201  {object}  map[string]interface{} "Booking created"
// @Failure      400  {object}  map[string]interface{} "Validation error"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Service not found"
// @Failure      500  {object}  map[string]interface{} "Internal server error"
// @Router       /bookings [post]
func (h *bookingHandler) CreateBooking(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	var req request.CreateBooking
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.CreateBooking(c.Request().Context(), userUUID, req)
	if err != nil {
		h.log.Error("Failed to create booking", "error", err, "request_id", middleware.GetRequestID(c), "user_id", userID)
		if errors.Is(err, usecase.ErrServiceNotFound) {
			return response.NotFound(c, "Service not found")
		}
		if errors.Is(err, usecase.ErrBookingTimeInPast) {
			return response.BadRequest(c, "Booking time must be in the future", nil)
		}
		if errors.Is(err, usecase.ErrBookingConflict) {
			return response.BadRequest(c, "Booking time conflicts with existing booking", nil)
		}
		return response.InternalError(c, "Failed to create booking")
	}

	return response.Created(c, result)
}

// GetBooking handles get booking by ID
// @Summary      Get booking by ID
// @Description  Get a single booking by its ID
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Booking ID"
// @Success      200  {object}  map[string]interface{} "Booking retrieved"
// @Failure      400  {object} map[string]interface{} "Invalid booking ID"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "Booking not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /bookings/{id} [get]
func (h *bookingHandler) GetBooking(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid booking ID", nil)
	}

	result, err := h.usecase.GetBooking(c.Request().Context(), id)
	if err != nil {
		h.log.Error("Failed to get booking", "error", err, "request_id", middleware.GetRequestID(c), "booking_id", id)
		if errors.Is(err, usecase.ErrBookingNotFound) || errors.Is(err, bookingRepository.ErrBookingNotFound) {
			return response.NotFound(c, "Booking not found")
		}
		return response.InternalError(c, "Failed to get booking")
	}

	return response.OK(c, result)
}

// CancelBooking handles booking cancellation
// @Summary      Cancel a booking
// @Description  Cancel an existing booking (only by the booking owner)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Booking ID"
// @Success      200  {object}  map[string]interface{} "Booking canceled"
// @Failure      400  {object} map[string]interface{} "Cannot cancel completed booking"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      403  {object} map[string]interface{} "Forbidden - not your booking"
// @Failure      404  {object} map[string]interface{} "Booking not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /bookings/{id}/cancel [put]
func (h *bookingHandler) CancelBooking(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid booking ID", nil)
	}

	result, err := h.usecase.CancelBooking(c.Request().Context(), id, userUUID)
	if err != nil {
		h.log.Error("Failed to cancel booking", "error", err, "request_id", middleware.GetRequestID(c), "booking_id", id)
		if errors.Is(err, usecase.ErrBookingNotFound) || errors.Is(err, bookingRepository.ErrBookingNotFound) {
			return response.NotFound(c, "Booking not found")
		}
		if errors.Is(err, usecase.ErrUnauthorizedAccess) {
			return response.Forbidden(c, "You don't have permission to cancel this booking")
		}
		if errors.Is(err, usecase.ErrCannotCancelComplete) {
			return response.BadRequest(c, "Cannot cancel a completed booking", nil)
		}
		return response.InternalError(c, "Failed to cancel booking")
	}

	return response.OK(c, result)
}

// UpdateBookingStatus handles booking status update
// @Summary      Update booking status
// @Description  Update the status of a booking (admin/staff only)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path     string                        true  "Booking ID"
// @Param        request  body     request.UpdateBookingStatus  true  "Status update"
// @Success      200      {object}  map[string]interface{}  "Status updated"
// @Failure      400      {object} map[string]interface{} "Invalid status transition"
// @Failure      401      {object} map[string]interface{} "Unauthorized"
// @Failure      404      {object} map[string]interface{} "Booking not found"
// @Failure      500      {object} map[string]interface{} "Internal server error"
// @Router       /bookings/{id}/status [put]
func (h *bookingHandler) UpdateBookingStatus(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid booking ID", nil)
	}

	var req request.UpdateBookingStatus
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.UpdateBookingStatus(c.Request().Context(), id, req)
	if err != nil {
		h.log.Error("Failed to update booking status", "error", err, "request_id", middleware.GetRequestID(c), "booking_id", id)
		if errors.Is(err, usecase.ErrBookingNotFound) || errors.Is(err, bookingRepository.ErrBookingNotFound) {
			return response.NotFound(c, "Booking not found")
		}
		if errors.Is(err, usecase.ErrInvalidStatus) {
			return response.BadRequest(c, "Invalid status transition", nil)
		}
		return response.InternalError(c, "Failed to update booking status")
	}

	return response.OK(c, result)
}

// ListBookingsByUser handles list user's bookings
// @Summary      List user's bookings
// @Description  Get a paginated list of the authenticated user's bookings
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        limit   query  int  false  "Number of items (max 100)"  default(20)
// @Param        offset  query  int  false  "Number of items to skip"   default(0)
// @Success      200     {object}  map[string]interface{}  "Bookings retrieved"
// @Failure      401     {object} map[string]interface{} "Unauthorized"
// @Failure      500     {object} map[string]interface{} "Internal server error"
// @Router       /bookings [get]
func (h *bookingHandler) ListBookingsByUser(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	result, err := h.usecase.ListBookingsByUser(c.Request().Context(), userUUID, limit, offset)
	if err != nil {
		h.log.Error("Failed to list bookings by user", "error", err, "request_id", middleware.GetRequestID(c), "user_id", userID)
		return response.InternalError(c, "Failed to list bookings")
	}

	return response.OK(c, result)
}

// ListBookingsByService handles list bookings by service
// @Summary      List bookings by service
// @Description  Get a paginated list of bookings for a specific service (public)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id      path  string  true  "Service ID"
// @Param        limit   query  int     false  "Number of items (max 100)"  default(20)
// @Param        offset  query  int     false  "Number of items to skip"   default(0)
// @Success      200     {object} map[string]interface{} "Bookings retrieved"
// @Failure      400     {object} map[string]interface{} "Invalid service ID"
// @Failure      500     {object} map[string]interface{} "Internal server error"
// @Router       /bookings/services/{id} [get]
func (h *bookingHandler) ListBookingsByService(c echo.Context) error {
	idStr := c.Param("id")
	serviceID, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid service ID", nil)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	result, err := h.usecase.ListBookingsByService(c.Request().Context(), serviceID, limit, offset)
	if err != nil {
		h.log.Error("Failed to list bookings by service", "error", err, "request_id", middleware.GetRequestID(c), "service_id", serviceID)
		return response.InternalError(c, "Failed to list bookings")
	}

	return response.OK(c, map[string]interface{}{
		"bookings": result,
		"count":    len(result),
	})
}

// ListBookingsByDateRange handles list bookings by date range
// @Summary      List bookings by date range
// @Description  Get a list of bookings within a specific date range (public)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        start_date  query  string  true   "Start date (RFC3339 format)"
// @Param        end_date    query  string  true   "End date (RFC3339 format)"
// @Param        limit       query  int     false  "Number of items (max 100)"  default(20)
// @Param        offset      query  int     false  "Number of items to skip"   default(0)
// @Success      200         {object} map[string]interface{} "Bookings retrieved"
// @Failure      400         {object} map[string]interface{} "Invalid date range or format"
// @Failure      500         {object} map[string]interface{} "Internal server error"
// @Router       /bookings/dates [get]
func (h *bookingHandler) ListBookingsByDateRange(c echo.Context) error {
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	if startDateStr == "" || endDateStr == "" {
		return response.BadRequest(c, "start_date and end_date are required", nil)
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return response.BadRequest(c, "Invalid start_date format. Use RFC3339 format", nil)
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return response.BadRequest(c, "Invalid end_date format. Use RFC3339 format", nil)
	}

	if startDate.After(endDate) {
		return response.BadRequest(c, "start_date must be before or equal to end_date", nil)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	req := request.ListBookingsByDateRange{
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.usecase.ListBookingsByDateRange(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to list bookings by date range", "error", err, "request_id", middleware.GetRequestID(c))
		return response.InternalError(c, "Failed to list bookings")
	}

	return response.OK(c, map[string]interface{}{
		"bookings": result,
		"count":    len(result),
	})
}
