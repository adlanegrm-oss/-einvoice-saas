package main

import (
"bytes"
"context"
"encoding/json"
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

func setupRouter(keyStore middleware.APIKeyStore) http.Handler {
mux := http.NewServeMux()

schematronEngine := validator.NewSchematronEngine(nil)
normativeValidator := validator.NewNormativeValidator(true)
frFiscalValidator := fr.NewFranceCanonicalValidator()

// Endpoint : Validation complète multi-niveaux (Schematron -> Canonique -> Arithmétique -> Fiscale)
mux.HandleFunc("POST /v1/invoices/validate", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")

// 1. Filtrage et protection XML (anti-XXE / DoS)
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

response := UnifiedValidationResponse{
Valid:         true,
TargetProfile: string(profile),
}

// 2. Contrôle Schematron / SVRL via ValidateProfile
schemReport, err := schematronEngine.ValidateProfile(xmlData, profile)
if err == nil {
response.SchematronReport = schemReport
if !schemReport.Valid {
response.Valid = false
}
}

// 3. Détection de syntaxe & normalisation pivot
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
w.WriteHeader(http.StatusUnprocessableEntity)
_ = json.NewEncoder(w).Encode(map[string]interface{}{
"valid":             false,
"error":             "normalization_failed",
"message":           normErr.Error(),
"schematron_report": response.SchematronReport,
})
return
}

response.CanonicalInvoice = canonical

// 4. Contrôles arithmétiques EN 16931 sur modèle pivot
arithResult, err := normativeValidator.ValidateCanonical(canonical)
if err == nil {
response.ArithmeticReport = arithResult
if !arithResult.Valid {
response.Valid = false
}
}

// 5. Contrôles fiscaux nationaux (France CIUS-FR)
fiscalReport := frFiscalValidator.Validate(canonical)
response.FiscalReport = &fiscalReport
if !fiscalReport.Valid {
response.Valid = false
}

if !response.Valid {
w.WriteHeader(http.StatusUnprocessableEntity)
} else {
w.WriteHeader(http.StatusOK)
}

_ = json.NewEncoder(w).Encode(response)
})

// Endpoint Ingestion : Ingestion + Ledger d'intégrité SHA-256
mux.HandleFunc("POST /v1/invoices", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
tenantID, _ := middleware.GetTenantID(r.Context())

ev := evidence.AuditEvent{
EventID:        "ev_init_" + time.Now().Format("150405"),
TenantID:       tenantID,
InvoiceID:      "inv_mock_001",
EventType:      "INVOICE_INGESTED",
Actor:          "api_key_gateway",
TimestampUTC:   time.Now().UTC(),
DocumentSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
PreviousHash:   "0000000000000000000000000000000000000000000000000000000000000000",
}
ev.CurrentHash = evidence.CalculateChainHash(&ev)

w.WriteHeader(http.StatusAccepted)
_ = json.NewEncoder(w).Encode(map[string]interface{}{
"status":     "accepted",
"invoice_id": ev.InvoiceID,
"tenant_id":  tenantID,
"audit_hash": ev.CurrentHash,
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
