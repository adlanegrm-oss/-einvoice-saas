package as4

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrMaxRetriesExceeded = errors.New("nombre maximum de tentatives atteint : routage vers DLQ")
)

type OutboxStatus string

const (
	OutboxPending    OutboxStatus = "PENDING"
	OutboxProcessing OutboxStatus = "PROCESSING"
	OutboxSent       OutboxStatus = "SENT"
	OutboxFailed     OutboxStatus = "FAILED"
	OutboxDLQ        OutboxStatus = "DLQ"
)

// OutboxMessage représente un message AS4 en file d'attente d'expédition
type OutboxMessage struct {
	ID             string       `json:"id"`
	InvoiceID      string       `json:"invoice_id"`
	ReceiverID     string       `json:"receiver_id"`
	AS4Endpoint    string       `json:"as4_endpoint"`
	Payload        []byte       `json:"payload"`
	MessageID      string       `json:"message_id,omitempty"`
	RefToMessageID string       `json:"ref_to_message_id,omitempty"`
	ReceiptNRR     string       `json:"receipt_nrr,omitempty"`
	Status         OutboxStatus `json:"status"`
	Attempts       int          `json:"attempts"`
	MaxAttempts    int          `json:"max_attempts"`
	LastError      string       `json:"last_error,omitempty"`
	NextRetryAt    time.Time    `json:"next_retry_at"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// OutboxRepository définit l'abstraction de persistance pour l'Outbox / DLQ
type OutboxRepository interface {
	Enqueue(ctx context.Context, msg *OutboxMessage) error
	FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkSent(ctx context.Context, id, messageID, receiptNRR string) error
	MarkRetry(ctx context.Context, id string, lastErr string, nextRetry time.Time) error
	MoveToDLQ(ctx context.Context, id string, reason string) error
}

// InMemoryOutboxRepository fournit un backend mémoire thread-safe pour tests et dev
type InMemoryOutboxRepository struct {
	mu       sync.RWMutex
	messages map[string]*OutboxMessage
}

func NewInMemoryOutboxRepository() *InMemoryOutboxRepository {
	return &InMemoryOutboxRepository{
		messages: make(map[string]*OutboxMessage),
	}
}

func (r *InMemoryOutboxRepository) Enqueue(ctx context.Context, msg *OutboxMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	msg.Status = OutboxPending
	msg.CreatedAt = time.Now().UTC()
	msg.UpdatedAt = msg.CreatedAt
	if msg.MaxAttempts == 0 {
		msg.MaxAttempts = 3
	}
	r.messages[msg.ID] = msg
	return nil
}

func (r *InMemoryOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := time.Now().UTC()
	var result []*OutboxMessage
	for _, m := range r.messages {
		if (m.Status == OutboxPending || m.Status == OutboxFailed) && now.After(m.NextRetryAt) {
			result = append(result, m)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *InMemoryOutboxRepository) MarkSent(ctx context.Context, id, messageID, receiptNRR string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, exists := r.messages[id]
	if !exists {
		return fmt.Errorf("message %s non trouvé", id)
	}
	m.Status = OutboxSent
	m.MessageID = messageID
	m.ReceiptNRR = receiptNRR
	m.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *InMemoryOutboxRepository) MarkRetry(ctx context.Context, id string, lastErr string, nextRetry time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, exists := r.messages[id]
	if !exists {
		return fmt.Errorf("message %s non trouvé", id)
	}
	m.Attempts++
	m.Status = OutboxFailed
	m.LastError = lastErr
	m.NextRetryAt = nextRetry
	m.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *InMemoryOutboxRepository) MoveToDLQ(ctx context.Context, id string, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, exists := r.messages[id]
	if !exists {
		return fmt.Errorf("message %s non trouvé", id)
	}
	m.Attempts++
	m.Status = OutboxDLQ
	m.LastError = reason
	m.UpdatedAt = time.Now().UTC()
	return nil
}