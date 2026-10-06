package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/evidence"
	"einvoice-saas/internal/middleware"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/parser"
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

type IngestionResponse struct {
	Status         string                     `json:"status"`
	InvoiceID      string                     `json:"invoice_id"`
	TenantID       string                     `json:"tenant_id"`
	DocumentSHA256 string                     `json:"document_sha256"`
	AuditHash      string                     `json:"audit_hash"`
	Validation     *UnifiedValidationResponse `json:"validation,omitempty"`
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

	// 1. Validation Schematron
	schemReport, err := schematronEngine.ValidateProfile(xmlData, profile)
	if err != nil {
		response.Valid = false
		return response, fmt.Errorf("schematron_execution_failed: %w", err)
	}
	response.SchematronReport = schemReport
	if !schemReport.Valid {
		response.Valid = false
	}

	// 2. Détection syntaxe et normalisation
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

	// 3. Validation normative arithmétique
	arithResult, err := normativeValidator.ValidateCanonical(canonical)
	if err != nil {
		response.Valid = false
		return response, fmt.Errorf("arithmetic_validation_failed: %w", err)
	}
	response.ArithmeticReport = arithResult
	if !arithResult.Valid {
		response.Valid = false
	}

	// 4. Validation fiscale juridique nationale
	// Les règles fiscales françaises sont spécifiques à CIUS-FR.
	// Le profil EN16931 reste juridiction-neutre.
	if profile == validator.ProfileCIUSFR {
		fiscalReport := frFiscalValidator.Validate(canonical)
		response.FiscalReport = &fiscalReport
		if !fiscalReport.Valid {
			response.Valid = false
		}
	}

	return response, nil
}

func setupRouter(keyStore middleware.APIKeyStore) http.Handler {
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

		hashBytes := sha256.Sum256(xmlData)
		docSHA256 := hex.EncodeToString(hashBytes[:])

		profileParam := r.URL.Query().Get("profile")
		profile := validator.ProfileCIUSFR
		if strings.EqualFold(profileParam, string(validator.ProfileEN16931)) {
			profile = validator.ProfileEN16931
		}

		validationResp, err := executeValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":          "rejected",
				"error":           "invoice_pipeline_failed",
				"document_sha256": docSHA256,
				"validation":      validationResp,
				"message":         err.Error(),
			})
			return
		}
		if !validationResp.Valid {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":          "rejected",
				"error":           "invoice_compliance_failed",
				"document_sha256": docSHA256,
				"validation":      validationResp,
			})
			return
		}

		invoiceNumber := validationResp.CanonicalInvoice.InvoiceNumber
		if invoiceNumber == "" {
			invoiceNumber = "INV-" + time.Now().Format("20060102150405")
		}

		ev := evidence.AuditEvent{
			EventID:        "ev_ingest_" + time.Now().Format("150405.000"),
			TenantID:       tenantID,
			InvoiceID:      invoiceNumber,
			EventType:      "INVOICE_INGESTED",
			Actor:          "api_key_gateway",
			TimestampUTC:   time.Now().UTC(),
			DocumentSHA256: docSHA256,
			PayloadSummary: fmt.Sprintf("Syntax: %s, Profile: %s", validationResp.Syntax, profile),
			PreviousHash:   "0000000000000000000000000000000000000000000000000000000000000000",
		}
		ev.CurrentHash = evidence.CalculateChainHash(&ev)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(IngestionResponse{
			Status:         "accepted",
			InvoiceID:      invoiceNumber,
			TenantID:       tenantID,
			DocumentSHA256: docSHA256,
			AuditHash:      ev.CurrentHash,
			Validation:     validationResp,
		})
	})

	return middleware.RequireAPIKey(keyStore)(mux)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	keyStore := &InMemoryKeyStore{}
	handler := setupRouter(keyStore)

	log.Printf("[READY] E-Invoicing Gateway démarrée sur le port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
