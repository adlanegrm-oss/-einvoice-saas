package lifecycle

import (
"errors"
"fmt"
)

type State string

const (
StateDraft              State = "DRAFT"
StateValidating         State = "VALIDATING"
StateValidated          State = "VALIDATED"
StateValidationRejected State = "VALIDATION_REJECTED"
StateSealed             State = "SEALED"
StateQueued             State = "QUEUED"
StateSubmitted          State = "SUBMITTED"
StateAccepted           State = "ACCEPTED"
StateTransmissionError  State = "TRANSMISSION_ERROR"
StateDelivered          State = "DELIVERED"
StateArchived           State = "ARCHIVED"
StateCreditNote         State = "CREDIT_NOTE"
)

var (
ErrInvalidTransition = errors.New("transition d'état métier interdite")
ErrTerminalState     = errors.New("l'état est final et immuable : aucune modification permise")
)

// AllowedTransitions définit les transitions orientées et irréversibles
var allowedTransitions = map[State][]State{
StateDraft:              {StateValidating},
StateValidating:         {StateValidated, StateValidationRejected},
StateValidationRejected: {}, // État terminal de validation
StateValidated:          {StateSealed},
StateSealed:             {StateQueued},
StateQueued:             {StateSubmitted},
StateSubmitted:          {StateAccepted, StateTransmissionError},
StateTransmissionError:  {StateQueued}, // Reprise d'acheminement possible
StateAccepted:           {StateDelivered, StateCreditNote},
StateDelivered:          {StateArchived, StateCreditNote},
StateCreditNote:         {StateArchived},
StateArchived:           {}, // État terminal légal
}

// CanTransition vérifie formellement la possibilité de passer d'un état à l'autre
func CanTransition(from, to State) error {
allowed, exists := allowedTransitions[from]
if !exists || len(allowed) == 0 {
return fmt.Errorf("%w: l'état %s ne peut plus évoluer", ErrTerminalState, from)
}

for _, target := range allowed {
if target == to {
return nil
}
}
return fmt.Errorf("%w: impossible de passer de %s à %s", ErrInvalidTransition, from, to)
}
