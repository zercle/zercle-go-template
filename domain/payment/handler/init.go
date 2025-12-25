package handler

import (
	"github.com/zercle/zercle-go-template/domain/payment"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the payment handler with dependencies
func Initialize(usecase payment.Usecase, log *logger.Logger) payment.Handler {
	return NewPaymentHandler(usecase, log)
}
