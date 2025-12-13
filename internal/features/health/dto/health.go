package dto

import "time"

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database,omitempty"`
}

// LivenessResponse represents the liveness check response.
type LivenessResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database,omitempty"`
	Error     string    `json:"error,omitempty"`
}
