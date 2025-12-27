package usecase

import (
	"github.com/zercle/zercle-go-template/domain/user"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

// Initialize initializes the user use case with dependencies
func Initialize(repo user.Repository, cfg *config.Config, log *logger.Logger) user.Usecase {
	return NewUserUseCase(repo, &cfg.JWT, log)
}
