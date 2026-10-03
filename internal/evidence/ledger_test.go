package evidence_test

import (
"testing"
"time"

"github.com/adlanegrm-oss/einvoice-saas/internal/evidence"
)

func TestHashChain_SequentialIntegrityAndTamperEvidence(t *testing.T) {
t0 := time.Now().UTC()
prev0 := "0000000000000000000000000000000000000000000000000000000000000000"

h0 := evidence.HashStep(prev0, "payload-hash-1", "INVOICE_INGESTED", t0)
evt0 := evidence.EventLog{Index: 0, EventType: "INVOICE_INGESTED", PayloadHash: "payload-hash-1", PrevHash: prev0, EventHash: h0, Timestamp: t0}

t1 := t0.Add(time.Second)
h1 := evidence.HashStep(h0, "payload-hash-2", "INVOICE_VALIDATED", t1)
evt1 := evidence.EventLog{Index: 1, EventType: "INVOICE_VALIDATED", PayloadHash: "payload-hash-2", PrevHash: h0, EventHash: h1, Timestamp: t1}

t2 := t1.Add(time.Second)
h2 := evidence.HashStep(h1, "payload-hash-3", "INVOICE_SUBMITTED", t2)
evt2 := evidence.EventLog{Index: 2, EventType: "INVOICE_SUBMITTED", PayloadHash: "payload-hash-3", PrevHash: h1, EventHash: h2, Timestamp: t2}

chain := []evidence.EventLog{evt0, evt1, evt2}

// 1. Chaîne valide
valid, idx := evidence.VerifyHashChain(chain)
if !valid {
t.Fatalf("La chaîne valide a échoué à l'index %d", idx)
}

// 2. Altération de contenu historique (payload)
chainCorruptPayload := make([]evidence.EventLog, len(chain))
copy(chainCorruptPayload, chain)
chainCorruptPayload[1].PayloadHash = "payload-altéré"
if ok, fault := evidence.VerifyHashChain(chainCorruptPayload); ok || fault != 1 {
t.Fatalf("L'altération de payload à l'index 1 aurait dû échouer (obtenu ok=%v, fault=%d)", ok, fault)
}

// 3. Altération de séquence (prevHash rompu)
chainCorruptSeq := make([]evidence.EventLog, len(chain))
copy(chainCorruptSeq, chain)
chainCorruptSeq[2].PrevHash = "mauvais-prev-hash"
if ok, fault := evidence.VerifyHashChain(chainCorruptSeq); ok || fault != 2 {
t.Fatalf("La rupture de séquence à l'index 2 aurait dû échouer (obtenu ok=%v, fault=%d)", ok, fault)
}
}
