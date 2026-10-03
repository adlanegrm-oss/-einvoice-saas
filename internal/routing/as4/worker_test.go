package as4_test

import (
	"context"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/as4"
)

func TestOutboxWorker_SuccessFlow(t *testing.T) {
	repo := as4.NewInMemoryOutboxRepository()
	client := as4.NewAS4Client()
	worker := as4.NewOutboxWorker(repo, client, as4.WorkerConfig{BatchSize: 5})

	ctx := context.Background()
	msg := &as4.OutboxMessage{
		ID:          "MSG-01",
		InvoiceID:   "INV-TEST-001",
		ReceiverID:  "0009:123456789",
		AS4Endpoint: "mock://peppol-ap.test/as4",
		Payload:     []byte("<Invoice>Test AS4 Outbox</Invoice>"),
		MaxAttempts: 3,
	}

	if err := repo.Enqueue(ctx, msg); err != nil {
		t.Fatalf("échec Enqueue: %v", err)
	}

	count, err := worker.ProcessBatch(ctx)
	if err != nil {
		t.Fatalf("échec ProcessBatch: %v", err)
	}
	if count != 1 {
		t.Fatalf("attendu 1 message traité, obtenu %d", count)
	}

	pending, _ := repo.FetchPending(ctx, 10)
	if len(pending) != 0 {
		t.Errorf("le message devrait être marqué comme envoyé")
	}
}

func TestOutboxWorker_RetryAndDLQ(t *testing.T) {
	repo := as4.NewInMemoryOutboxRepository()
	client := as4.NewAS4Client()
	worker := as4.NewOutboxWorker(repo, client, as4.WorkerConfig{
		BatchSize:      5,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     50 * time.Millisecond,
	})

	ctx := context.Background()
	// Endpoint invalide pour forcer une erreur réseau immédiate
	msg := &as4.OutboxMessage{
		ID:          "MSG-FAIL-01",
		InvoiceID:   "INV-FAIL-001",
		ReceiverID:  "0009:123456789",
		AS4Endpoint: "http://127.0.0.1:1/as4-dead-endpoint",
		Payload:     []byte("<Invoice>Test Fail</Invoice>"),
		MaxAttempts: 2,
	}

	_ = repo.Enqueue(ctx, msg)

	// Tentative 1 -> Doit passer en RETRY (Status FAILED, Attempts=1)
	_, _ = worker.ProcessBatch(ctx)

	// Avance rapide pour expirer le backoff
	time.Sleep(70 * time.Millisecond)

	// Tentative 2 -> Doit atteindre MaxAttempts et basculer en DLQ
	_, _ = worker.ProcessBatch(ctx)

	pending, _ := repo.FetchPending(ctx, 10)
	if len(pending) != 0 {
		t.Errorf("attendu 0 pending (message basculé en DLQ), obtenu %d", len(pending))
	}
}

func TestOutboxWorker_BackoffCalculation(t *testing.T) {
	repo := as4.NewInMemoryOutboxRepository()
	client := as4.NewAS4Client()
	worker := as4.NewOutboxWorker(repo, client, as4.WorkerConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
	})

	d1 := worker.CalculateBackoff(1)
	d2 := worker.CalculateBackoff(3)

	if d2 <= d1 {
		t.Errorf("le délai de rejeu doit augmenter avec les tentatives : d1=%v, d2=%v", d1, d2)
	}
}
