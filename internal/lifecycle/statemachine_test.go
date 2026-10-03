package lifecycle_test

import (
"testing"

"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle"
)

func TestStateTransitions_AntiReculEtTransitionsIllegales(t *testing.T) {
illegalTransitions := []struct {
from lifecycle.State
to   lifecycle.State
desc string
}{
{lifecycle.StateAccepted, lifecycle.StateDraft, "ACCEPTED -> DRAFT interdit"},
{lifecycle.StateDelivered, lifecycle.StateValidated, "DELIVERED -> VALIDATED interdit"},
{lifecycle.StateSealed, lifecycle.StateDraft, "SEALED -> DRAFT interdit"},
{lifecycle.StateSubmitted, lifecycle.StateDraft, "SUBMITTED -> DRAFT interdit"},
{lifecycle.StateArchived, lifecycle.StateDraft, "ARCHIVED -> DRAFT interdit"},
{lifecycle.StateAccepted, lifecycle.StateAccepted, "Auto-transition ACCEPTED -> ACCEPTED interdite"},
{lifecycle.StateDraft, lifecycle.StateSubmitted, "Saut d'étape direct DRAFT -> SUBMITTED interdit"},
{lifecycle.StateValidationRejected, lifecycle.StateQueued, "Facture rejetée ne peut pas être mise en file"},
}

for _, tt := range illegalTransitions {
t.Run(tt.desc, func(t *testing.T) {
err := lifecycle.CanTransition(tt.from, tt.to)
if err == nil {
t.Fatalf("SÉCURITÉ VIOLÉE : %s a été autorisée !", tt.desc)
}
})
}

legalPath := []struct {
from lifecycle.State
to   lifecycle.State
}{
{lifecycle.StateDraft, lifecycle.StateValidating},
{lifecycle.StateValidating, lifecycle.StateValidated},
{lifecycle.StateValidated, lifecycle.StateSealed},
{lifecycle.StateSealed, lifecycle.StateQueued},
{lifecycle.StateQueued, lifecycle.StateSubmitted},
{lifecycle.StateSubmitted, lifecycle.StateAccepted},
{lifecycle.StateAccepted, lifecycle.StateDelivered},
{lifecycle.StateDelivered, lifecycle.StateArchived},
}

for _, step := range legalPath {
if err := lifecycle.CanTransition(step.from, step.to); err != nil {
t.Fatalf("Transition légale %s -> %s refusée: %v", step.from, step.to, err)
}
}
}
