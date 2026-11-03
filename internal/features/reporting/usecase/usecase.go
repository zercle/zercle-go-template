// Package usecase holds the reporting feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound repository.
//
// Each operation lives in its own file: summary.go. This file holds the shared
// implementation type and its constructor.
package usecase

import (
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository"
)

const (
	defaultTopMachinesFallback int32 = 5
	maxTopMachinesFallback     int32 = 20
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo               repository.Repository
	defaultTopMachines int32
	maxTopMachines     int32
}

// NewUsecase returns a Usecase backed by the provided repository. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (5/20).
func NewUsecase(repo repository.Repository, defaultTopMachines, maxTopMachines int32) *Usecase {
	if defaultTopMachines <= 0 {
		defaultTopMachines = defaultTopMachinesFallback
	}
	if maxTopMachines <= 0 {
		maxTopMachines = maxTopMachinesFallback
	}
	return &Usecase{
		repo:               repo,
		defaultTopMachines: defaultTopMachines,
		maxTopMachines:     maxTopMachines,
	}
}
