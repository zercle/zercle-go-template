package db

import (
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

func NewDatabase(cfg *config.DatabaseConfig) (Database, error) {
	return NewPostgresDatabase(cfg)
}
