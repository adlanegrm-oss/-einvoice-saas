package as4

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
)

type WorkerConfig struct {
	BatchSize      int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

type OutboxWorker struct {
	repo   OutboxRepository
	client *AS4Client
	cfg    WorkerConfig
}

func NewOutboxWorker(repo OutboxRepository, client *AS4Client, cfg WorkerConfig) *OutboxWorker {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = 200 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 10 * time.Second
	}
	return &OutboxWorker{
		repo:   repo,
		client: client,
		cfg:    cfg,
	}
}

// CalculateBackoff calcule le délai de rejeu avec backoff exponentiel et jitter
func (w *OutboxWorker) CalculateBackoff(attempt int) time.Duration {
	backoff := float64(w.cfg.InitialBackoff) * math.Pow(2, float64(attempt))
	if backoff > float64(w.cfg.MaxBackoff) {
		backoff = float64(w.cfg.MaxBackoff)
	}
	// Jitter +/- 20%
	jitter := (rand.Float64()*0.4 - 0.2) * backoff
	result := time.Duration(backoff + jitter)
	if result < 0 {
		return w.cfg.InitialBackoff
	}
	return result
}

// ProcessBatch dépile et traite un lot de messages prêts à être émis en AS4
func (w *OutboxWorker) ProcessBatch(ctx context.Context) (int, error) {
	messages, err := w.repo.FetchPending(ctx, w.cfg.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("lecture outbox : %w", err)
	}

	processed := 0
	for _, msg := range messages {
		processed++
		endpoint := &dispatcher.TargetEndpoint{
			ReceiverID:  msg.ReceiverID,
			AS4Endpoint: msg.AS4Endpoint,
		}

		receipt, err := w.client.SendPayload(ctx, endpoint, msg.Payload)
		if err != nil {
			if msg.Attempts+1 >= msg.MaxAttempts {
				// Dépassement de quota -> transfert en Dead Letter Queue
				_ = w.repo.MoveToDLQ(ctx, msg.ID, fmt.Sprintf("%v (max retries: %d)", err, msg.MaxAttempts))
			} else {
				// Replanification avec backoff exponentiel
				nextRetry := time.Now().UTC().Add(w.CalculateBackoff(msg.Attempts))
				_ = w.repo.MarkRetry(ctx, msg.ID, err.Error(), nextRetry)
			}
			continue
		}

		// Succès de transport et NRR reçu
		genMessageID := fmt.Sprintf("msg_%s", msg.ID)
		_ = w.repo.MarkSent(ctx, msg.ID, genMessageID, receipt)
	}

	return processed, nil
}
