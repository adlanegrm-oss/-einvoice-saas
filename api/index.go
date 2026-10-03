package handler

import (
"fmt"
"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
fmt.Fprintln(w, `{"status":"ok","service":"einvoice-saas","version":"v0.2.0-alpha"}`)
}