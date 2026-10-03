package lifecycle

import (
	"fmt"
	"time"
)

// State represents official DGFIP / PDP invoice lifecycle states.
type State string

const (
	StateDraft     State = "BROUILLON" // Internal draft
	StateDeposited State = "DEPOSEE"   // Facture déposée sur la plateforme
	StateRejected  State = "REJETEE"   // Rejet technique ou de conformité (Terminal)
	StateRefused   State = "REFUSEE"   // Refusée par le destinataire/acheteur (Terminal)
	StateCollected State = "ENCAISSEE" // Statut final de paiement/encaissement (Terminal)
)

func (s State) IsTerminal() bool {
	switch s {
	case StateRejected, StateRefused, StateCollected:
		return true
	default:
		return false
	}
}

// StateMachine manages transitions strictly without compatibility aliases.
type StateMachine struct {
	CurrentState State
	UpdatedAt    time.Time
}

func New(initial State) StateMachine {
	return StateMachine{
		CurrentState: initial,
		UpdatedAt:    time.Now().UTC(),
	}
}

// TransitionTo attempts a state change adhering to the French invoice lifecycle.
func (sm *StateMachine) TransitionTo(target State) error {
	if sm.CurrentState.IsTerminal() {
		return fmt.Errorf("illegal transition: state %s is terminal", sm.CurrentState)
	}

	switch sm.CurrentState {
	case StateDraft:
		if target == StateDeposited {
			sm.CurrentState = target
			sm.UpdatedAt = time.Now().UTC()
			return nil
		}
	case StateDeposited:
		if target == StateRejected || target == StateRefused || target == StateCollected {
			sm.CurrentState = target
			sm.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("invalid state transition from %s to %s", sm.CurrentState, target)
}

// PurgeQueryPredicate returns the SQL condition for safe record purging.
// Replaces obsolete 'CLOSED' status with actual terminal states.
func PurgeQueryPredicate() string {
	return "status IN ('REJETEE', 'REFUSEE', 'ENCAISSEE') AND updated_at < ?"
}
