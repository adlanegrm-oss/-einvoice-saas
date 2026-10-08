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
	invoiceRepo     repository.InvoiceRepository
	eventRepo       repository.EventRepository
	idempotencyRepo repository.IdempotencyRepository
	validateFn      ValidationPipelineFunc
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

func (s *InvoiceService) IngestInvoice(ctx context.Context, cmd IngestionCommand) (*IngestionResult, error) {
	// 1. Calcul du digest SHA-256 du document brut
	h := sha256.Sum256(cmd.RawXML)
	docSHA256 := hex.EncodeToString(h[:])

	// 2. Traitement d'idempotence technique et dÃ©tection de conflit de payload
	if cmd.IdempotencyKey != "" {
		existing, err := s.idempotencyRepo.Get(ctx, cmd.TenantID, cmd.IdempotencyKey)
		if err == nil && existing != nil {
			if existing.RequestHash != docSHA256 {
				return nil, ErrIdempotencyConflict
			}
			var cachedResult IngestionResult
			if unmarshalErr := json.Unmarshal(existing.ResponseBody, &cachedResult); unmarshalErr == nil {
				cachedResult.IsIdempotentReplay = true
				return &cachedResult, nil
			}
		}
	}

	// 3. DÃ©tection de doublon documentaire par SHA-256
	existingDoc, err := s.invoiceRepo.GetBySHA256(ctx, cmd.TenantID, docSHA256)
	if err == nil && existingDoc != nil {
		return nil, ErrDuplicateDocument
	}

	// 4. ExÃ©cution du moteur de validation (Schematron -> Pivot -> Normatif -> Fiscal)
	syntax, canonical, valid, valReport, valErr := s.validateFn(cmd.RawXML, cmd.Profile)
	if valErr != nil {
		return nil, fmt.Errorf("pipeline_failure: %w", valErr)
	}

	invoiceID := "INV-" + time.Now().UTC().Format("20060102150405.000")
	var sellerID, buyerID, invNum, currency string
	var issueDate time.Time
	var totalTaxInc = decimal.Zero

	if canonical != nil {
		if canonical.ID != "" {
			invoiceID = canonical.ID
		}
		invNum = canonical.ID
		sellerID = canonical.Seller.LegalID
		buyerID = canonical.Buyer.LegalID
		currency = canonical.DocumentCurrency
		if parsed, err := time.Parse("2006-01-02", canonical.IssueDate); err == nil {
issueDate = parsed
}
		totalTaxInc = canonical.Totals.TaxInclusiveAmount
	}
	if issueDate.IsZero() {
		issueDate = time.Now().UTC()
	}

	// 5. Persistance initiale en Ã©tat RECEIVED
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

	if err := s.invoiceRepo.Create(ctx, invRecord); err != nil {
		return nil, fmt.Errorf("failed_to_persist_invoice: %w", err)
	}

	// 6. Transition vers VALIDATING (contrÃ´le strict de la machine Ã  Ã©tats)
	if err := s.invoiceRepo.UpdateStatus(ctx, cmd.TenantID, invoiceID, repository.StatusValidating); err != nil {
		return nil, fmt.Errorf("failed_to_set_validating_status: %w", err)
	}

	targetStatus := repository.StatusValidated
	responseStatus := "accepted"
	httpStatusCode := 202
	if !valid {
		targetStatus = repository.StatusRejected
		responseStatus = "rejected"
		httpStatusCode = 422
	}

	if err := s.invoiceRepo.UpdateStatus(ctx, cmd.TenantID, invoiceID, targetStatus); err != nil {
		return nil, fmt.Errorf("failed_to_update_status: %w", err)
	}

	// 7. Scellement de la chaÃ®ne d'audit
	nowUTC := time.Now().UTC()
	auditPayload := evidence.AuditEvent{
		TenantID:       cmd.TenantID,
		InvoiceID:      invoiceID,
		EventType:      string(targetStatus),
		Actor:          cmd.Actor,
		TimestampUTC:   nowUTC,
		DocumentSHA256: docSHA256,
		PayloadSummary: fmt.Sprintf("Syntax: %s, Profile: %s, Valid: %v", syntax, cmd.Profile, valid),
		PreviousHash:   genesisHash,
	}
	currentHash := evidence.CalculateChainHash(&auditPayload)

	eventRecord := &repository.InvoiceEventRecord{
		TenantID:       cmd.TenantID,
		InvoiceID:      invoiceID,
		Sequence:       1,
		EventID:        "ev_" + nowUTC.Format("150405.000"),
		EventType:      string(targetStatus),
		Actor:          cmd.Actor,
		DocumentSHA256: docSHA256,
		PayloadSummary: auditPayload.PayloadSummary,
		PreviousHash:   genesisHash,
		CurrentHash:    currentHash,
		TimestampUTC:   nowUTC,
	}

	if err := s.eventRepo.Append(ctx, eventRecord); err != nil {
		return nil, fmt.Errorf("audit_event_failed: %w", err)
	}

	result := &IngestionResult{
		Status:           responseStatus,
		InvoiceID:        invoiceID,
		TenantID:         cmd.TenantID,
		DocumentSHA256:   docSHA256,
		AuditHash:        currentHash,
		ValidationReport: valReport,
	}

	// 8. Enregistrement transactionnel du rÃ©sultat d'idempotence (erreur non masquÃ©e)
	if cmd.IdempotencyKey != "" {
		respBytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed_to_marshal_idempotency_payload: %w", err)
		}
		if err := s.idempotencyRepo.Save(ctx, &repository.IdempotencyRecord{
			TenantID:       cmd.TenantID,
			Key:            cmd.IdempotencyKey,
			RequestHash:    docSHA256,
			InvoiceID:      invoiceID,
			ResponseStatus: httpStatusCode,
			ResponseBody:   respBytes,
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}); err != nil {
			return nil, fmt.Errorf("failed_to_save_idempotency_key: %w", err)
		}
	}

	return result, nil
}
