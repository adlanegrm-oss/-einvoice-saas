package lifecycle_test

import (
"testing"

"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle"
)

func TestStateTransitions_AntiReculFiscal(t *testing.T) {
// 1. Transition légitime
if err := lifecycle.CanTransition(lifecycle.StateDraft, lifecycle.StateValidating); err != nil {
t.Fatalf("Transition valide refusée: %v", err)
}

// 2. Interdiction formelle de recul : SUBMITTED -> DRAFT
if err := lifecycle.CanTransition(lifecycle.StateSubmitted, lifecycle.StateDraft); err == nil {
t.Fatal("Violation de sécurité fiscale : le retour d'un état SUBMITTED vers DRAFT a été autorisé !")
}

// 3. Immuabilité d'un état d'archive
if err := lifecycle.CanTransition(lifecycle.StateArchived, lifecycle.StateValidating); err == nil {
t.Fatal("Violation : un état archivé ne doit plus jamais être modifié !")
}
}
