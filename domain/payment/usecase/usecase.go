package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/payment"
	"github.com/zercle/zercle-go-template/domain/payment/model"
	"github.com/zercle/zercle-go-template/domain/payment/repository"
	"github.com/zercle/zercle-go-template/domain/payment/request"
	paymentResponse "github.com/zercle/zercle-go-template/domain/payment/response"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

var (
	// ErrPaymentNotFound is returned when a payment cannot be found
	ErrPaymentNotFound = errors.New("payment not found")
	// ErrBookingNotFound is returned when a booking cannot be found
	ErrBookingNotFound = errors.New("booking not found")
	// ErrInvalidPaymentStatus is returned when payment status is invalid
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	// ErrCannotConfirmPending is returned when trying to confirm a pending payment
	ErrCannotConfirmPending = errors.New("cannot confirm payment with pending status")
	// ErrCannotRefundPending is returned when trying to refund a pending payment
	ErrCannotRefundPending = errors.New("cannot refund a pending payment")
	// ErrCannotRefundRefunded is returned when trying to refund an already refunded payment
	ErrCannotRefundRefunded = errors.New("payment already refunded")
	// ErrDuplicateTransactionID is returned when transaction ID already exists
	ErrDuplicateTransactionID = errors.New("payment with this transaction ID already exists")
)

type paymentUseCase struct {
	repo payment.Repository
	log  *logger.Logger
}

// NewPaymentUseCase creates a new payment use case with dependencies
func NewPaymentUseCase(repo payment.Repository, log *logger.Logger) payment.Usecase {
	return &paymentUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *paymentUseCase) CreatePayment(ctx context.Context, req request.CreatePayment) (*paymentResponse.PaymentResponse, error) {
	// Check if transaction ID already exists (if provided)
	if req.TransactionID != "" {
		existing, err := uc.repo.GetByTransactionID(ctx, req.TransactionID)
		if err == nil && existing != nil {
			return nil, ErrDuplicateTransactionID
		}
		if !errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, err
		}
	}

	// Create payment model
	paymentModel := &model.Payment{
		BookingID:     req.BookingID,
		Status:        model.PaymentStatusPending,
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
	}

	created, err := uc.repo.Create(ctx, paymentModel)
	if err != nil {
		uc.log.Error("Failed to create payment", "error", err, "booking_id", req.BookingID)
		// Check if it's a foreign key violation (booking not found)
		if strings.Contains(err.Error(), "violates foreign key constraint") ||
			strings.Contains(err.Error(), "SQLSTATE 23503") {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	return toPaymentResponse(created), nil
}

func (uc *paymentUseCase) GetPayment(ctx context.Context, id uuid.UUID) (*paymentResponse.PaymentResponse, error) {
	paymentModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return toPaymentResponse(paymentModel), nil
}

func (uc *paymentUseCase) GetPaymentByBooking(ctx context.Context, bookingID uuid.UUID) (*paymentResponse.PaymentResponse, error) {
	paymentModel, err := uc.repo.GetByBookingID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return toPaymentResponse(paymentModel), nil
}

func (uc *paymentUseCase) ConfirmPayment(ctx context.Context, id uuid.UUID) (*paymentResponse.PaymentResponse, error) {
	// Get payment first to check current status
	paymentModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// Check if can be confirmed
	if paymentModel.Status == model.PaymentStatusCompleted {
		return toPaymentResponse(paymentModel), nil
	}

	if paymentModel.Status != model.PaymentStatusPending {
		return nil, ErrCannotConfirmPending
	}

	// Confirm payment
	confirmed, err := uc.repo.Confirm(ctx, id)
	if err != nil {
		uc.log.Error("Failed to confirm payment", "error", err, "payment_id", id)
		return nil, err
	}

	return toPaymentResponse(confirmed), nil
}

func (uc *paymentUseCase) RefundPayment(ctx context.Context, id uuid.UUID, req request.RefundPayment) (*paymentResponse.PaymentResponse, error) {
	// Get payment first to check current status
	paymentModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// Check if can be refunded
	if paymentModel.Status == model.PaymentStatusRefunded {
		return nil, ErrCannotRefundRefunded
	}

	if paymentModel.Status != model.PaymentStatusCompleted {
		return nil, ErrCannotRefundPending
	}

	// Refund payment
	refunded, err := uc.repo.Refund(ctx, id)
	if err != nil {
		uc.log.Error("Failed to refund payment", "error", err, "payment_id", id)
		return nil, err
	}

	return toPaymentResponse(refunded), nil
}

func (uc *paymentUseCase) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, req request.UpdatePaymentStatus) (*paymentResponse.PaymentResponse, error) {
	// Get payment first to check current status
	paymentModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// Validate status transition
	if !isValidPaymentStatusTransition(paymentModel.Status, req.Status) {
		return nil, ErrInvalidPaymentStatus
	}

	// Update status
	updated, err := uc.repo.UpdateStatus(ctx, id, req.Status, req.TransactionID)
	if err != nil {
		uc.log.Error("Failed to update payment status", "error", err, "payment_id", id)
		return nil, err
	}

	return toPaymentResponse(updated), nil
}

func (uc *paymentUseCase) ListPayments(ctx context.Context, req request.ListPayments) (*paymentResponse.ListPaymentsResponse, error) {
	// Apply defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	payments, err := uc.repo.ListByStatus(ctx, req.Status, int(req.Limit), int(req.Offset))
	if err != nil {
		uc.log.Error("Failed to list payments", "error", err, "status", req.Status)
		return nil, err
	}

	paymentResponses := make([]paymentResponse.PaymentResponse, len(payments))
	for i, p := range payments {
		paymentResponses[i] = *toPaymentResponse(p)
	}

	return &paymentResponse.ListPaymentsResponse{
		Payments: paymentResponses,
		Total:    len(paymentResponses),
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// toPaymentResponse converts a payment model to response DTO
func toPaymentResponse(p *model.Payment) *paymentResponse.PaymentResponse {
	return &paymentResponse.PaymentResponse{
		ID:            p.ID,
		BookingID:     p.BookingID,
		Amount:        p.Amount,
		Status:        p.Status,
		PaymentMethod: p.PaymentMethod,
		TransactionID: p.TransactionID,
		PaidAt:        p.PaidAt,
		RefundedAt:    p.RefundedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// isValidPaymentStatusTransition validates if a status transition is allowed
func isValidPaymentStatusTransition(current, newStatus model.PaymentStatus) bool {
	// Define valid transitions
	validTransitions := map[model.PaymentStatus][]model.PaymentStatus{
		model.PaymentStatusPending:   {model.PaymentStatusCompleted, model.PaymentStatusFailed},
		model.PaymentStatusCompleted: {model.PaymentStatusRefunded},
		model.PaymentStatusRefunded:  {},                           // No transitions from refunded
		model.PaymentStatusFailed:    {model.PaymentStatusPending}, // Can retry failed payments
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}
