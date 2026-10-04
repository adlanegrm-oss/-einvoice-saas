package lifecycle

import (
	"errors"
	"testing"
)

func TestStateTransitions(t *testing.T) {
	s, err := TransitionStatus(StatusDraft, StatusValidating)
	if err != nil || s != StatusValidating {
		t.Fatalf("attendu VALIDATING, obtenu %s (%v)", s, err)
	}

	s, err = TransitionStatus(StatusAccepted, StatusIssued)
	if err != nil || s != StatusIssued {
		t.Fatalf("attendu ISSUED, obtenu %s (%v)", s, err)
	}

	_, err = TransitionStatus(StatusIssued, StatusDraft)
	if !errors.Is(err, ErrInvoiceImmutable) {
		t.Fatalf("attendu ErrInvoiceImmutable, obtenu %v", err)
	}

	_, err = TransitionStatus(StatusDraft, StatusIssued)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("attendu ErrInvalidTransition, obtenu %v", err)
	}
}
