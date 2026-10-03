package handler

import (
"fmt"
"net/http"
)

// Handler est le point d'entree invoque par le runtime Serverless Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, "{\"status\":\"ok\",\"service\":\"einvoice-saas\",\"version\":\"v0.2.0-alpha\"}\n")
}