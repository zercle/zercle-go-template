package usecase

import (
	"github.com/samber/do/v2"
	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// service implements task.Usecase interface.
type service struct {
	repo task.Repository
}

// NewService creates a new task usecase service using samber/do v2 dependency injection.
func NewService(i do.Injector) (task.Usecase, error) {
	repo := do.MustInvoke[task.Repository](i)
	return &service{repo: repo}, nil
}

// Compile-time interface check.
var _ task.Usecase = (*service)(nil)
