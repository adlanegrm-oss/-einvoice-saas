package repository

import (
	"testing"
)

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     InvoiceStatus
		to       InvoiceStatus
		expected bool
	}{
		// Transitions valides
		{"Received -> Validating", StatusReceived, StatusValidating, true},
		{"Validating -> Validated", StatusValidating, StatusValidated, true},
		{"Validating -> Rejected", StatusValidating, StatusRejected, true},
		{"Validated -> Submitted", StatusValidated, StatusSubmitted, true},
		{"Submitted -> Cleared", StatusSubmitted, StatusCleared, true},
		{"Submitted -> Rejected", StatusSubmitted, StatusRejected, true},
		{"Cleared -> Delivered", StatusCleared, StatusDelivered, true},

		// Transitions invalides
		{"Received -> Validated (saut interdit)", StatusReceived, StatusValidated, false},
		{"Received -> Rejected", StatusReceived, StatusRejected, false},
		{"Received -> Submitted", StatusReceived, StatusSubmitted, false},
		{"Validated -> Rejected (retour arrière)", StatusValidated, StatusRejected, false},
		{"Validated -> Validating", StatusValidated, StatusValidating, false},
		{"Validated -> Cleared (saut)", StatusValidated, StatusCleared, false},
		{"Cleared -> Rejected", StatusCleared, StatusRejected, false},
		{"Cleared -> Submitted", StatusCleared, StatusSubmitted, false},
		{"Delivered -> Validating (terminal)", StatusDelivered, StatusValidating, false},
		{"Delivered -> Rejected (terminal)", StatusDelivered, StatusRejected, false},
		{"Delivered -> Cleared (terminal)", StatusDelivered, StatusCleared, false},
		{"Rejected -> Validated (terminal)", StatusRejected, StatusValidated, false},
		{"Rejected -> Validating (terminal)", StatusRejected, StatusValidating, false},
		{"Rejected -> Submitted (terminal)", StatusRejected, StatusSubmitted, false},

		// Même statut
		{"Received -> Received", StatusReceived, StatusReceived, false},
		{"Validated -> Validated", StatusValidated, StatusValidated, false},
		{"Delivered -> Delivered", StatusDelivered, StatusDelivered, false},

		// Statut inconnu
		{"Statut inconnu -> Validating", InvoiceStatus("UNKNOWN"), StatusValidating, false},
		{"Received -> Statut inconnu", StatusReceived, InvoiceStatus("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTransition(tt.from, tt.to)
			if got != tt.expected {
				t.Errorf("ValidateTransition(%q, %q) = %v; attendu %v", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}

func TestInvoiceStatusConstants(t *testing.T) {
	if StatusReceived != "RECEIVED" {
		t.Errorf("StatusReceived = %q; attendu RECEIVED", StatusReceived)
	}
	if StatusDelivered != "DELIVERED" {
		t.Errorf("StatusDelivered = %q; attendu DELIVERED", StatusDelivered)
	}
	if StatusRejected != "REJECTED" {
		t.Errorf("StatusRejected = %q; attendu REJECTED", StatusRejected)
	}
}

func TestErrorVariables(t *testing.T) {
	if ErrNotFound == nil {
		t.Error("ErrNotFound ne doit pas être nil")
	}
	if ErrDuplicateBusiness == nil {
		t.Error("ErrDuplicateBusiness ne doit pas être nil")
	}
	if ErrDuplicatePayload == nil {
		t.Error("ErrDuplicatePayload ne doit pas être nil")
	}
	if ErrInvalidTransition == nil {
		t.Error("ErrInvalidTransition ne doit pas être nil")
	}
	if ErrNotFound == ErrDuplicateBusiness {
		t.Error("ErrNotFound et ErrDuplicateBusiness ne doivent pas être identiques")
	}
}
