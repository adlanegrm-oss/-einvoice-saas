package test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/clearance"
	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/as4"
)

func TestE2E_FullPipeline_Sprint3(t *testing.T) {
	ctx := context.Background()

	// 1. Initialisation des composants
	fsm := status.NewStateMachine()
	validator := en16931.NewSchematronValidator()
	ppf := clearance.NewPPFConnector(clearance.PPFConfig{
		BaseURL:  "https://api.sandbox.chorus-pro.gouv.fr",
		ClientID: "test-pdp-client",
	})
	as4Repo := as4.NewInMemoryOutboxRepository()
	as4Client := as4.NewAS4Client()
	as4Worker := as4.NewOutboxWorker(as4Repo, as4Client, as4.WorkerConfig{
		BatchSize:      10,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     200 * time.Millisecond,
	})

	// Journal d'audit en mémoire pour la PAF
	var auditTrail []*status.StatusEvent
	currentHash := "GENESIS"

	recordTransition := func(invID string, from, to status.InvoiceState, actor, reason, payloadHex string) {
		evt, err := fsm.Transition(invID, from, to, actor, reason, payloadHex, currentHash)
		if err != nil {
			t.Fatalf("Transition FSM illégale de %s vers %s : %v", from, to, err)
		}
		if evt.Hash != evt.ComputeHash() {
			t.Fatalf("Intégrité cryptographique PAF corrompue au passage %s -> %s", from, to)
		}
		currentHash = evt.Hash
		auditTrail = append(auditTrail, evt)
	}

	invoiceID := "INV-FR-2026-9999"
	validUBLInvoice := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
    <cbc:CustomizationID>urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0</cbc:CustomizationID>
    <cbc:ID>INV-FR-2026-9999</cbc:ID>
    <cbc:IssueDate>2026-10-02</cbc:IssueDate>
    <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
</Invoice>`)

	h := sha256.Sum256(validUBLInvoice)
	payloadDigest := hex.EncodeToString(h[:])

	// --- Étape A : Ingestion ---
	currentState := status.StateIngested
	recordTransition(invoiceID, currentState, status.StateValidating, "api:ingest", "Réception de la facture UBL", payloadDigest)
	currentState = status.StateValidating

	// --- Étape B : Contrôle de conformité sémantique (Validator) ---
	valResult, err := validator.QuickValidateProfile(ctx, validUBLInvoice)
	if err != nil {
		t.Fatalf("Erreur d'exécution du validateur : %v", err)
	}
	if !valResult.IsValid {
		t.Fatalf("La facture de test valide a été rejetée par le validateur : %+v", valResult.Diagnostics)
	}

	recordTransition(invoiceID, currentState, status.StateDeposited, "engine:validator", "Validation EN 16931 conforme", payloadDigest)
	currentState = status.StateDeposited

	// --- Étape C : Enregistrement fiscal Clearance (PPF) ---
	recordTransition(invoiceID, currentState, status.StateClearancePending, "worker:clearance", "Soumission au PPF Chorus Pro", payloadDigest)
	currentState = status.StateClearancePending

	clearanceResp, err := ppf.SubmitInvoice(ctx, invoiceID, validUBLInvoice)
	if err != nil {
		t.Fatalf("Échec de clearance PPF : %v", err)
	}
	if clearanceResp.Status != "CLEARED" {
		t.Fatalf("Statut de clearance attendu CLEARED, obtenu %s", clearanceResp.Status)
	}

	recordTransition(invoiceID, currentState, status.StateCleared, "connector:ppf", "Clearance validée par le PPF ("+clearanceResp.ClearanceID+")", payloadDigest)
	currentState = status.StateCleared

	// --- Étape D : Mise en file d'attente Outbox AS4 ---
	recordTransition(invoiceID, currentState, status.StateTransportPending, "router:dispatcher", "Enfilement dans l'Outbox eDelivery", payloadDigest)
	currentState = status.StateTransportPending

	outboxMsg := &as4.OutboxMessage{
		ID:          "OUTBOX-" + invoiceID,
		InvoiceID:   invoiceID,
		ReceiverID:  "0009:123456789",
		AS4Endpoint: "mock://peppol-ap-receiver.test/as4",
		Payload:     validUBLInvoice,
		MaxAttempts: 3,
	}
	if err := as4Repo.Enqueue(ctx, outboxMsg); err != nil {
		t.Fatalf("Échec de mise en file Outbox : %v", err)
	}

	// --- Étape E : Exécution du Worker AS4 & Réception NRR ---
	processed, err := as4Worker.ProcessBatch(ctx)
	if err != nil {
		t.Fatalf("Échec de dépilage par le worker AS4 : %v", err)
	}
	if processed != 1 {
		t.Fatalf("Attendu 1 message AS4 émis, obtenu %d", processed)
	}

	recordTransition(invoiceID, currentState, status.StateAS4Sent, "worker:as4", "Message SOAP AS4 expédié avec succès", payloadDigest)
	currentState = status.StateAS4Sent

	recordTransition(invoiceID, currentState, status.StateDeliveredNRR, "receiver:as4", "Accusé technique NRR reçu et vérifié", payloadDigest)
	currentState = status.StateDeliveredNRR

	// --- Étape F : Contrôles finaux ---
	if currentState != status.StateDeliveredNRR {
		t.Errorf("Statut final attendu DELIVERED_NRR, obtenu %s", currentState)
	}
	if !currentState.IsTerminal() {
		t.Errorf("L'état final DELIVERED_NRR doit être terminal")
	}

	// Vérification de la chaîne d'audit
	if len(auditTrail) != 7 {
		t.Fatalf("Attendu 7 événements d'audit scellés, obtenu %d", len(auditTrail))
	}

	for i := 1; i < len(auditTrail); i++ {
		if auditTrail[i].PrevHash != auditTrail[i-1].Hash {
			t.Errorf("Rupture du chaînage PAF à l'index %d : attendu %s, obtenu %s", i, auditTrail[i-1].Hash, auditTrail[i].PrevHash)
		}
	}
}
