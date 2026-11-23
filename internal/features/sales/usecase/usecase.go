// Package usecase holds the sales feature's inbound use-case service and its
// implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound port.
//
// Each operation lives in its own file: purchase.go. This file holds the
// shared implementation type and its constructor.
package usecase

import (
	"github.com/zercle/zercle-go-template/internal/features/sales/port"
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo port.Repository
}

// NewUsecase returns a Usecase backed by the provided port.
func NewUsecase(repo port.Repository) *Usecase {
	return &Usecase{repo: repo}
}
