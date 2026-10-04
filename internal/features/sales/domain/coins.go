package domain

import (
	"fmt"
	"maps"
	"slices"
)

// SupportedDenominations lists the coin values the machine accepts, in cents,
// ascending. It is a sorted slice rather than a set so MakeChange can iterate
// it largest-first; membership is a five-element scan, so no parallel set is
// kept that could drift from the list.
var SupportedDenominations = []int32{5, 10, 25, 50, 100}

// CoinBank maps a denomination in cents to the number of coins available for
// change. It is a reference type, so callers who hand a bank to another
// function must assume that function treats it as read-only; the functions in
// this package never mutate a bank they receive.
type CoinBank map[int32]int32

// Clone returns a deep copy of the bank. A plain assignment would alias the
// underlying map, letting a later write in one place silently change the
// other, so every function that needs to modify a bank copies it first.
func (b CoinBank) Clone() CoinBank {
	out := make(CoinBank, len(b))
	maps.Copy(out, b)
	return out
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

// SumCoins returns the total value of coins in cents.
func SumCoins(coins []int32) int32 {
	var total int32
	for _, coin := range coins {
		total += coin
	}
	return total
}

// MakeChange composes amount from bank using the fewest coins and returns the
// coins used alongside the bank with those coins removed. The input bank is
// never mutated. A non-positive amount yields empty change and an equal bank
// copy (it cannot return ErrExactChangeRequired); a machine that cannot pay out
// exactly must refuse the sale.
//
// The denominations {5, 10, 25, 50, 100} form a canonical coin system, so
// greedily taking the largest coin that still fits always composes the exact
// amount whenever one exists; this is why a simple greedy pass is correct here
// rather than requiring dynamic programming.
func MakeChange(bank CoinBank, amount int32) (change []int32, remaining CoinBank, err error) {
	if amount <= 0 {
		return []int32{}, bank.Clone(), nil
	}
	remaining = bank.Clone()
	change = make([]int32, 0, len(SupportedDenominations))
	left := amount
	for _, denom := range slices.Backward(SupportedDenominations) {
		take := left / denom
		if have := remaining[denom]; have < take {
			take = have
		}
		if take <= 0 {
			continue
		}
		for range take {
			change = append(change, denom)
		}
		remaining[denom] -= take
		left -= take * denom
		if left == 0 {
			break
		}
	}
	if left != 0 {
		return nil, nil, ErrExactChangeRequired
	}
	return change, remaining, nil
}

// isSupportedDenomination reports whether coin is one of the accepted values.
func isSupportedDenomination(coin int32) bool {
	return slices.Contains(SupportedDenominations, coin)
}
