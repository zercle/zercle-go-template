package handler

import (
	"github.com/zercle/zercle-go-template/domain/user"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the user handler with dependencies
func Initialize(usecase user.Usecase, log *logger.Logger) user.Handler {
	return NewUserHandler(usecase, log)
}
