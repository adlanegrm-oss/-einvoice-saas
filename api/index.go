package handler

import (
"encoding/json"
"net/http"
)

type HealthResponse struct {
Status  string `json:"status"`
Service string `json:"service"`
Version string `json:"version"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(HealthResponse{
Status:  "ok",
Service: "einvoice-saas",
Version: "v0.2.0-alpha",
})
}
