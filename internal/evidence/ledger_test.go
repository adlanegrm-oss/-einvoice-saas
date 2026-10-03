package evidence_test

import (
"testing"
"time"

"github.com/adlanegrm-oss/einvoice-saas/internal/evidence"
)

func TestEvidenceLedger_TamperDetection(t *testing.T) {
t0 := time.Now().UTC()
prev0 := "0000000000000000000000000000000000000000000000000000000000000000"

h0 := evidence.HashStep(prev0, "payload-v1", "INGESTION", t0)
evt0 := evidence.EventLog{Index: 0, EventType: "INGESTION", PayloadHash: "payload-v1", PrevHash: prev0, EventHash: h0, Timestamp: t0}

t1 := t0.Add(time.Second)
h1 := evidence.HashStep(h0, "sealed-hash", "SEALED", t1)
evt1 := evidence.EventLog{Index: 1, EventType: "SEALED", PayloadHash: "sealed-hash", PrevHash: h0, EventHash: h1, Timestamp: t1}

chain := []evidence.EventLog{evt0, evt1}

valid, _ := evidence.VerifyMerkleChain(chain)
if !valid {
t.Fatal("La chaîne intègre a été signalée comme invalide")
}

// Altération simulée du payload historique
chain[0].PayloadHash = "payload-altéré-frauduleux"
validTampered, faultIdx := evidence.VerifyMerkleChain(chain)
if validTampered {
t.Fatal("Échec critique : l'altération de l'événement historique n'a pas été interceptée !")
}
if faultIdx != 0 {
t.Fatalf("Indice d'erreur incorrect : attendu 0, obtenu %d", faultIdx)
}
}
