package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/booking/model"
)

// BookingResponse represents a booking response
type BookingResponse struct {
	ID          uuid.UUID           `json:"id"`
	UserID      uuid.UUID           `json:"user_id"`
	ServiceID   uuid.UUID           `json:"service_id"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	Status      model.BookingStatus `json:"status"`
	TotalPrice  float64             `json:"total_price"`
	Notes       string              `json:"notes"`
	CancelledAt *time.Time          `json:"canceled_at,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// ListBookingsResponse represents a paginated list of bookings
type ListBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}
