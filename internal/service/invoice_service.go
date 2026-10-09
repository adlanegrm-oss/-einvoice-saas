package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"einvoice-saas/internal/evidence"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/repository"
	"einvoice-saas/internal/validator"

	"github.com/shopspring/decimal"
)

const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

var (
	ErrDuplicateDocument   = errors.New("document_already_ingested")
	ErrIdempotencyConflict = errors.New("idempotency_key_payload_mismatch")
)

type ValidationPipelineFunc func(
	xmlData []byte,
	profile validator.ValidationProfile,
) (syntax string, canonical *model.CanonicalInvoice, valid bool, rep interface{}, err error)

type IngestionCommand struct {
	TenantID       string
	IdempotencyKey string
	Profile        validator.ValidationProfile
	RawXML         []byte
	Actor          string
}

type IngestionResult struct {
	IsIdempotentReplay bool        `json:"is_idempotent_replay"`
	Status             string      `json:"status"`
	InvoiceID          string      `json:"invoice_id"`
	TenantID           string      `json:"tenant_id"`
	DocumentSHA256     string      `json:"document_sha256"`
	AuditHash          string      `json:"audit_hash"`
	ValidationReport   interface{} `json:"validation,omitempty"`
	ErrorMessage       string      `json:"error,omitempty"`
}

type InvoiceService struct {
	invoiceRepo       repository.InvoiceRepository
	eventRepo         repository.EventRepository
	idempotencyRepo   repository.IdempotencyRepository
	transactionRunner repository.TransactionRunner
	validateFn        ValidationPipelineFunc
}

func NewInvoiceService(
	invRepo repository.InvoiceRepository,
	evtRepo repository.EventRepository,
	idemRepo repository.IdempotencyRepository,
	valFn ValidationPipelineFunc,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:     invRepo,
		eventRepo:       evtRepo,
		idempotencyRepo: idemRepo,
		validateFn:      valFn,
	}
}

