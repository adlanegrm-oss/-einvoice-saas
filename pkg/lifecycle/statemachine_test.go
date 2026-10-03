package lifecycle_test

import (
"testing"
"github.com/adlanegrm-oss/einvoice-saas/pkg/lifecycle"
)

func TestLifecycleTransitions(t *testing.T) {
sm := lifecycle.New(lifecycle.StateDraft)

// DRAFT -> DEPOSEE : Valide
if err := sm.TransitionTo(lifecycle.StateDeposited); err != nil {
t.Fatalf("Transition DRAFT -> DEPOSEE refusée: %v", err)
}

// DEPOSEE -> ENCAISSEE : Valide
if err := sm.TransitionTo(lifecycle.StateCollected); err != nil {
t.Fatalf("Transition DEPOSEE -> ENCAISSEE refusée: %v", err)
}

// ENCAISSEE est terminal -> tentative de transition doit échouer
if err := sm.TransitionTo(lifecycle.StateRefused); err == nil {
t.Fatalf("Une transition depuis un état terminal aurait dû échouer")
}
}

func TestPurgePredicate(t *testing.T) {
predicate := lifecycle.PurgeQueryPredicate()
expected := "status IN ('REJETEE', 'REFUSEE', 'ENCAISSEE') AND updated_at < ?"
if predicate != expected {
t.Fatalf("Prédicat de purge invalide.\nAttendu: %s\nReçu: %s", expected, predicate)
}
}
