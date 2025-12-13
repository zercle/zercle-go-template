package port

import (
	"context"

	healthDto "github.com/zercle/zercle-go-template/internal/features/health/dto"
)

//go:generate mockgen -destination=./mocks/$GOFILE -package=mocks -source=$GOFILE

// HealthService defines the input port for Health operations.
type HealthService interface {
	HealthCheck(ctx context.Context) (*healthDto.HealthResponse, error)
	LivenessCheck(ctx context.Context) (*healthDto.LivenessResponse, error)
}
