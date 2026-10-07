package evidence

import (
	"testing"
	"time"
)

func TestVerifyChain_Success(t *testing.T) {
	tNow := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	genesisPrev := "0000000000000000000000000000000000000000000000000000000000000000"

	ev1 := AuditEvent{
		EventID:        "ev_01",
		TenantID:       "tenant_alpha",
		InvoiceID:      "inv_1001",
		EventType:      "INVOICE_INGESTED",
		Actor:          "api_key_system",
		TimestampUTC:   tNow,
		DocumentSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		PayloadSummary: "raw payload received",
		PreviousHash:   genesisPrev,
	}
	ev1.CurrentHash = CalculateChainHash(&ev1)

	ev2 := AuditEvent{
		EventID:        "ev_02",
		TenantID:       "tenant_alpha",
		InvoiceID:      "inv_1001",
		EventType:      "VALIDATION_PASSED",
		Actor:          "rules_engine",
		TimestampUTC:   tNow.Add(150 * time.Millisecond),
		DocumentSHA256: ev1.DocumentSHA256,
		PayloadSummary: "en16931 rules verified without errors",
		PreviousHash:   ev1.CurrentHash,
	}
	ev2.CurrentHash = CalculateChainHash(&ev2)

	chain := []AuditEvent{ev1, ev2}
	valid, err := VerifyChain(chain)
	if err != nil || !valid {
		t.Fatalf("expected valid chain, got err: %v", err)
	}
}

func TestVerifyChain_TamperDetection(t *testing.T) {
	tNow := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	genesisPrev := "0000000000000000000000000000000000000000000000000000000000000000"

	ev1 := AuditEvent{
		EventID:        "ev_01",
		TenantID:       "tenant_alpha",
		InvoiceID:      "inv_1001",
		EventType:      "INVOICE_INGESTED",
		Actor:          "api_key_system",
		TimestampUTC:   tNow,
		DocumentSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		PreviousHash:   genesisPrev,
	}
	ev1.CurrentHash = CalculateChainHash(&ev1)

	ev2 := AuditEvent{
		EventID:        "ev_02",
		TenantID:       "tenant_alpha",
		InvoiceID:      "inv_1001",
		EventType:      "VALIDATION_PASSED",
		Actor:          "rules_engine",
		TimestampUTC:   tNow.Add(150 * time.Millisecond),
		DocumentSHA256: ev1.DocumentSHA256,
		PreviousHash:   ev1.CurrentHash,
	}
	ev2.CurrentHash = CalculateChainHash(&ev2)

	// Altération malveillante après coup de la charge utile de ev1
	ev1.PayloadSummary = "ALTERED_BY_ATTACKER"

	chain := []AuditEvent{ev1, ev2}
	valid, err := VerifyChain(chain)
	if valid || err == nil {
		t.Fatal("expected chain verification to catch tampered event data")
	}
}
