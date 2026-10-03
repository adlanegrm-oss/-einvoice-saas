package status

import "testing"

func TestStateMachine_Transitions(t *testing.T) {
	sm := NewStateMachine()

	validCases := [][2]InvoiceState{
		{StateDeposited, StateIssued},
		{StateIssued, StateTransmitted},
		{StateTransmitted, StateReceived},
		{StateReceived, StateApproved},
		{StateApproved, StatePaid},
		{StateReceived, StateSuspended},
		{StateSuspended, StateApproved},
		{StateSuspended, StateRejectedByBuyer},
	}

	for _, tc := range validCases {
		if err := sm.ValidateTransition(tc[0], tc[1]); err != nil {
			t.Errorf("transition valide refusée : %s -> %s (%v)", tc[0], tc[1], err)
		}
	}

	invalidCases := [][2]InvoiceState{
		{StateDeposited, StatePaid},
		{StateTransmitted, StatePaid},
		{StateRejected, StateApproved},
		{StatePaid, StateSuspended},
		{StateRejectedByBuyer, StatePaid},
	}

	for _, tc := range invalidCases {
		if err := sm.ValidateTransition(tc[0], tc[1]); err == nil {
			t.Errorf("transition invalide acceptée : %s -> %s", tc[0], tc[1])
		}
	}
}
