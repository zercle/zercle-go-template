package usecase

import (
	"github.com/zercle/zercle-go-template/domain/payment"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the payment use case with dependencies
func Initialize(repo payment.Repository, log *logger.Logger) payment.Usecase {
	return NewPaymentUseCase(repo, log)
}
