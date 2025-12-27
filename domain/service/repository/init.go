package repository

import (
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// Initialize initializes the service repository with dependencies
func Initialize(sqlcQuery *db.Queries, log *logger.Logger) service.Repository {
	return NewServiceRepository(sqlcQuery, log)
}
