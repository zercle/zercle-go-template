package domain

import "time"

// HealthStatus represents the health status of the system.
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database,omitempty"`
	Error     string    `json:"error,omitempty"`
}
