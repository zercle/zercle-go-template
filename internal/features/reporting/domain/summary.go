package domain

import "uuid"

// Overview is the reporting feature's cross-feature aggregate: one read of
// totals from the catalog, machines, and sales tables. The zero value is the
// empty-database report, so an absent table contributes zero rather than an
// error.
type Overview struct {
	ProductCount   int64
	TotalStock     int64
	MachineCount   int64
	TotalBankCents int64
	PurchaseCount  int64
	RevenueCents   int64
}

// MachineSales is one machine's sales rollup for the top-machines leaderboard.
// It carries only the fields a report reads; it is a projection, not the
// machines feature's entity.
type MachineSales struct {
	MachineID     uuid.UUID
	Label         string
	PurchaseCount int64
	RevenueCents  int64
}
