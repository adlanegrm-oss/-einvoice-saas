package money

import (
	"fmt"
	"math"
)

// Currency represents ISO 4217 3-letter currency code
type Currency string

const EUR Currency = "EUR"

// Money represents a monetary amount stored as an integer (in minor units / cents, e.g. 100 = 1.00 EUR).
// Prevents IEEE-754 binary floating-point representation and rounding drift.
type Money struct {
	amount   int64
	currency Currency
}

// New creates a Money instance from minor units (cents).
func New(cents int64, curr Currency) Money {
	if curr == "" {
		curr = EUR
	}
	return Money{amount: cents, currency: curr}
}

// FromFloat creates Money using strict Banker's Rounding (Half-Even).
func FromFloat(val float64, curr Currency) Money {
	if curr == "" {
		curr = EUR
	}
	return Money{
		amount:   RoundHalfEven(val * 100.0),
		currency: curr,
	}
}

// Cents returns the raw minor units.
func (m Money) Cents() int64 {
	return m.amount
}

// Currency returns the currency code.
func (m Money) Currency() Currency {
	return m.currency
}

// Float64 returns the floating point representation (for display only).
func (m Money) Float64() float64 {
	return float64(m.amount) / 100.0
}

// String prints the formatted amount with 2 decimal places.
func (m Money) String() string {
	sign := ""
	amt := m.amount
	if amt < 0 {
		sign = "-"
		amt = -amt
	}
	return fmt.Sprintf("%s%d.%02d %s", sign, amt/100, amt%100, m.currency)
}

// Add adds two Money amounts with currency check.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("currency mismatch: %s != %s", m.currency, other.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Sub subtracts two Money amounts with currency check.
func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("currency mismatch: %s != %s", m.currency, other.currency)
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// MulQuantity computes Line Net Amount (BT-131) from quantity and unit price in 4 decimals base (10^4).
// Formula: RoundHalfEven( (qty * unitPriceBase10k) / 10000 )
func MulQuantity(qty float64, unitPriceCents4Dec int64, curr Currency) Money {
	// Exact calculation using integer base scaled by quantity
	rawCents := float64(unitPriceCents4Dec) * qty / 100.0
	return Money{
		amount:   RoundHalfEven(rawCents),
		currency: curr,
	}
}

// RoundHalfEven implements IEEE 754 half-even (banker's) rounding to nearest integer.
func RoundHalfEven(val float64) int64 {
	floor := math.Floor(val)
	diff := val - floor

	if diff > 0.5 {
		return int64(floor) + 1
	}
	if diff < 0.5 {
		return int64(floor)
	}

	// diff == 0.5: tie-breaker -> round to nearest even number
	intFloor := int64(floor)
	if intFloor%2 == 0 {
		return intFloor
	}
	return intFloor + 1
}
