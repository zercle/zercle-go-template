package model

import (
	"time"

	"github.com/google/uuid"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	// BookingStatusPending indicates a booking is awaiting confirmation
	BookingStatusPending BookingStatus = "pending"
	// BookingStatusConfirmed indicates a booking has been confirmed
	BookingStatusConfirmed BookingStatus = "confirmed"
	// BookingStatusCompleted indicates a booking has been completed
	BookingStatusCompleted BookingStatus = "completed"
	// BookingStatusCancelled indicates a booking has been canceled
	BookingStatusCancelled BookingStatus = "canceled"
)

// Booking represents a service booking
type Booking struct {
	StartTime   time.Time
	EndTime     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CancelledAt *time.Time
	Status      BookingStatus
	Notes       string
	TotalPrice  float64
	ID          uuid.UUID
	UserID      uuid.UUID
	ServiceID   uuid.UUID
}
