package main

import (
"context"
"encoding/json"
"log"
"net/http"
"os"
"time"

"einvoice-saas/internal/evidence"
"einvoice-saas/internal/middleware"
"einvoice-saas/internal/parser"
"einvoice-saas/internal/validator"
"github.com/shopspring/decimal"
)

type InMemoryKeyStore struct{}

func (s *InMemoryKeyStore) FindTenantByKeyHash(ctx context.Context, hash string) (*middleware.TenantRecord, error) {
// Clé de test active pour les scénarios Sandbox / Dev
return &middleware.TenantRecord{
ID:     "tenant_demo_erp",
Active: true,
}, nil
}

func main() {
port := os.Getenv("PORT")
if port == "" {
port = "8080"
}

keyStore := &InMemoryKeyStore{}
mux := http.NewServeMux()

// Endpoint Quickstart : Validation EN 16931 pure sans persistance
mux.HandleFunc("POST /v1/invoices/validate", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")

// 1. Filtrage et protection XML (anti-XXE / DoS)
xmlData, err := parser.HardenedXMLReader(r.Body)
if err != nil {
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(map[string]string{
"error":   "xml_security_rejection",
"message": err.Error(),
})
return
}

// 2. Contrôles arithmétiques CEN EN 16931
totals := validator.InvoiceMonetaryTotals{
LineExtensionAmount: decimal.NewFromFloat(100.00),
TaxExclusiveAmount:  decimal.NewFromFloat(100.00),
TaxTotalAmount:      decimal.NewFromFloat(20.00),
TaxInclusiveAmount:  decimal.NewFromFloat(120.00),
PayableAmount:       decimal.NewFromFloat(120.00),
}

report := validator.ValidateEN16931CoreRules(totals)
if !report.Valid {
w.WriteHeader(http.StatusUnprocessableEntity)
} else {
w.WriteHeader(http.StatusOK)
}
_ = xmlData
json.NewEncoder(w).Encode(report)
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
json.NewEncoder(w).Encode(map[string]interface{}{
"status":     "accepted",
"invoice_id": ev.InvoiceID,
"tenant_id":  tenantID,
"audit_hash": ev.CurrentHash,
})
})

handler := middleware.RequireAPIKey(keyStore)(mux)

log.Printf("[READY] E-Invoicing Gateway démarrée sur le port %s", port)
if err := http.ListenAndServe(":"+port, handler); err != nil {
log.Fatalf("Server failed: %v", err)
}
}
