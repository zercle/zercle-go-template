package handler

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/payment"
	"github.com/zercle/zercle-go-template/domain/payment/repository"
	"github.com/zercle/zercle-go-template/domain/payment/request"
	"github.com/zercle/zercle-go-template/domain/payment/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type paymentHandler struct {
	usecase payment.Usecase
	log     *logger.Logger
}

// NewPaymentHandler creates a new payment HTTP handler
func NewPaymentHandler(usecase payment.Usecase, log *logger.Logger) payment.Handler {
	return &paymentHandler{
		usecase: usecase,
		log:     log,
	}
}

// CreatePayment handles payment creation
// @Summary      Create a new payment
// @Description  Create a new payment for a booking
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body request.CreatePayment true "Payment details"
// @Success      201  {object}  map[string]interface{} "Payment created"
// @Failure      400  {object} map[string]interface{} "Validation error"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /payments [post]
func (h *paymentHandler) CreatePayment(c echo.Context) error {
	var req request.CreatePayment
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.CreatePayment(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to create payment", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrDuplicateTransactionID) {
			return response.BadRequest(c, "Payment with this transaction ID already exists", nil)
		}
		if errors.Is(err, usecase.ErrBookingNotFound) {
			return response.NotFound(c, "Booking not found")
		}
		return response.InternalError(c, "Failed to create payment")
	}

	return response.Created(c, result)
}

// GetPayment handles get payment by ID
// @Summary      Get payment by ID
// @Description  Get a single payment by its ID
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Payment ID"
// @Success      200  {object}  map[string]interface{} "Payment retrieved"
// @Failure      400  {object} map[string]interface{} "Invalid payment ID"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "Payment not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /payments/{id} [get]
func (h *paymentHandler) GetPayment(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID", nil)
	}

	result, err := h.usecase.GetPayment(c.Request().Context(), id)
	if err != nil {
		h.log.Error("Failed to get payment", "error", err, "request_id", middleware.GetRequestID(c), "payment_id", id)
		if errors.Is(err, usecase.ErrPaymentNotFound) || errors.Is(err, repository.ErrPaymentNotFound) {
			return response.NotFound(c, "Payment not found")
		}
		return response.InternalError(c, "Failed to get payment")
	}

	return response.OK(c, result)
}

// GetPaymentByBooking handles get payment by booking ID
// @Summary      Get payment by booking ID
// @Description  Get the payment associated with a specific booking
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        booking_id  path      string  true  "Booking ID"
// @Success      200         {object}  map[string]interface{}  "Payment retrieved"
// @Failure      400         {object} map[string]interface{} "Invalid booking ID"
// @Failure      401         {object} map[string]interface{} "Unauthorized"
// @Failure      404         {object} map[string]interface{} "Payment not found"
// @Failure      500         {object} map[string]interface{} "Internal server error"
// @Router       /bookings/{booking_id}/payments [get]
func (h *paymentHandler) GetPaymentByBooking(c echo.Context) error {
	idStr := c.Param("booking_id")
	bookingID, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid booking ID", nil)
	}

	result, err := h.usecase.GetPaymentByBooking(c.Request().Context(), bookingID)
	if err != nil {
		h.log.Error("Failed to get payment by booking", "error", err, "request_id", middleware.GetRequestID(c), "booking_id", bookingID)
		if errors.Is(err, usecase.ErrPaymentNotFound) || errors.Is(err, repository.ErrPaymentNotFound) {
			return response.NotFound(c, "Payment not found")
		}
		return response.InternalError(c, "Failed to get payment")
	}

	return response.OK(c, result)
}

// ConfirmPayment handles payment confirmation
// @Summary      Confirm a payment
// @Description  Confirm a pending payment (marks as completed)
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Payment ID"
// @Success      200  {object}  map[string]interface{} "Payment confirmed"
// @Failure      400  {object} map[string]interface{} "Cannot confirm payment with current status"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "Payment not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /payments/{id}/confirm [put]
func (h *paymentHandler) ConfirmPayment(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID", nil)
	}

	result, err := h.usecase.ConfirmPayment(c.Request().Context(), id)
	if err != nil {
		h.log.Error("Failed to confirm payment", "error", err, "request_id", middleware.GetRequestID(c), "payment_id", id)
		if errors.Is(err, usecase.ErrPaymentNotFound) || errors.Is(err, repository.ErrPaymentNotFound) {
			return response.NotFound(c, "Payment not found")
		}
		if errors.Is(err, usecase.ErrCannotConfirmPending) {
			return response.BadRequest(c, "Cannot confirm payment with current status", nil)
		}
		return response.InternalError(c, "Failed to confirm payment")
	}

	return response.OK(c, result)
}

// RefundPayment handles payment refund
// @Summary      Refund a payment
// @Description  Refund a completed payment
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path     string                true  "Payment ID"
// @Param        request  body     request.RefundPayment  true  "Refund reason"
// @Success      200      {object}  map[string]interface{}  "Payment refunded"
// @Failure      400      {object} map[string]interface{} "Cannot refund pending/refunded payment"
// @Failure      401      {object} map[string]interface{} "Unauthorized"
// @Failure      404      {object} map[string]interface{} "Payment not found"
// @Failure      500      {object} map[string]interface{} "Internal server error"
// @Router       /payments/{id}/refund [put]
func (h *paymentHandler) RefundPayment(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID", nil)
	}

	var req request.RefundPayment
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.RefundPayment(c.Request().Context(), id, req)
	if err != nil {
		h.log.Error("Failed to refund payment", "error", err, "request_id", middleware.GetRequestID(c), "payment_id", id)
		if errors.Is(err, usecase.ErrPaymentNotFound) || errors.Is(err, repository.ErrPaymentNotFound) {
			return response.NotFound(c, "Payment not found")
		}
		if errors.Is(err, usecase.ErrCannotRefundPending) {
			return response.BadRequest(c, "Cannot refund a pending payment", nil)
		}
		if errors.Is(err, usecase.ErrCannotRefundRefunded) {
			return response.BadRequest(c, "Payment already refunded", nil)
		}
		return response.InternalError(c, "Failed to refund payment")
	}

	return response.OK(c, result)
}

// ListPayments handles list payments
// @Summary      List payments
// @Description  Get a paginated list of payments with optional status filter
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        status  query  string  false  "Filter by payment status"  Enums(pending, completed, failed, refunded)
// @Param        limit   query  int     false  "Number of items (max 100)"  default(20)
// @Param        offset  query  int     false  "Number of items to skip"   default(0)
// @Success      200     {object}  map[string]interface{}  "Payments retrieved"
// @Failure      401     {object} map[string]interface{} "Unauthorized"
// @Failure      500     {object} map[string]interface{} "Internal server error"
// @Router       /payments [get]
func (h *paymentHandler) ListPayments(c echo.Context) error {
	status := c.QueryParam("status")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	req := request.ListPayments{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	result, err := h.usecase.ListPayments(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to list payments", "error", err, "request_id", middleware.GetRequestID(c))
		return response.InternalError(c, "Failed to list payments")
	}

	return response.OK(c, result)
}
