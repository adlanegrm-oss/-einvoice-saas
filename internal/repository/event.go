package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBrokenHashChain  = errors.New("broken_event_hash_chain")
	ErrInvalidEventHash = errors.New("invalid_event_hash_mismatch")
)

type InvoiceEventRecord struct {
	TenantID       string
	InvoiceID      string
	Sequence       int
	EventID        string
	EventType      string
	Actor          string
	DocumentSHA256 string
	PayloadSummary string
	PreviousHash   string
	CurrentHash    string
	TimestampUTC   time.Time
	CreatedAt      time.Time
}

type EventRepository interface {
	Append(ctx context.Context, event *InvoiceEventRecord) error
	GetHistory(ctx context.Context, tenantID, invoiceID string) ([]InvoiceEventRecord, error)
	GetLatestSequence(ctx context.Context, tenantID, invoiceID string) (int, string, error)
}
