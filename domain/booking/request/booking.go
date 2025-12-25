package request

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/booking/model"
)

// CreateBooking represents a request to create a new booking
type CreateBooking struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	Notes     string    `json:"notes" validate:"max=500"`
	ServiceID uuid.UUID `json:"service_id" validate:"required"`
}

// UpdateBookingStatus represents a request to update booking status
type UpdateBookingStatus struct {
	Status model.BookingStatus `json:"status" validate:"required,oneof=pending confirmed completed canceled"`
}

// ListBookings represents a request to list bookings with pagination
type ListBookings struct {
	Status string `query:"status" validate:"omitempty,oneof=pending confirmed completed canceled"`
	Limit  int    `query:"limit" validate:"min=1,max=100"`
	Offset int    `query:"offset" validate:"min=0"`
}

// ListBookingsByDateRange represents a request to list bookings within a date range
type ListBookingsByDateRange struct {
	StartDate time.Time `query:"start_date" validate:"required"`
	EndDate   time.Time `query:"end_date" validate:"required"`
	Limit     int       `query:"limit" validate:"min=1,max=100"`
	Offset    int       `query:"offset" validate:"min=0"`
}
