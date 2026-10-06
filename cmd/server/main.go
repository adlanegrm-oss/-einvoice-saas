package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/middleware"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/parser"
	"einvoice-saas/internal/repository"
	"einvoice-saas/internal/repository/postgres"
	"einvoice-saas/internal/service"
	"einvoice-saas/internal/validator"
)

type InMemoryKeyStore struct{}

func (s *InMemoryKeyStore) FindTenantByKeyHash(ctx context.Context, hash string) (*middleware.TenantRecord, error) {
	return &middleware.TenantRecord{
		ID:     "tenant_demo_erp",
		Active: true,
	}, nil
}

type UnifiedValidationResponse struct {
	Valid            bool                        `json:"valid"`
	Syntax           string                      `json:"syntax"`
	TargetProfile    string                      `json:"target_profile"`
	SchematronReport *validator.SchematronReport `json:"schematron_report,omitempty"`
	ArithmeticReport *validator.ValidationResult `json:"arithmetic_report,omitempty"`
	FiscalReport     *model.ValidationReport     `json:"fiscal_report,omitempty"`
	CanonicalInvoice *model.CanonicalInvoice     `json:"canonical_invoice,omitempty"`
}

func executeValidationPipeline(
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

type inMemInvoiceRepo struct {
	mu   sync.Mutex
	data map[string]*repository.InvoiceRecord
}

func newInMemInvoiceRepo() *inMemInvoiceRepo {
	return &inMemInvoiceRepo{data: make(map[string]*repository.InvoiceRecord)}
}

func (r *inMemInvoiceRepo) Create(ctx context.Context, inv *repository.InvoiceRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := inv.TenantID + ":" + inv.ID
	if _, ok := r.data[key]; ok {
		return repository.ErrDuplicateBusiness
	}
	r.data[key] = inv
	return nil
}

func (r *inMemInvoiceRepo) GetByID(ctx context.Context, tenantID, id string) (*repository.InvoiceRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.data[tenantID+":"+id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return inv, nil
}

func (r *inMemInvoiceRepo) GetBySHA256(ctx context.Context, tenantID, sha256Hash string) (*repository.InvoiceRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.data {
		if inv.TenantID == tenantID && inv.DocumentSHA256 == sha256Hash {
			return inv, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (r *inMemInvoiceRepo) UpdateStatus(ctx context.Context, tenantID, id string, target repository.InvoiceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.data[tenantID+":"+id]
	if !ok {
		return repository.ErrNotFound
	}
	inv.Status = target
	return nil
}

type inMemEventRepo struct {
	mu   sync.Mutex
	data map[string][]repository.InvoiceEventRecord
}

func newInMemEventRepo() *inMemEventRepo {
	return &inMemEventRepo{data: make(map[string][]repository.InvoiceEventRecord)}
}

func (r *inMemEventRepo) Append(ctx context.Context, ev *repository.InvoiceEventRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := ev.TenantID + ":" + ev.InvoiceID
	r.data[key] = append(r.data[key], *ev)
	return nil
}

func (r *inMemEventRepo) GetHistory(ctx context.Context, tenantID, invoiceID string) ([]repository.InvoiceEventRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data[tenantID+":"+invoiceID], nil
}

func (r *inMemEventRepo) GetLatestSequence(ctx context.Context, tenantID, invoiceID string) (int, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	events := r.data[tenantID+":"+invoiceID]
	if len(events) == 0 {
		return 0, "", nil
	}
	last := events[len(events)-1]
	return last.Sequence, last.CurrentHash, nil
}

type inMemIdemRepo struct {
	mu   sync.Mutex
	data map[string]*repository.IdempotencyRecord
}

func newInMemIdemRepo() *inMemIdemRepo {
	return &inMemIdemRepo{data: make(map[string]*repository.IdempotencyRecord)}
}

func (r *inMemIdemRepo) Save(ctx context.Context, rec *repository.IdempotencyRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[rec.TenantID+":"+rec.Key] = rec
	return nil
}

func (r *inMemIdemRepo) Get(ctx context.Context, tenantID, key string) (*repository.IdempotencyRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.data[tenantID+":"+key]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return rec, nil
}

func setupRouter(keyStore middleware.APIKeyStore, invoiceSvc *service.InvoiceService) http.Handler {
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

		resp, err := executeValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	schematronEngine := validator.NewSchematronEngine(nil)
	normativeValidator := validator.NewNormativeValidator(true)
	frFiscalValidator := fr.NewFranceCanonicalValidator()

	valFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		resp, err := executeValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
		if err != nil {
			return "", nil, false, resp, err
		}
		return resp.Syntax, resp.CanonicalInvoice, resp.Valid, resp, nil
	}

	var invRepo repository.InvoiceRepository
	var evtRepo repository.EventRepository
	var idemRepo repository.IdempotencyRepository

	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("ENVIRONMENT"))
	}
	isProduction := env == "production" || env == "prod"

	dbURL := os.Getenv("DATABASE_URL")
	if isProduction && dbURL == "" {
		log.Fatalf("[FATAL] Demarrage impossible en production : DATABASE_URL est obligatoire")
	}

	if dbURL != "" {
		db, err := postgres.OpenDB(dbURL)
		if err != nil {
			log.Fatalf("PostgreSQL connection failed: %v", err)
		}

		}

		invRepo = postgres.NewInvoiceRepo(db)
		evtRepo = postgres.NewEventRepo(db)
		idemRepo = postgres.NewIdempotencyRepo(db)
		log.Printf("[PERSISTENCE] PostgreSQL connecte avec succes (env=%s)", env)
	} else {
		invRepo = newInMemInvoiceRepo()
		evtRepo = newInMemEventRepo()
		idemRepo = newInMemIdemRepo()
		log.Printf("[PERSISTENCE] Mode In-Memory actif (dev/test uniquement)")
	}

	invoiceSvc := service.NewInvoiceService(invRepo, evtRepo, idemRepo, valFn)

	keyStore := &InMemoryKeyStore{}
	handler := setupRouter(keyStore, invoiceSvc)

	log.Printf("[READY] E-Invoicing Gateway demarree sur le port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
