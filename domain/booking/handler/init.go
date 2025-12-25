package handler

import (
	"github.com/zercle/zercle-go-template/domain/booking"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the booking handler with dependencies
func Initialize(usecase booking.Usecase, log *logger.Logger) booking.Handler {
	return NewBookingHandler(usecase, log)
}
