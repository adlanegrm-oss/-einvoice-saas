package handler

import (
"encoding/json"
"fmt"
"net/http"

"github.com/adlanegrm-oss/einvoice-saas/pkg/validation"
)

type HealthResponse struct {
Status  string `json:"status"`
Service string `json:"service"`
Version string `json:"version"`
}

type ValidationRequest struct {
Lines        []validation.InvoiceLine `json:"lines"`
TaxSubtotals []validation.TaxSubtotal `json:"tax_subtotals"`
TotalHT      int64                    `json:"total_ht"`
TotalTVA     int64                    `json:"total_tva"`
TotalTTC     int64                    `json:"total_ttc"`
}

type ValidationResponse struct {
Valid  bool     `json:"valid"`
Errors []string `json:"errors,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")

switch r.Method {
case http.MethodGet:
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(HealthResponse{
Status:  "ok",
Service: "einvoice-saas",
Version: "v0.2.0-alpha",
})

case http.MethodPost:
var req ValidationRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(ValidationResponse{
Valid:  false,
Errors: []string{fmt.Sprintf("JSON invalide: %v", err)},
})
return
}

errs := validation.ValidateStrictEN16931(req.Lines, req.TaxSubtotals, req.TotalHT, req.TotalTVA, req.TotalTTC)
if len(errs) > 0 {
w.WriteHeader(http.StatusUnprocessableEntity)
json.NewEncoder(w).Encode(ValidationResponse{
Valid:  false,
Errors: errs,
})
return
}

w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(ValidationResponse{Valid: true})

default:
w.WriteHeader(http.StatusMethodNotAllowed)
json.NewEncoder(w).Encode(map[string]string{"error": "Methode non autorisee"})
}
}