package audit

import (
"testing"
)

func TestAuditChainIntegrityAndTampering(t *testing.T) {
tenantID := "org_demo"
invoiceID := "inv_2026_001"

// 1. Simuler la création séquentielle de 3 événements
e1 := CreateNextEvent(tenantID, invoiceID, "DEPOSITED", "hash_payload_v1", nil)
e2 := CreateNextEvent(tenantID, invoiceID, "ISSUED", "hash_payload_v1", &e1)
e3 := CreateNextEvent(tenantID, invoiceID, "TRANSMITTED", "hash_payload_v1", &e2)

chain := []AuditEvent{e1, e2, e3}

// 2. Vérifier que la chaîne valide passe
valid, err := VerifyAuditChain(chain)
if !valid || err != nil {
t.Fatalf("La chaîne légitime devrait être valide: %v", err)
}

// 3. Simuler une altération manuelle directe en base (modification d'un hash)
chainCorrompue := make([]AuditEvent, len(chain))
copy(chainCorrompue, chain)
chainCorrompue[1].PayloadHash = "hash_payload_hacked"

validCorrupted, errCorrupted := VerifyAuditChain(chainCorrompue)
if validCorrupted || errCorrupted == nil {
t.Fatalf("La chaîne altérée aurait dû échouer à la vérification!")
}
t.Logf("Succès : Altération interceptée avec succès -> %v", errCorrupted)
}
