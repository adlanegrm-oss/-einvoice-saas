package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/middleware"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/parser"
	"einvoice-saas/internal/repository"
	"einvoice-saas/internal/service"
	"einvoice-saas/internal/validator"
)

type UnifiedValidationResponse struct {
	Valid            bool                        `json:"valid"`
	Syntax           string                      `json:"syntax"`
	TargetProfile    string                      `json:"target_profile"`
	SchematronReport *validator.SchematronReport `json:"schematron_report,omitempty"`
	ArithmeticReport *validator.NormativeValidationResult `json:"arithmetic_report,omitempty"`
	FiscalReport     *model.ValidationReport     `json:"fiscal_report,omitempty"`
	CanonicalInvoice *model.CanonicalInvoice     `json:"canonical_invoice,omitempty"`
}

func ExecuteValidationPipeline(
	xmlData []byte,
	profile validator.ValidationProfile,
	schematronEngine *validator.SchematronEngine,
	normativeValidator *validator.NormativeValidator,
	frFiscalValidator *fr.FranceCanonicalValidator,
) (*UnifiedValidationResponse, error) {
	response := &UnifiedValidationResponse{
		Valid:         true,
		TargetProfile: string(profile),
	}

	schemReport, err := schematronEngine.ValidateProfile(xmlData, profile)
	if err != nil {
		response.Valid = false
		return response, fmt.Errorf("schematron_execution_failed: %w", err)
	}
	response.SchematronReport = schemReport
	if !schemReport.Valid {
		response.Valid = false
	}

	var canonical *model.CanonicalInvoice
	var normErr error

	if bytes.Contains(xmlData, []byte("CrossIndustryInvoice")) {
		response.Syntax = "CII-D16B"
		canonical, normErr = model.NormalizeCIIToCanonical(xmlData)
	} else {
		response.Syntax = "UBL-2.1"
		canonical, normErr = model.NormalizeUBLToCanonical(xmlData)
	}

	if normErr != nil {
		response.Valid = false
		return response, fmt.Errorf("normalization_failed: %w", normErr)
	}
	response.CanonicalInvoice = canonical

	arithResult, err := normativeValidator.ValidateCanonical(canonical)
	if err != nil {
		response.Valid = false
		return response, fmt.Errorf("arithmetic_validation_failed: %w", err)
	}
	response.ArithmeticReport = arithResult
	if !arithResult.Valid {
		response.Valid = false
	}

	if profile == validator.ProfileCIUSFR {
		fiscalReport := frFiscalValidator.Validate(canonical)
		response.FiscalReport = &fiscalReport
		if !fiscalReport.Valid {
			response.Valid = false
		}
	}

	return response, nil
}

type InMemInvoiceRepo struct {
	mu   sync.Mutex
	data map[string]*repository.InvoiceRecord
}

func NewInMemInvoiceRepo() *InMemInvoiceRepo {
	return &InMemInvoiceRepo{data: make(map[string]*repository.InvoiceRecord)}
}

func (r *InMemInvoiceRepo) Create(ctx context.Context, inv *repository.InvoiceRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := inv.TenantID + ":" + inv.ID
	if _, ok := r.data[key]; ok {
		return repository.ErrDuplicateBusiness
	}
	r.data[key] = inv
	return nil
}