func NewInvoiceServiceWithTransactionRunner(
	invRepo repository.InvoiceRepository,
	evtRepo repository.EventRepository,
	idemRepo repository.IdempotencyRepository,
	valFn ValidationPipelineFunc,
	runner repository.TransactionRunner,
) *InvoiceService {
	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, valFn)
	svc.transactionRunner = runner
	return svc
}
func (s *InvoiceService) IngestInvoice(ctx context.Context, cmd IngestionCommand) (*IngestionResult, error) {
	h := sha256.Sum256(cmd.RawXML)
	docSHA256 := hex.EncodeToString(h[:])

	lockKey := "doc:" + cmd.TenantID + ":" + docSHA256
	if cmd.IdempotencyKey != "" {
		lockKey = "idem:" + cmd.TenantID + ":" + cmd.IdempotencyKey
	}

	var result *IngestionResult

	ingest := func(txCtx context.Context) error {
		if cmd.IdempotencyKey != "" {
			existing, err := s.idempotencyRepo.Get(txCtx, cmd.TenantID, cmd.IdempotencyKey)
			if err == nil {
				if existing == nil {
					return fmt.Errorf("idempotency_lookup_returned_nil_record")
				}
				if existing.RequestHash != docSHA256 {
					return ErrIdempotencyConflict
				}

				var cached IngestionResult
				if err := json.Unmarshal(existing.ResponseBody, &cached); err != nil {
					return fmt.Errorf("invalid_cached_idempotency_response: %w", err)
				}
				cached.IsIdempotentReplay = true
				result = &cached
				return nil
			}
			if !errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("failed_to_check_idempotency_key: %w", err)
			}
		}

		existingDoc, err := s.invoiceRepo.GetBySHA256(txCtx, cmd.TenantID, docSHA256)
		if err == nil {
			if existingDoc != nil {
				return ErrDuplicateDocument
			}
			return fmt.Errorf("document_lookup_returned_nil_record")
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("failed_to_check_duplicate_document: %w", err)
		}

		syntax, canonical, valid, valReport, valErr := s.validateFn(cmd.RawXML, cmd.Profile)
		if valErr != nil {
			return fmt.Errorf("pipeline_failure: %w", valErr)
		}

		now := time.Now().UTC()
		invoiceID := "INV-" + now.Format("20060102150405.000")
		var sellerID, buyerID, invNum, currency string
		issueDate := now
		totalTaxInc := decimal.Zero

		if canonical != nil {
			if canonical.ID != "" {
				invoiceID = canonical.ID
			}
			invNum = canonical.ID
			sellerID = canonical.Seller.LegalID
			buyerID = canonical.Buyer.LegalID
			currency = canonical.DocumentCurrency
			if parsed, parseErr := time.Parse("2006-01-02", canonical.IssueDate); parseErr == nil {
				issueDate = parsed
			}
			totalTaxInc = canonical.Totals.TaxInclusiveAmount
		}

		invRecord := &repository.InvoiceRecord{
			TenantID:          cmd.TenantID,
			ID:                invoiceID,
			InvoiceNumber:     invNum,
			SellerIdentifier:  sellerID,
			BuyerIdentifier:   buyerID,
			IssueDate:         issueDate,
			Currency:          currency,
			TotalTaxInclusive: totalTaxInc,
			Syntax:            syntax,
			Profile:           string(cmd.Profile),
			Status:            repository.StatusReceived,
			DocumentSHA256:    docSHA256,
		}
		if err := s.invoiceRepo.Create(txCtx, invRecord); err != nil {
			if errors.Is(err, repository.ErrDuplicatePayload) {
				return ErrDuplicateDocument
			}
			return fmt.Errorf("failed_to_persist_invoice: %w", err)
		}

		if err := s.invoiceRepo.UpdateStatus(txCtx, cmd.TenantID, invoiceID, repository.StatusValidating); err != nil {
			return fmt.Errorf("failed_to_set_validating_status: %w", err)
		}
		targetStatus := repository.StatusValidated
		responseStatus := "accepted"
		httpStatusCode := 202
		if !valid {
			targetStatus = repository.StatusRejected
			responseStatus = "rejected"
			httpStatusCode = 422
		}
		if err := s.invoiceRepo.UpdateStatus(txCtx, cmd.TenantID, invoiceID, targetStatus); err != nil {
			return fmt.Errorf("failed_to_update_invoice_status: %w", err)
		}

		auditPayload := evidence.AuditEvent{
			TenantID:       cmd.TenantID,
			InvoiceID:      invoiceID,
			EventType:      string(targetStatus),
			Actor:          cmd.Actor,
			TimestampUTC:   now,
			DocumentSHA256: docSHA256,
			PayloadSummary: fmt.Sprintf("Syntax: %s, Profile: %s, Valid: %v", syntax, cmd.Profile, valid),
			PreviousHash:   genesisHash,
		}
		currentHash := evidence.CalculateChainHash(&auditPayload)

		eventRecord := &repository.InvoiceEventRecord{
			TenantID:       cmd.TenantID,
			InvoiceID:      invoiceID,
			Sequence:       1,
			EventID:        "ev_" + now.Format("150405.000"),
			EventType:      string(targetStatus),
			Actor:          cmd.Actor,
			DocumentSHA256: docSHA256,
			PayloadSummary: auditPayload.PayloadSummary,
			PreviousHash:   genesisHash,
			CurrentHash:    currentHash,
			TimestampUTC:   now,
		}
		if err := s.eventRepo.Append(txCtx, eventRecord); err != nil {
			return fmt.Errorf("audit_event_failed: %w", err)
		}

		result = &IngestionResult{
			Status:           responseStatus,
			InvoiceID:        invoiceID,
			TenantID:         cmd.TenantID,
			DocumentSHA256:   docSHA256,
			AuditHash:        currentHash,
			ValidationReport: valReport,
		}

		if cmd.IdempotencyKey != "" {
			respBytes, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed_to_marshal_idempotency_payload: %w", err)
			}
			if err := s.idempotencyRepo.Save(txCtx, &repository.IdempotencyRecord{
				TenantID:       cmd.TenantID,
				Key:            cmd.IdempotencyKey,
				RequestHash:    docSHA256,
				InvoiceID:      invoiceID,
				ResponseStatus: httpStatusCode,
				ResponseBody:   respBytes,
				ExpiresAt:      time.Now().Add(24 * time.Hour),
			}); err != nil {
				return fmt.Errorf("failed_to_save_idempotency_key: %w", err)
			}
		}
		return nil
	}

	var err error
	if s.transactionRunner != nil {
		err = s.transactionRunner.WithinTransaction(ctx, lockKey, ingest)
	} else {
		err = ingest(ctx)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
