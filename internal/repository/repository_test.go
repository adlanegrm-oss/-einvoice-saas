package repository

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestRepository_Smoke(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("context.Background() ne doit pas Ãªtre nil")
	}
}

func TestInvoiceRecord_ZeroValue(t *testing.T) {
	var inv InvoiceRecord
	if inv.Status != "" {
		t.Errorf("Status zero value = %q; attendu vide", inv.Status)
	}
	if !inv.TotalTaxInclusive.IsZero() {
		t.Error("TotalTaxInclusive zero value devrait Ãªtre 0")
	}
}

func TestInvoiceRecord_Fields(t *testing.T) {
	now := time.Now().UTC()
	inv := InvoiceRecord{
		TenantID:           "tenant-1",
		ID:      "FACT-2026-001",
		SellerIdentifier:   "12345678900012",
		BuyerIdentifier:    "98765432100034",
		IssueDate:          now,
		Currency:           "EUR",
		TotalTaxInclusive:  decimal.NewFromFloat(1200.50),
		Syntax:             "UBL",
		Profile:            "EN16931",
		Status:             StatusReceived,
		DocumentSHA256:     "abc123",
		DocumentStorageKey: "s3://bucket/key",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if inv.TenantID != "tenant-1" {
		t.Errorf("TenantID = %q", inv.TenantID)
	}
	if inv.Status != StatusReceived {
		t.Errorf("Status = %q; attendu RECEIVED", inv.Status)
	}
	if inv.TotalTaxInclusive.String() != "1200.5" {
		t.Errorf("TotalTaxInclusive = %s; attendu 1200.5", inv.TotalTaxInclusive)
	}
}

func TestEventErrors(t *testing.T) {
	if ErrBrokenHashChain == nil {
		t.Error("ErrBrokenHashChain ne doit pas Ãªtre nil")
	}
	if ErrInvalidEventHash == nil {
		t.Error("ErrInvalidEventHash ne doit pas Ãªtre nil")
	}
}

func TestIdempotencyErrors(t *testing.T) {
	if ErrIdempotencyKeyExists == nil {
		t.Error("ErrIdempotencyKeyExists ne doit pas Ãªtre nil")
	}
}
