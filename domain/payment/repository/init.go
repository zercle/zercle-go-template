package repository

import (
	"github.com/zercle/zercle-go-template/domain/payment"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// Initialize initializes the payment repository with dependencies
func Initialize(sqlcQuery *db.Queries, log *logger.Logger) payment.Repository {
	return NewPaymentRepository(sqlcQuery, log)
}
