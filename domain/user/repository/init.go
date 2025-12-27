package repository

import (
	"github.com/zercle/zercle-go-template/domain/user"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// Initialize initializes the user repository with dependencies
func Initialize(sqlcQuery *db.Queries, log *logger.Logger) user.Repository {
	return NewUserRepository(sqlcQuery, log)
}
