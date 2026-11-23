// Package models defines read-only projections over tables owned by other
// features. The reporting repository scans its aggregate queries into these row
// shapes; it never writes them.
//
// This is a deliberate single-database compromise: the reporting feature reads
// catalog_products, machines, and sales_purchases directly through its own port
// rather than calling the other features' usecases. In a truly distributed
// deployment this becomes cross-service calls or a read model, and this package
// is already the seam where that swap happens. The reporting feature declares
// its own row shapes instead of importing the catalog, machines, or sales
// packages, so no feature depends on another feature's persistence internals.
package models

import (
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
)

// OverviewModel is the scan target of the single cross-feature aggregate query.
// Its fields map one-to-one to that query's column aliases.
type OverviewModel struct {
	ProductCount   int64
	TotalStock     int64
	MachineCount   int64
	TotalBankCents int64
	PurchaseCount  int64
	RevenueCents   int64
}

// ToDomain maps the overview projection to the domain aggregate.
func (m OverviewModel) ToDomain() domain.Overview {
	return domain.Overview{
		ProductCount:   m.ProductCount,
		TotalStock:     m.TotalStock,
		MachineCount:   m.MachineCount,
		TotalBankCents: m.TotalBankCents,
		PurchaseCount:  m.PurchaseCount,
		RevenueCents:   m.RevenueCents,
	}
}

// MachineSalesModel is the scan target of one top-machines leaderboard row.
type MachineSalesModel struct {
	MachineID     uuid.UUID
	Label         string
	PurchaseCount int64
	RevenueCents  int64
}

// ToDomain maps the leaderboard projection to the domain rollup.
func (m MachineSalesModel) ToDomain() domain.MachineSales {
	return domain.MachineSales{
		MachineID:     m.MachineID,
		Label:         m.Label,
		PurchaseCount: m.PurchaseCount,
		RevenueCents:  m.RevenueCents,
	}
}
