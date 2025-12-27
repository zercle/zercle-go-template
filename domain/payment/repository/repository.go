package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zercle/zercle-go-template/domain/payment"
	"github.com/zercle/zercle-go-template/domain/payment/model"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// ErrPaymentNotFound is returned when a payment cannot be found
var ErrPaymentNotFound = errors.New("payment not found")

type paymentRepository struct {
	sqlc *db.Queries
	log  *logger.Logger
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(sqlc *db.Queries, log *logger.Logger) payment.Repository {
	return &paymentRepository{
		sqlc: sqlc,
		log:  log,
	}
}

// Helper functions for pgtype conversions
func toUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func fromUUID(u pgtype.UUID) uuid.UUID {
	return u.Bytes
}

func toText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromTimestamptz(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func fromTimestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	result := t.Time
	return &result
}

// fromNumeric converts pgtype.Numeric to float64
func fromNumeric(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	var str string
	if n.Int != nil {
		str = n.Int.String()
		if n.Exp > 0 {
			if len(str) <= int(n.Exp) {
				for len(str) <= int(n.Exp) {
					str = "0" + str
				}
			}
			pos := len(str) - int(n.Exp)
			str = str[:pos] + "." + str[pos:]
		}
	}
	var f float64
	_, _ = fmt.Sscanf(str, "%f", &f)
	return f
}

// toNumeric converts float64 to pgtype.Numeric
func toNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}

// toInt32Safe safely converts int to int32 with overflow check.
// Panics if value is outside int32 range (should not happen with validated input).
func toInt32Safe(i int) int32 {
	if i < math.MinInt32 || i > math.MaxInt32 {
		panic(fmt.Sprintf("value %d overflows int32", i))
	}
	return int32(i)
}

func (r *paymentRepository) Create(ctx context.Context, paymentModel *model.Payment) (*model.Payment, error) {
	now := time.Now()
	id := uuid.New()

	params := db.CreatePaymentParams{
		ID:            toUUID(id),
		BookingID:     toUUID(paymentModel.BookingID),
		Amount:        toNumeric(paymentModel.Amount),
		Status:        string(paymentModel.Status),
		PaymentMethod: toText(string(paymentModel.PaymentMethod)),
		TransactionID: toText(paymentModel.TransactionID),
		CreatedAt:     toTimestamptz(now),
		UpdatedAt:     toTimestamptz(now),
	}

	row, err := r.sqlc.CreatePayment(ctx, params)
	if err != nil {
		r.log.Error("Failed to create payment", "error", err, "booking_id", paymentModel.BookingID)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	row, err := r.sqlc.GetPayment(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.log.Error("Failed to get payment by ID", "error", err, "payment_id", id)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error) {
	row, err := r.sqlc.GetPaymentByBooking(ctx, toUUID(bookingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.log.Error("Failed to get payment by booking ID", "error", err, "booking_id", bookingID)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*model.Payment, error) {
	row, err := r.sqlc.GetPaymentByTransactionId(ctx, toText(transactionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.log.Error("Failed to get payment by transaction ID", "error", err, "transaction_id", transactionID)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus, transactionID string) (*model.Payment, error) {
	params := db.UpdatePaymentStatusParams{
		ID:            toUUID(id),
		Status:        string(status),
		UpdatedAt:     toTimestamptz(time.Now()),
		TransactionID: toText(transactionID),
	}

	row, err := r.sqlc.UpdatePaymentStatus(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.log.Error("Failed to update payment status", "error", err, "payment_id", id)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) Confirm(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	now := time.Now()
	params := db.ConfirmPaymentParams{
		ID:        toUUID(id),
		PaidAt:    toTimestamptz(now),
		UpdatedAt: toTimestamptz(now),
	}

	row, err := r.sqlc.ConfirmPayment(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.log.Error("Failed to confirm payment", "error", err, "payment_id", id)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) Refund(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	now := time.Now()
	params := db.RefundPaymentParams{
		ID:         toUUID(id),
		RefundedAt: toTimestamptz(now),
		UpdatedAt:  toTimestamptz(now),
	}

	row, err := r.sqlc.RefundPayment(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.log.Error("Failed to refund payment", "error", err, "payment_id", id)
		return nil, err
	}

	return &model.Payment{
		ID:            fromUUID(row.ID),
		BookingID:     fromUUID(row.BookingID),
		Amount:        fromNumeric(row.Amount),
		Status:        model.PaymentStatus(row.Status),
		PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
		TransactionID: fromText(row.TransactionID),
		PaidAt:        fromTimestamptzPtr(row.PaidAt),
		RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
		CreatedAt:     fromTimestamptz(row.CreatedAt),
		UpdatedAt:     fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *paymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.sqlc.DeletePayment(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentNotFound
		}
		r.log.Error("Failed to delete payment", "error", err, "payment_id", id)
		return err
	}
	return nil
}

func (r *paymentRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*model.Payment, error) {
	params := db.ListPaymentsByStatusParams{
		Status: status,
		Limit:  toInt32Safe(limit),
		Offset: toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListPaymentsByStatus(ctx, params)
	if err != nil {
		r.log.Error("Failed to list payments by status", "error", err, "status", status)
		return nil, err
	}

	payments := make([]*model.Payment, len(rows))
	for i, row := range rows {
		payments[i] = &model.Payment{
			ID:            fromUUID(row.ID),
			BookingID:     fromUUID(row.BookingID),
			Amount:        fromNumeric(row.Amount),
			Status:        model.PaymentStatus(row.Status),
			PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
			TransactionID: fromText(row.TransactionID),
			PaidAt:        fromTimestamptzPtr(row.PaidAt),
			RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
			CreatedAt:     fromTimestamptz(row.CreatedAt),
			UpdatedAt:     fromTimestamptz(row.UpdatedAt),
		}
	}

	return payments, nil
}

func (r *paymentRepository) ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]*model.Payment, error) {
	rows, err := r.sqlc.ListPaymentsByBooking(ctx, toUUID(bookingID))
	if err != nil {
		r.log.Error("Failed to list payments by booking", "error", err, "booking_id", bookingID)
		return nil, err
	}

	payments := make([]*model.Payment, len(rows))
	for i, row := range rows {
		payments[i] = &model.Payment{
			ID:            fromUUID(row.ID),
			BookingID:     fromUUID(row.BookingID),
			Amount:        fromNumeric(row.Amount),
			Status:        model.PaymentStatus(row.Status),
			PaymentMethod: model.PaymentMethod(fromText(row.PaymentMethod)),
			TransactionID: fromText(row.TransactionID),
			PaidAt:        fromTimestamptzPtr(row.PaidAt),
			RefundedAt:    fromTimestamptzPtr(row.RefundedAt),
			CreatedAt:     fromTimestamptz(row.CreatedAt),
			UpdatedAt:     fromTimestamptz(row.UpdatedAt),
		}
	}

	return payments, nil
}
