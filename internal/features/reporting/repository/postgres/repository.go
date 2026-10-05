package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository/postgres/models"
)

// Repository is a read-only GORM implementation of the repository.Repository
// interface. It owns no schema and never writes.
type Repository struct {
	db *gorm.DB
}

// NewRepository returns a Repository backed by the provided *gorm.DB.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// overviewQuery computes every cross-feature total in one round trip using
// scalar subqueries, so the report is a single statement over the three tables.
//
// The coin-bank subquery sums denomination times count: machines stores
// coin_bank as a JSON object keyed by denomination in cents ({"5":10,"25":4}),
// not as an array, so jsonb_each_text is the shape that yields total cents
// (jsonb_array_elements_text rejects an object).
const overviewQuery = `
SELECT
    (SELECT COUNT(*) FROM catalog_products) AS product_count,
    (SELECT COALESCE(SUM(stock), 0) FROM catalog_products) AS total_stock,
    (SELECT COUNT(*) FROM machines) AS machine_count,
    (SELECT COALESCE(SUM(b.key::bigint * b.value::bigint), 0)
       FROM machines m, jsonb_each_text(m.coin_bank) AS b(key, value)) AS total_bank_cents,
    (SELECT COUNT(*) FROM sales_purchases) AS purchase_count,
    (SELECT COALESCE(SUM(price_cents), 0) FROM sales_purchases) AS revenue_cents
`

// topMachinesQuery ranks machines by revenue and purchase count. The join is
// inner, so a machine with no sales has no leaderboard row. The trailing m.id
// tie-break makes the order deterministic when revenue and count are equal.
const topMachinesQuery = `
SELECT m.id AS machine_id, m.label AS label, COUNT(p.id) AS purchase_count, COALESCE(SUM(p.price_cents), 0) AS revenue_cents
FROM machines m
JOIN sales_purchases p ON p.machine_id = m.id
GROUP BY m.id, m.label
ORDER BY revenue_cents DESC, purchase_count DESC, m.id
LIMIT ?
`

// GetOverview returns the cross-feature totals in one statement. An empty
// database yields a zero-value Overview, never a not-found.
func (r *Repository) GetOverview(ctx context.Context) (domain.Overview, error) {
	var row models.OverviewModel
	if err := r.db.WithContext(ctx).Raw(overviewQuery).Scan(&row).Error; err != nil {
		return domain.Overview{}, fmt.Errorf("get overview: %w", err)
	}
	return row.ToDomain(), nil
}

// GetTopMachines returns up to limit machines ranked by revenue, then purchase
// count, then id. Only machines with at least one sale appear.
func (r *Repository) GetTopMachines(ctx context.Context, limit int32) ([]domain.MachineSales, error) {
	var rows []models.MachineSalesModel
	if err := r.db.WithContext(ctx).Raw(topMachinesQuery, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get top machines: %w", err)
	}

	machines := make([]domain.MachineSales, len(rows))
	for i := range rows {
		machines[i] = rows[i].ToDomain()
	}
	return machines, nil
}
