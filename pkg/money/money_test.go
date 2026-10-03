package money

import "testing"

func TestRoundHalfEven(t *testing.T) {
	tests := []struct {
		input    float64
		expected int64
	}{
		{2.5, 2},
		{3.5, 4},
		{1.5, 2},
		{4.5, 4},
		{2.5001, 3},
		{2.4999, 2},
		{0.5, 0},
		{-2.5, -2},
		{-3.5, -4},
	}

	for _, tt := range tests {
		got := RoundHalfEven(tt.input)
		if got != tt.expected {
			t.Errorf("RoundHalfEven(%v) = %d; want %d", tt.input, got, tt.expected)
		}
	}
}

func TestMoneyOperations(t *testing.T) {
	m1 := FromFloat(10.50, EUR)
	m2 := FromFloat(5.25, EUR)

	sum, err := m1.Add(m2)
	if err != nil || sum.Cents() != 1575 {
		t.Fatalf("unexpected sum: %v, err: %v", sum, err)
	}

	diff, err := m1.Sub(m2)
	if err != nil || diff.Cents() != 525 {
		t.Fatalf("unexpected diff: %v, err: %v", diff, err)
	}
}
