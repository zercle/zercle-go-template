//go:build unit

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, domain.ErrMachineNotFound, domain.ErrMachineNotFound)
	assert.ErrorIs(t, domain.ErrInvalidID, domain.ErrInvalidID)
	assert.ErrorIs(t, domain.ErrInvalidMachineLabel, domain.ErrInvalidMachineLabel)
	assert.ErrorIs(t, domain.ErrUnsupportedCoin, domain.ErrUnsupportedCoin)
}

func TestValidateCoins(t *testing.T) {
	t.Parallel()

	assert.NoError(t, domain.ValidateCoins(domain.Denominations))
	assert.NoError(t, domain.ValidateCoins(nil))

	err := domain.ValidateCoins([]int32{5, 7})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	assert.Contains(t, err.Error(), "7")
}

func TestAddCoins(t *testing.T) {
	t.Parallel()

	bank := domain.CoinBank{5: 1, 25: 2}
	out := domain.AddCoins(bank, []int32{5, 100})

	assert.EqualValues(t, 2, out[5])
	assert.EqualValues(t, 2, out[25])
	assert.EqualValues(t, 1, out[100])
	assert.EqualValues(t, 1, bank[5], "input bank must not be mutated")
}
