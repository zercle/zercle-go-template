// Package contract holds the reporting feature's inbound wire types. They are
// dependency-free so the published pkg/api/v1 facade can alias them without
// dragging in server internals.
package contract

// SummaryRequest carries the top-machines parameter for GET /reports/summary.
// It is bound from the GET query string (the query tag); the json tag exists so
// the type still marshals to the published shape as part of pkg/api/v1, not
// because the handler reads a JSON body. omitempty lets an absent parameter
// fall through to the configured default, while an explicit value must be at
// least one. The upper bound is deployment-configurable and enforced in the
// usecase layer, so a hardcoded max= tag would drift.
type SummaryRequest struct {
	TopMachines int32 `json:"top" query:"top" validate:"omitempty,min=1"`
}

// CatalogStats is the catalog table's contribution to the report.
type CatalogStats struct {
	ProductCount int64 `json:"product_count"`
	TotalStock   int64 `json:"total_stock"`
}

// MachineStats is the machines table's contribution to the report.
type MachineStats struct {
	MachineCount       int64 `json:"machine_count"`
	TotalCoinBankCents int64 `json:"total_coin_bank_cents"`
}

// SalesStats is the sales table's contribution to the report.
type SalesStats struct {
	PurchaseCount int64 `json:"purchase_count"`
	RevenueCents  int64 `json:"revenue_cents"`
}

// MachineSales is one machine's sales aggregate in the leaderboard.
type MachineSales struct {
	MachineID     string `json:"machine_id"`
	Label         string `json:"label"`
	PurchaseCount int64  `json:"purchase_count"`
	RevenueCents  int64  `json:"revenue_cents"`
}

// SummaryResponse is the cross-feature report: one block per source table plus
// the top machines by revenue.
type SummaryResponse struct {
	Catalog     CatalogStats   `json:"catalog"`
	Machines    MachineStats   `json:"machines"`
	Sales       SalesStats     `json:"sales"`
	TopMachines []MachineSales `json:"top_machines"`
}
