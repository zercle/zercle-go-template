package service

import (
	"context"
	"time"

	"github.com/zercle/zercle-go-template/internal/core/port"
	healthDto "github.com/zercle/zercle-go-template/internal/features/health/dto"
)

type healthService struct {
	repo port.HealthRepository
}

// NewHealthService creates a new instance of HealthService.
func NewHealthService(repo port.HealthRepository) port.HealthService {
	return &healthService{
		repo: repo,
	}
}

// HealthCheck performs a basic health check.
func (s *healthService) HealthCheck(ctx context.Context) (*healthDto.HealthResponse, error) {
	return &healthDto.HealthResponse{
		Status:    "OK",
		Timestamp: time.Now(),
	}, nil
}

// LivenessCheck performs a liveness check including database connectivity.
func (s *healthService) LivenessCheck(ctx context.Context) (*healthDto.LivenessResponse, error) {
	// Check database connectivity
	dbStatus, err := s.repo.CheckDatabase(ctx)
	if err != nil {
		return &healthDto.LivenessResponse{
			Status:    "down",
			Timestamp: time.Now(),
			Database:  "unreachable",
			Error:     err.Error(),
		}, err
	}

	return &healthDto.LivenessResponse{
		Status:    "alive",
		Timestamp: time.Now(),
		Database:  dbStatus,
	}, nil
}
