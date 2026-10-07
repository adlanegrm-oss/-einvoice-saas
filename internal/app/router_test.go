package app

import (
	"context"
	"testing"
	"time"

	"einvoice-saas/internal/repository"
	"github.com/shopspring/decimal"
)

func TestInMemInvoiceRepo(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemInvoiceRepo()

	inv := &repository.InvoiceRecord{
		TenantID:           "tenant-1",
		ID:                 "inv-001",
		InvoiceNumber:      "FACT-001",
		SellerIdentifier:   "12345678900012",
		BuyerIdentifier:    "98765432100034",
		IssueDate:          time.Now().UTC(),
		Currency:           "EUR",
		TotalTaxInclusive:  decimal.NewFromFloat(1200),
		Syntax:             "UBL",
		Profile:            "EN16931",
		Status:             repository.StatusReceived,
		DocumentSHA256:     "abc123sha",
		DocumentStorageKey: "key/1",
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	// Create
	if err := repo.Create(ctx, inv); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Create duplicate → ErrDuplicateBusiness
	if err := repo.Create(ctx, inv); err != repository.ErrDuplicateBusiness {
		t.Errorf("expected ErrDuplicateBusiness, got %v", err)
	}

	// GetByID
	got, err := repo.GetByID(ctx, "tenant-1", "inv-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.InvoiceNumber != "FACT-001" {
		t.Errorf("InvoiceNumber = %s; want FACT-001", got.InvoiceNumber)
	}

	// GetByID not found
	_, err = repo.GetByID(ctx, "tenant-1", "does-not-exist")
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// GetBySHA256
	got, err = repo.GetBySHA256(ctx, "tenant-1", "abc123sha")
	if err != nil {
		t.Fatalf("GetBySHA256 failed: %v", err)
	}
	if got.ID != "inv-001" {
		t.Errorf("ID = %s; want inv-001", got.ID)
	}

	// GetBySHA256 not found
	_, err = repo.GetBySHA256(ctx, "tenant-1", "unknown")
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// UpdateStatus
	if err := repo.UpdateStatus(ctx, "tenant-1", "inv-001", repository.StatusValidated); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
	got, _ = repo.GetByID(ctx, "tenant-1", "inv-001")
	if got.Status != repository.StatusValidated {
		t.Errorf("Status = %s; want VALIDATED", got.Status)
	}

	// UpdateStatus not found
	if err := repo.UpdateStatus(ctx, "tenant-1", "missing", repository.StatusValidated); err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInMemEventRepo(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemEventRepo()

	ev1 := &repository.InvoiceEventRecord{
		TenantID:     "tenant-1",
		InvoiceID:    "inv-001",
		Sequence:     1,
		EventID:      "evt-1",
		EventType:    "RECEIVED",
		Actor:        "system",
		CurrentHash:  "hash1",
		TimestampUTC: time.Now().UTC(),
	}
	ev2 := &repository.InvoiceEventRecord{
		TenantID:     "tenant-1",
		InvoiceID:    "inv-001",
		Sequence:     2,
		EventID:      "evt-2",
		EventType:    "VALIDATED",
		Actor:        "system",
		PreviousHash: "hash1",
		CurrentHash:  "hash2",
		TimestampUTC: time.Now().UTC(),
	}

	// Append
	if err := repo.Append(ctx, ev1); err != nil {
		t.Fatalf("Append 1 failed: %v", err)
	}
	if err := repo.Append(ctx, ev2); err != nil {
		t.Fatalf("Append 2 failed: %v", err)
	}

	// GetHistory
	history, err := repo.GetHistory(ctx, "tenant-1", "inv-001")
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("history length = %d; want 2", len(history))
	}

	// GetLatestSequence
	seq, hash, err := repo.GetLatestSequence(ctx, "tenant-1", "inv-001")
	if err != nil {
		t.Fatalf("GetLatestSequence failed: %v", err)
	}
	if seq != 2 || hash != "hash2" {
		t.Errorf("GetLatestSequence = (%d, %s); want (2, hash2)", seq, hash)
	}

	// Empty history
	seq, hash, err = repo.GetLatestSequence(ctx, "tenant-1", "unknown")
	if err != nil {
		t.Fatalf("GetLatestSequence empty failed: %v", err)
	}
	if seq != 0 || hash != "" {
		t.Errorf("expected (0, \"\"), got (%d, %s)", seq, hash)
	}
}

func TestInMemIdemRepo(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemIdemRepo()

	rec := &repository.IdempotencyRecord{
		TenantID:       "tenant-1",
		Key:            "idem-key-1",
		RequestHash:    "reqhash",
		InvoiceID:      "inv-001",
		ResponseStatus: 202,
		ResponseBody:   []byte(`{"status":"accepted"}`),
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour),
	}

	// Save
	if err := repo.Save(ctx, rec); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Get
	got, err := repo.Get(ctx, "tenant-1", "idem-key-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.InvoiceID != "inv-001" {
		t.Errorf("InvoiceID = %s; want inv-001", got.InvoiceID)
	}

	// Get not found
	_, err = repo.Get(ctx, "tenant-1", "unknown-key")
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