func (r *InMemInvoiceRepo) GetByID(ctx context.Context, tenantID, id string) (*repository.InvoiceRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.data[tenantID+":"+id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return inv, nil
}

func (r *InMemInvoiceRepo) GetBySHA256(ctx context.Context, tenantID, sha256Hash string) (*repository.InvoiceRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.data {
		if inv.TenantID == tenantID && inv.DocumentSHA256 == sha256Hash {
			return inv, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (r *InMemInvoiceRepo) UpdateStatus(ctx context.Context, tenantID, id string, target repository.InvoiceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.data[tenantID+":"+id]
	if !ok {
		return repository.ErrNotFound
	}
	inv.Status = target
	return nil
}

type InMemEventRepo struct {
	mu   sync.Mutex
	data map[string][]repository.InvoiceEventRecord
}

func NewInMemEventRepo() *InMemEventRepo {
	return &InMemEventRepo{data: make(map[string][]repository.InvoiceEventRecord)}
}

func (r *InMemEventRepo) Append(ctx context.Context, ev *repository.InvoiceEventRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := ev.TenantID + ":" + ev.InvoiceID
	r.data[key] = append(r.data[key], *ev)
	return nil
}

func (r *InMemEventRepo) GetHistory(ctx context.Context, tenantID, invoiceID string) ([]repository.InvoiceEventRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data[tenantID+":"+invoiceID], nil
}

func (r *InMemEventRepo) GetLatestSequence(ctx context.Context, tenantID, invoiceID string) (int, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	events := r.data[tenantID+":"+invoiceID]
	if len(events) == 0 {
		return 0, "", nil
	}
	last := events[len(events)-1]
	return last.Sequence, last.CurrentHash, nil
}

type InMemIdemRepo struct {
	mu   sync.Mutex
	data map[string]*repository.IdempotencyRecord
}

func NewInMemIdemRepo() *InMemIdemRepo {
	return &InMemIdemRepo{data: make(map[string]*repository.IdempotencyRecord)}
}

func (r *InMemIdemRepo) Save(ctx context.Context, rec *repository.IdempotencyRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[rec.TenantID+":"+rec.Key] = rec
	return nil
}

func (r *InMemIdemRepo) Get(ctx context.Context, tenantID, key string) (*repository.IdempotencyRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.data[tenantID+":"+key]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return rec, nil
}

func SetupRouter(keyStore middleware.APIKeyStore, invoiceSvc *service.InvoiceService) http.Handler {
	mux := http.NewServeMux()

	schematronEngine := validator.NewSchematronEngine(nil)
	normativeValidator := validator.NewNormativeValidator(true)
	frFiscalValidator := fr.NewFranceCanonicalValidator()

	mux.HandleFunc("POST /v1/invoices/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		xmlData, err := parser.HardenedXMLReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "xml_security_rejection",
				"message": err.Error(),
			})
			return
		}

		profileParam := r.URL.Query().Get("profile")
		profile := validator.ProfileCIUSFR
		if strings.EqualFold(profileParam, string(validator.ProfileEN16931)) {
			profile = validator.ProfileEN16931
		}

		resp, err := ExecuteValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":             false,
				"error":             "pipeline_error",
				"message":           err.Error(),
				"schematron_report": resp.SchematronReport,
			})
			return
		}

		if !resp.Valid {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("POST /v1/invoices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tenantID, _ := middleware.GetTenantID(r.Context())

		xmlData, err := parser.HardenedXMLReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "xml_security_rejection",
				"message": err.Error(),
			})
			return
		}

		profileParam := r.URL.Query().Get("profile")
		profile := validator.ProfileCIUSFR
		if strings.EqualFold(profileParam, string(validator.ProfileEN16931)) {
			profile = validator.ProfileEN16931
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")

		cmd := service.IngestionCommand{
			TenantID:       tenantID,
			IdempotencyKey: idempotencyKey,
			Profile:        profile,
			RawXML:         xmlData,
			Actor:          "api_key_gateway",
		}

		result, err := invoiceSvc.IngestInvoice(r.Context(), cmd)
		if err != nil {
			if errors.Is(err, service.ErrDuplicateDocument) {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "duplicate_document",
					"message": "ce document a deja ete ingere pour ce tenant",
				})
				return
			}
			if errors.Is(err, service.ErrIdempotencyConflict) {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "idempotency_conflict",
					"message": "la cle d'idempotence a deja ete utilisee avec un document different",
				})
				return
			}

			h := sha256.Sum256(xmlData)
			docSHA256 := hex.EncodeToString(h[:])

			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":          "rejected",
				"error":           "invoice_pipeline_failed",
				"document_sha256": docSHA256,
				"message":         err.Error(),
			})
			return
		}

		if result.Status == "rejected" {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}

		_ = json.NewEncoder(w).Encode(result)
	})

	return middleware.RequireAPIKey(keyStore)(mux)
}
