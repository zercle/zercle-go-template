package usecase

import (
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the service use case with dependencies
func Initialize(repo service.Repository, log *logger.Logger) service.Usecase {
	return NewServiceUseCase(repo, log)
}
