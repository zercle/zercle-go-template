package usecase

import (
	"github.com/rs/zerolog"
	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// service implements user.Usecase interface.
type service struct {
	repo   user.Repository
	logger zerolog.Logger
}

// NewService creates a new user usecase service.
func NewService(repo user.Repository) user.Usecase {
	return &service{
		repo:   repo,
		logger: zerolog.Nop(),
	}
}

// NewServiceWithLogger creates a new user usecase service with a custom logger.
func NewServiceWithLogger(repo user.Repository, logger zerolog.Logger) user.Usecase {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Compile-time interface check.
var _ user.Usecase = (*service)(nil)
