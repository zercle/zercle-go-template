package domain

import (
	"fmt"
	"maps"
	"slices"
	"time"
	"uuid"
)

// Denominations lists the coin values the machine accepts, in cents,
// ascending. It is duplicated from the sales feature's
// SupportedDenominations on purpose: the architecture gates forbid importing
// another feature's domain package, and the accepted set is small and stable,
// so a second declaration is cheaper than a shared kernel that would couple
// the two features.
var Denominations = []int32{5, 10, 25, 50, 100}

// CoinBank maps a denomination in cents to the number of such coins the
// machine holds. It is a reference type, so callers who hand a bank to another
// function must assume that function treats it as read-only; the functions in
// this package never mutate a bank they receive.
type CoinBank map[int32]int32

// Machine is a registered vending machine and the coin bank it can pay change
// from.
type Machine struct {
	ID        uuid.UUID
	Label     string
	CoinBank  CoinBank
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ValidateCoins reports whether every coin is an accepted denomination. The
// returned error wraps ErrUnsupportedCoin and names the offending value so a
// caller can still match the sentinel with errors.Is while surfacing detail.
func ValidateCoins(coins []int32) error {
	for _, coin := range coins {
		if !isSupportedDenomination(coin) {
			return fmt.Errorf("%w: %d", ErrUnsupportedCoin, coin)
		}
	}
	return nil
}

// AddCoins returns a new bank with every coin in coins added. The input bank is
// never mutated.
func AddCoins(bank CoinBank, coins []int32) CoinBank {
	out := make(CoinBank, len(bank)+len(coins))
	maps.Copy(out, bank)
	for _, coin := range coins {
		out[coin]++
	}
	return out
}

// isSupportedDenomination reports whether coin is one of the accepted values.
func isSupportedDenomination(coin int32) bool {
	return slices.Contains(Denominations, coin)
}
