package repository

import (
	"github.com/zercle/zercle-go-template/domain/booking"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// Initialize initializes the booking repository with dependencies
func Initialize(sqlcQuery *db.Queries, log *logger.Logger) booking.Repository {
	return NewBookingRepository(sqlcQuery, log)
}
