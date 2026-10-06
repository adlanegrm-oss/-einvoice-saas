package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrIdempotencyKeyExists = errors.New("idempotency_key_already_processed")
)

type IdempotencyRecord struct {
	TenantID       string
	Key            string
	RequestHash    string
	InvoiceID      string
	ResponseStatus int
	ResponseBody   json.RawMessage
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

type IdempotencyRepository interface {
	Get(ctx context.Context, tenantID, key string) (*IdempotencyRecord, error)
	Save(ctx context.Context, record *IdempotencyRecord) error
}
