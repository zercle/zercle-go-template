//go:build unit

package domain_test

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
)

func TestErrInvalidTopMachines(t *testing.T) {
	t.Parallel()

	require.Error(t, domain.ErrInvalidTopMachines)
	assert.ErrorIs(t, domain.ErrInvalidTopMachines, domain.ErrInvalidTopMachines)
}

// TestOverviewZeroValue pins the empty-database report: every total is zero, so
// a read over empty tables is a valid report rather than a missing one.
func TestOverviewZeroValue(t *testing.T) {
	t.Parallel()

	var overview domain.Overview
	assert.Zero(t, overview.ProductCount)
	assert.Zero(t, overview.TotalStock)
	assert.Zero(t, overview.MachineCount)
	assert.Zero(t, overview.TotalBankCents)
	assert.Zero(t, overview.PurchaseCount)
	assert.Zero(t, overview.RevenueCents)
}

// TestMachineSalesFields pins the leaderboard projection: it carries the
// machine identity and the two aggregate values the report exposes, and no
// purchase detail.
func TestMachineSalesFields(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	sales := domain.MachineSales{
		MachineID:     id,
		Label:         "lobby-1",
		PurchaseCount: 3,
		RevenueCents:  450,
	}

	assert.Equal(t, id, sales.MachineID)
	assert.Equal(t, "lobby-1", sales.Label)
	assert.EqualValues(t, 3, sales.PurchaseCount)
	assert.EqualValues(t, 450, sales.RevenueCents)
}
