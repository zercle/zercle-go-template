package response

import (
	"time"

	"github.com/google/uuid"
)

// ServiceResponse represents a service response
type ServiceResponse struct {
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           float64   `json:"price"`
	MaxCapacity     int       `json:"max_capacity"`
	ID              uuid.UUID `json:"id"`
	IsActive        bool      `json:"is_active"`
}

// ListServicesResponse represents a paginated list of services
type ListServicesResponse struct {
	Services []ServiceResponse `json:"services"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	IsActive bool              `json:"is_active"`
}
