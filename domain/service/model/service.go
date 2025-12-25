package model

import (
	"time"

	"github.com/google/uuid"
)

// Service represents a bookable service in the system
type Service struct {
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Description     string
	DurationMinutes int
	Price           float64
	MaxCapacity     int
	ID              uuid.UUID
	IsActive        bool
}
