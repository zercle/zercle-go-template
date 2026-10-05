// Package repository declares the reporting feature's outbound (driven)
// interface. The usecase layer consumes it; the gorm implementation under
// repository/postgres satisfies it structurally.
package repository

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
)

// Repository is the outbound interface for reporting. It reads the catalog,
// machines, and sales tables through its own port (see models.refs) rather
// than calling the other features' usecases. That is the same deliberate
// single-database compromise the sales feature documents: the tables are
// shared, but this port is the seam where a future service split replaces the
// implementation without touching the reporting domain or usecase.
//
//go:generate go tool mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type Repository interface {
	// GetOverview returns the cross-feature totals in one round trip. A
	// read-only aggregate over empty tables yields a zero-value Overview, not
	// an error.
	GetOverview(ctx context.Context) (domain.Overview, error)
	// GetTopMachines returns up to limit machines ranked by revenue (then
	// purchase count, then id) with their purchase rollup. It never reports a
	// not-found: a machine with no sales appears with zero totals.
	GetTopMachines(ctx context.Context, limit int32) ([]domain.MachineSales, error)
}
