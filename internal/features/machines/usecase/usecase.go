// Package usecase holds the machines feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound port.
//
// Each operation lives in its own file: create_machine.go, get_machine.go,
// list_machines.go, restock_bank.go. This file holds the shared implementation
// type and its constructor.
package usecase

import (
	"github.com/zercle/zercle-go-template/internal/features/machines/port"
)

const (
	defaultPageSizeFallback int32 = 20
	maxPageSizeFallback     int32 = 100
	maxLabelLengthFallback  int32 = 255
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo            port.Repository
	defaultPageSize int32
	maxPageSize     int32
	maxLabelLength  int32
}

// NewUsecase returns a Usecase backed by the provided port. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (20/100/255).
func NewUsecase(repo port.Repository, defaultPageSize, maxPageSize, maxLabelLength int32) *Usecase {
	if defaultPageSize <= 0 {
		defaultPageSize = defaultPageSizeFallback
	}
	if maxPageSize <= 0 {
		maxPageSize = maxPageSizeFallback
	}
	if maxLabelLength <= 0 {
		maxLabelLength = maxLabelLengthFallback
	}
	return &Usecase{
		repo:            repo,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
		maxLabelLength:  maxLabelLength,
	}
}
