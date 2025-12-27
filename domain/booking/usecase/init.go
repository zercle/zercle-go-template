package usecase

import (
	"github.com/zercle/zercle-go-template/domain/booking"
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the booking use case with dependencies
func Initialize(repo booking.Repository, serviceRepo service.Repository, log *logger.Logger) booking.Usecase {
	return NewBookingUseCase(repo, serviceRepo, log)
}
