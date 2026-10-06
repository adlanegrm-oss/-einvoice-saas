package repository

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrNotFound          = errors.New("entity_not_found")
	ErrDuplicateBusiness = errors.New("invoice_duplicate_business_key")
	ErrDuplicatePayload  = errors.New("invoice_duplicate_sha256")
	ErrInvalidTransition = errors.New("invalid_status_transition")
)

type InvoiceStatus string

const (
	StatusReceived   InvoiceStatus = "RECEIVED"
	StatusValidating InvoiceStatus = "VALIDATING"
	StatusValidated  InvoiceStatus = "VALIDATED"
	StatusRejected   InvoiceStatus = "REJECTED"
	StatusSubmitted  InvoiceStatus = "SUBMITTED"
	StatusCleared    InvoiceStatus = "CLEARED"
	StatusDelivered  InvoiceStatus = "DELIVERED"
)

// AllowedTransitions définit les règles formelles du cycle de vie
var allowedTransitions = map[InvoiceStatus]map[InvoiceStatus]bool{
	StatusReceived: {
		StatusValidating: true,
	},
	StatusValidating: {
		StatusValidated: true,
		StatusRejected:  true,
	},
	StatusValidated: {
		StatusSubmitted: true,
	},
	StatusSubmitted: {
		StatusCleared:  true,
		StatusRejected: true,
	},
	StatusCleared: {
		StatusDelivered: true,
	},
	StatusRejected:  {}, // État terminal
	StatusDelivered: {}, // État terminal
}

func ValidateTransition(from, to InvoiceStatus) bool {
	if nextStates, ok := allowedTransitions[from]; ok {
		return nextStates[to]
	}
	return false
}

type InvoiceRecord struct {
	TenantID           string
	ID                 string
	InvoiceNumber      string
	SellerIdentifier   string
	BuyerIdentifier    string
	IssueDate          time.Time
	Currency           string
	TotalTaxInclusive  decimal.Decimal
	Syntax             string
	Profile            string
	Status             InvoiceStatus
	DocumentSHA256     string
	DocumentStorageKey string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type InvoiceRepository interface {
	Create(ctx context.Context, inv *InvoiceRecord) error
	GetByID(ctx context.Context, tenantID, id string) (*InvoiceRecord, error)
	GetBySHA256(ctx context.Context, tenantID, sha256Hash string) (*InvoiceRecord, error)
	UpdateStatus(ctx context.Context, tenantID, id string, targetStatus InvoiceStatus) error
}
