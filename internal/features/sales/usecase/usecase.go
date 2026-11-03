// Package usecase holds the sales feature's inbound use-case service and its
// implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound repository.
//
// Each operation lives in its own file: purchase.go. This file holds the
// shared implementation type and its constructor.
package usecase

import (
	"github.com/zercle/zercle-go-template/internal/features/sales/repository"
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo repository.Repository
}

// NewUsecase returns a Usecase backed by the provided repository.
func NewUsecase(repo repository.Repository) *Usecase {
	return &Usecase{repo: repo}
}
