package handler

import (
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the service handler with dependencies
func Initialize(usecase service.Usecase, log *logger.Logger) service.Handler {
	return NewServiceHandler(usecase, log)
}
