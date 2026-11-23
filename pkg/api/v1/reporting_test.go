//go:build unit

package apiv1_test

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

func TestSummaryRequest_Validation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	// Absent (zero) falls through to the configured default; an explicit value
	// must be at least one.
	assert.NoError(t, v.Struct(apiv1.SummaryRequest{}))
	assert.NoError(t, v.Struct(apiv1.SummaryRequest{TopMachines: 1}))
	assert.NoError(t, v.Struct(apiv1.SummaryRequest{TopMachines: 5}))

	assert.Error(t, v.Struct(apiv1.SummaryRequest{TopMachines: -1}))
	assert.Error(t, v.Struct(apiv1.SummaryRequest{TopMachines: -20}))
}

func TestReportingResponseJSONFieldNames(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(apiv1.SummaryResponse{
		Catalog:  apiv1.CatalogStats{ProductCount: 2, TotalStock: 8},
		Machines: apiv1.MachineStats{MachineCount: 1, TotalCoinBankCents: 350},
		Sales:    apiv1.SalesStats{PurchaseCount: 4, RevenueCents: 600},
		TopMachines: []apiv1.MachineSales{
			{MachineID: "m1", Label: "lobby", PurchaseCount: 4, RevenueCents: 600},
		},
	})
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"catalog"`)
	assert.Contains(t, string(raw), `"product_count"`)
	assert.Contains(t, string(raw), `"total_stock"`)
	assert.Contains(t, string(raw), `"machines"`)
	assert.Contains(t, string(raw), `"machine_count"`)
	assert.Contains(t, string(raw), `"total_coin_bank_cents"`)
	assert.Contains(t, string(raw), `"sales"`)
	assert.Contains(t, string(raw), `"purchase_count"`)
	assert.Contains(t, string(raw), `"revenue_cents"`)
	assert.Contains(t, string(raw), `"top_machines"`)
	assert.Contains(t, string(raw), `"machine_id"`)
}
