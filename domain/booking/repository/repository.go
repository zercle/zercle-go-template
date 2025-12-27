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
	"github.com/zercle/zercle-go-template/domain/booking"
	"github.com/zercle/zercle-go-template/domain/booking/model"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// ErrBookingNotFound is returned when a booking cannot be found
var ErrBookingNotFound = errors.New("booking not found")

type bookingRepository struct {
	sqlc *db.Queries
	log  *logger.Logger
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(sqlc *db.Queries, log *logger.Logger) booking.Repository {
	return &bookingRepository{
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

func (r *bookingRepository) Create(ctx context.Context, booking *model.Booking) (*model.Booking, error) {
	now := time.Now()
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	params := db.CreateBookingParams{
		ID:         toUUID(id),
		UserID:     toUUID(booking.UserID),
		ServiceID:  toUUID(booking.ServiceID),
		StartTime:  toTimestamptz(booking.StartTime),
		EndTime:    toTimestamptz(booking.EndTime),
		Status:     string(booking.Status),
		TotalPrice: toNumeric(booking.TotalPrice),
		Notes:      toText(booking.Notes),
		CreatedAt:  toTimestamptz(now),
		UpdatedAt:  toTimestamptz(now),
	}

	row, err := r.sqlc.CreateBooking(ctx, params)
	if err != nil {
		r.log.Error("Failed to create booking", "error", err, "user_id", booking.UserID, "service_id", booking.ServiceID)
		return nil, err
	}

	return &model.Booking{
		ID:          fromUUID(row.ID),
		UserID:      fromUUID(row.UserID),
		ServiceID:   fromUUID(row.ServiceID),
		StartTime:   fromTimestamptz(row.StartTime),
		EndTime:     fromTimestamptz(row.EndTime),
		Status:      model.BookingStatus(row.Status),
		TotalPrice:  fromNumeric(row.TotalPrice),
		Notes:       fromText(row.Notes),
		CancelledAt: fromTimestamptzPtr(row.CancelledAt),
		CreatedAt:   fromTimestamptz(row.CreatedAt),
		UpdatedAt:   fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *bookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	row, err := r.sqlc.GetBooking(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		r.log.Error("Failed to get booking by ID", "error", err, "booking_id", id)
		return nil, err
	}

	return &model.Booking{
		ID:          fromUUID(row.ID),
		UserID:      fromUUID(row.UserID),
		ServiceID:   fromUUID(row.ServiceID),
		StartTime:   fromTimestamptz(row.StartTime),
		EndTime:     fromTimestamptz(row.EndTime),
		Status:      model.BookingStatus(row.Status),
		TotalPrice:  fromNumeric(row.TotalPrice),
		Notes:       fromText(row.Notes),
		CancelledAt: fromTimestamptzPtr(row.CancelledAt),
		CreatedAt:   fromTimestamptz(row.CreatedAt),
		UpdatedAt:   fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) (*model.Booking, error) {
	params := db.UpdateBookingStatusParams{
		ID:        toUUID(id),
		Status:    string(status),
		UpdatedAt: toTimestamptz(time.Now()),
	}

	row, err := r.sqlc.UpdateBookingStatus(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		r.log.Error("Failed to update booking status", "error", err, "booking_id", id)
		return nil, err
	}

	return &model.Booking{
		ID:          fromUUID(row.ID),
		UserID:      fromUUID(row.UserID),
		ServiceID:   fromUUID(row.ServiceID),
		StartTime:   fromTimestamptz(row.StartTime),
		EndTime:     fromTimestamptz(row.EndTime),
		Status:      model.BookingStatus(row.Status),
		TotalPrice:  fromNumeric(row.TotalPrice),
		Notes:       fromText(row.Notes),
		CancelledAt: fromTimestamptzPtr(row.CancelledAt),
		CreatedAt:   fromTimestamptz(row.CreatedAt),
		UpdatedAt:   fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *bookingRepository) Cancel(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	now := time.Now()
	params := db.CancelBookingParams{
		ID:          toUUID(id),
		CancelledAt: toTimestamptz(now),
		UpdatedAt:   toTimestamptz(now),
	}

	row, err := r.sqlc.CancelBooking(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		r.log.Error("Failed to cancel booking", "error", err, "booking_id", id)
		return nil, err
	}

	return &model.Booking{
		ID:          fromUUID(row.ID),
		UserID:      fromUUID(row.UserID),
		ServiceID:   fromUUID(row.ServiceID),
		StartTime:   fromTimestamptz(row.StartTime),
		EndTime:     fromTimestamptz(row.EndTime),
		Status:      model.BookingStatus(row.Status),
		TotalPrice:  fromNumeric(row.TotalPrice),
		Notes:       fromText(row.Notes),
		CancelledAt: fromTimestamptzPtr(row.CancelledAt),
		CreatedAt:   fromTimestamptz(row.CreatedAt),
		UpdatedAt:   fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *bookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.sqlc.DeleteBooking(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBookingNotFound
		}
		r.log.Error("Failed to delete booking", "error", err, "booking_id", id)
		return err
	}
	return nil
}

func (r *bookingRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	params := db.ListBookingsByUserParams{
		UserID: toUUID(userID),
		Limit:  toInt32Safe(limit),
		Offset: toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListBookingsByUser(ctx, params)
	if err != nil {
		r.log.Error("Failed to list bookings by user", "error", err, "user_id", userID)
		return nil, err
	}

	bookings := make([]*model.Booking, len(rows))
	for i, row := range rows {
		bookings[i] = &model.Booking{
			ID:          fromUUID(row.ID),
			UserID:      fromUUID(row.UserID),
			ServiceID:   fromUUID(row.ServiceID),
			StartTime:   fromTimestamptz(row.StartTime),
			EndTime:     fromTimestamptz(row.EndTime),
			Status:      model.BookingStatus(row.Status),
			TotalPrice:  fromNumeric(row.TotalPrice),
			Notes:       fromText(row.Notes),
			CancelledAt: fromTimestamptzPtr(row.CancelledAt),
			CreatedAt:   fromTimestamptz(row.CreatedAt),
			UpdatedAt:   fromTimestamptz(row.UpdatedAt),
		}
	}

	return bookings, nil
}

func (r *bookingRepository) ListByService(ctx context.Context, serviceID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	params := db.ListBookingsByServiceParams{
		ServiceID: toUUID(serviceID),
		Limit:     toInt32Safe(limit),
		Offset:    toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListBookingsByService(ctx, params)
	if err != nil {
		r.log.Error("Failed to list bookings by service", "error", err, "service_id", serviceID)
		return nil, err
	}

	bookings := make([]*model.Booking, len(rows))
	for i, row := range rows {
		bookings[i] = &model.Booking{
			ID:          fromUUID(row.ID),
			UserID:      fromUUID(row.UserID),
			ServiceID:   fromUUID(row.ServiceID),
			StartTime:   fromTimestamptz(row.StartTime),
			EndTime:     fromTimestamptz(row.EndTime),
			Status:      model.BookingStatus(row.Status),
			TotalPrice:  fromNumeric(row.TotalPrice),
			Notes:       fromText(row.Notes),
			CancelledAt: fromTimestamptzPtr(row.CancelledAt),
			CreatedAt:   fromTimestamptz(row.CreatedAt),
			UpdatedAt:   fromTimestamptz(row.UpdatedAt),
		}
	}

	return bookings, nil
}

func (r *bookingRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*model.Booking, error) {
	params := db.ListBookingsByStatusParams{
		Status: status,
		Limit:  toInt32Safe(limit),
		Offset: toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListBookingsByStatus(ctx, params)
	if err != nil {
		r.log.Error("Failed to list bookings by status", "error", err, "status", status)
		return nil, err
	}

	bookings := make([]*model.Booking, len(rows))
	for i, row := range rows {
		bookings[i] = &model.Booking{
			ID:          fromUUID(row.ID),
			UserID:      fromUUID(row.UserID),
			ServiceID:   fromUUID(row.ServiceID),
			StartTime:   fromTimestamptz(row.StartTime),
			EndTime:     fromTimestamptz(row.EndTime),
			Status:      model.BookingStatus(row.Status),
			TotalPrice:  fromNumeric(row.TotalPrice),
			Notes:       fromText(row.Notes),
			CancelledAt: fromTimestamptzPtr(row.CancelledAt),
			CreatedAt:   fromTimestamptz(row.CreatedAt),
			UpdatedAt:   fromTimestamptz(row.UpdatedAt),
		}
	}

	return bookings, nil
}

func (r *bookingRepository) ListByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*model.Booking, error) {
	params := db.ListBookingsByDateRangeParams{
		StartTime:   toTimestamptz(startDate),
		StartTime_2: toTimestamptz(endDate),
		Limit:       toInt32Safe(limit),
		Offset:      toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListBookingsByDateRange(ctx, params)
	if err != nil {
		r.log.Error("Failed to list bookings by date range", "error", err)
		return nil, err
	}

	bookings := make([]*model.Booking, len(rows))
	for i, row := range rows {
		bookings[i] = &model.Booking{
			ID:          fromUUID(row.ID),
			UserID:      fromUUID(row.UserID),
			ServiceID:   fromUUID(row.ServiceID),
			StartTime:   fromTimestamptz(row.StartTime),
			EndTime:     fromTimestamptz(row.EndTime),
			Status:      model.BookingStatus(row.Status),
			TotalPrice:  fromNumeric(row.TotalPrice),
			Notes:       fromText(row.Notes),
			CancelledAt: fromTimestamptzPtr(row.CancelledAt),
			CreatedAt:   fromTimestamptz(row.CreatedAt),
			UpdatedAt:   fromTimestamptz(row.UpdatedAt),
		}
	}

	return bookings, nil
}

func (r *bookingRepository) CheckConflict(ctx context.Context, serviceID uuid.UUID, startTime, endTime time.Time) ([]*model.Booking, error) {
	params := db.CheckBookingConflictParams{
		ServiceID:   toUUID(serviceID),
		StartTime:   toTimestamptz(startTime),
		StartTime_2: toTimestamptz(endTime),
	}

	rows, err := r.sqlc.CheckBookingConflict(ctx, params)
	if err != nil {
		r.log.Error("Failed to check booking conflicts", "error", err, "service_id", serviceID)
		return nil, err
	}

	bookings := make([]*model.Booking, len(rows))
	for i, row := range rows {
		bookings[i] = &model.Booking{
			ID:          fromUUID(row.ID),
			UserID:      fromUUID(row.UserID),
			ServiceID:   fromUUID(row.ServiceID),
			StartTime:   fromTimestamptz(row.StartTime),
			EndTime:     fromTimestamptz(row.EndTime),
			Status:      model.BookingStatus(row.Status),
			TotalPrice:  fromNumeric(row.TotalPrice),
			Notes:       fromText(row.Notes),
			CancelledAt: fromTimestamptzPtr(row.CancelledAt),
			CreatedAt:   fromTimestamptz(row.CreatedAt),
			UpdatedAt:   fromTimestamptz(row.UpdatedAt),
		}
	}

	return bookings, nil
}
