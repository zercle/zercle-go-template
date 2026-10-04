package fixtures

import (
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// NewMachine returns a sample machines Machine with the given label and coin
// bank. It uses a deterministic generated UUID for the ID and fixed timestamps
// so tests can assert against known values.
func NewMachine(label string, bank domain.CoinBank) domain.Machine {
	return domain.Machine{
		ID:        uuid.MustParse("32345678-1234-1234-1234-123456789abc"),
		Label:     label,
		CoinBank:  bank,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// NewBank returns a sample machines CoinBank holding ten of every supported
// denomination, deep enough to compose change for any single purchase.
func NewBank() domain.CoinBank {
	bank := make(domain.CoinBank, len(domain.Denominations))
	for _, denom := range domain.Denominations {
		bank[denom] = 10
	}
	return bank
}

// Coins returns the given values as a coin slice, for readable test inputs.
func Coins(values ...int32) []int32 {
	return append([]int32{}, values...)
}
