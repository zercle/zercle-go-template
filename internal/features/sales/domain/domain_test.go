//go:build unit

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, domain.ErrProductNotFound, domain.ErrProductNotFound)
	assert.ErrorIs(t, domain.ErrMachineNotFound, domain.ErrMachineNotFound)
	assert.ErrorIs(t, domain.ErrUnsupportedCoin, domain.ErrUnsupportedCoin)
	assert.ErrorIs(t, domain.ErrInsufficientPayment, domain.ErrInsufficientPayment)
	assert.ErrorIs(t, domain.ErrOutOfStock, domain.ErrOutOfStock)
	assert.ErrorIs(t, domain.ErrExactChangeRequired, domain.ErrExactChangeRequired)
}

func TestValidateCoins(t *testing.T) {
	t.Parallel()

	assert.NoError(t, domain.ValidateCoins(domain.SupportedDenominations))
	assert.NoError(t, domain.ValidateCoins(nil))
	assert.NoError(t, domain.ValidateCoins([]int32{5, 100, 25}))

	err := domain.ValidateCoins([]int32{5, 7})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	assert.Contains(t, err.Error(), "7")
}

func TestSumCoins(t *testing.T) {
	t.Parallel()

	assert.EqualValues(t, 0, domain.SumCoins(nil))
	assert.EqualValues(t, 165, domain.SumCoins([]int32{100, 50, 10, 5}))
}

func TestCoinBank_CloneIsDeep(t *testing.T) {
	t.Parallel()

	bank := domain.CoinBank{5: 2, 25: 1}
	clone := bank.Clone()
	require.Equal(t, bank, clone)

	clone[5] = 99
	assert.EqualValues(t, 2, bank[5], "mutating the clone must not touch the original")
}

func TestMakeChange(t *testing.T) {
	t.Parallel()

	// Ten of each denomination, so change is limited only by the amount.
	fullBank := domain.CoinBank{5: 10, 10: 10, 25: 10, 50: 10, 100: 10}

	tests := []struct {
		name      string
		bank      domain.CoinBank
		amount    int32
		want      []int32
		wantErr   error
		remaining domain.CoinBank
	}{
		{
			name:      "65 uses 50 10 5",
			bank:      fullBank,
			amount:    65,
			want:      []int32{50, 10, 5},
			remaining: domain.CoinBank{5: 9, 10: 9, 25: 10, 50: 9, 100: 10},
		},
		{
			name:      "40 uses 25 10 5",
			bank:      fullBank,
			amount:    40,
			want:      []int32{25, 10, 5},
			remaining: domain.CoinBank{5: 9, 10: 9, 25: 9, 50: 10, 100: 10},
		},
		{
			name:      "15 uses 10 5",
			bank:      fullBank,
			amount:    15,
			want:      []int32{10, 5},
			remaining: domain.CoinBank{5: 9, 10: 9, 25: 10, 50: 10, 100: 10},
		},
		{
			name:      "30 uses 25 5",
			bank:      fullBank,
			amount:    30,
			want:      []int32{25, 5},
			remaining: domain.CoinBank{5: 9, 10: 10, 25: 9, 50: 10, 100: 10},
		},
		{
			name:      "zero returns empty change and equal bank",
			bank:      fullBank,
			amount:    0,
			want:      []int32{},
			remaining: fullBank,
		},
		{
			name:    "unsupported amount cannot be composed exactly",
			bank:    fullBank,
			amount:  3,
			wantErr: domain.ErrExactChangeRequired,
		},
		{
			name:    "bank lacking small coins cannot compose exactly",
			bank:    domain.CoinBank{100: 1},
			amount:  40,
			wantErr: domain.ErrExactChangeRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			before := tc.bank.Clone()
			change, remaining, err := domain.MakeChange(tc.bank, tc.amount)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, change)
				assert.Nil(t, remaining)
				assert.Equal(t, before, tc.bank, "input bank must not be mutated on failure")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, change)
			assert.Equal(t, tc.remaining, remaining)
			assert.EqualValues(t, tc.amount, domain.SumCoins(change))
			assert.Equal(t, before, tc.bank, "input bank must not be mutated on success")
		})
	}
}

func TestPurchase(t *testing.T) {
	t.Parallel()

	fullBank := domain.CoinBank{5: 10, 10: 10, 25: 10, 50: 10, 100: 10}

	tests := []struct {
		name       string
		priceCents int32
		stock      int32
		bank       domain.CoinBank
		coins      []int32
		wantChange []int32
		wantErr    error
	}{
		{
			name:       "happy path returns change",
			priceCents: 100,
			stock:      5,
			bank:       fullBank,
			coins:      []int32{100, 25, 10},
			wantChange: []int32{25, 10},
		},
		{
			name:       "exact payment returns no change",
			priceCents: 100,
			stock:      5,
			bank:       fullBank,
			coins:      []int32{100},
			wantChange: []int32{},
		},
		{
			name:       "insufficient payment",
			priceCents: 100,
			stock:      5,
			bank:       fullBank,
			coins:      []int32{50},
			wantErr:    domain.ErrInsufficientPayment,
		},
		{
			name:       "out of stock",
			priceCents: 100,
			stock:      0,
			bank:       fullBank,
			coins:      []int32{100},
			wantErr:    domain.ErrOutOfStock,
		},
		{
			name:       "unsupported coin rejected before totals",
			priceCents: 100,
			stock:      5,
			bank:       fullBank,
			coins:      []int32{7},
			wantErr:    domain.ErrUnsupportedCoin,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			beforeBank := tc.bank.Clone()

			change, remaining, err := domain.Purchase(tc.priceCents, tc.stock, tc.bank, tc.coins)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, change)
				assert.Nil(t, remaining)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantChange, change)
				require.NotNil(t, remaining)
				// The change paid out must sum to the overpayment.
				wantChangeCents := domain.SumCoins(tc.coins) - tc.priceCents
				assert.EqualValues(t, wantChangeCents, domain.SumCoins(change))
				assert.EqualValues(t, tc.priceCents+domain.SumCoins(change), domain.SumCoins(tc.coins))
			}

			assert.Equal(t, beforeBank, tc.bank, "input bank must not be mutated")
		})
	}
}
