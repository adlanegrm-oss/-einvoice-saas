package repository

import "testing"

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     InvoiceStatus
		to       InvoiceStatus
		expected bool
	}{
		{"Received -> Validating", StatusReceived, StatusValidating, true},
		{"Received -> Validated (Invalid skip)", StatusReceived, StatusValidated, false},
		{"Validating -> Validated", StatusValidating, StatusValidated, true},
		{"Validating -> Rejected", StatusValidating, StatusRejected, true},
		{"Validated -> Submitted", StatusValidated, StatusSubmitted, true},
		{"Validated -> Rejected (Invalid backwards)", StatusValidated, StatusRejected, false},
		{"Submitted -> Cleared", StatusSubmitted, StatusCleared, true},
		{"Submitted -> Rejected", StatusSubmitted, StatusRejected, true},
		{"Cleared -> Delivered", StatusCleared, StatusDelivered, true},
		{"Delivered -> Validating (Terminal)", StatusDelivered, StatusValidating, false},
		{"Rejected -> Validated (Terminal)", StatusRejected, StatusValidated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTransition(tt.from, tt.to)
			if got != tt.expected {
				t.Errorf("ValidateTransition(%s, %s) = %v; attendu %v", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}
