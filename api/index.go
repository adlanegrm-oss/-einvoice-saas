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

type ErrorResponse struct {
Error string `json:"error"`
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
var inv validation.InvoiceTotals
if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(ErrorResponse{
Error: fmt.Sprintf("Corps JSON invalide: %v", err),
})
return
}

report := validation.ValidateStrictEN16931(inv)
if !report.Valid {
w.WriteHeader(http.StatusUnprocessableEntity)
json.NewEncoder(w).Encode(report)
return
}

w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(report)

default:
w.WriteHeader(http.StatusMethodNotAllowed)
json.NewEncoder(w).Encode(ErrorResponse{Error: "Methode non autorisee"})
}
}