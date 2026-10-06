package main

import (
"bytes"
"net/http"
"net/http/httptest"
"testing"
)

func TestValidateEndpoint_Integration(t *testing.T) {
keyStore := &InMemoryKeyStore{}
handler := setupRouter(keyStore)
validToken := "sk_test_demo_live_gateway_token_123456789" // >= 32 caractères

t.Run("Rejet_Sans_Cle_API", func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate", bytes.NewBuffer([]byte("<dummy/>")))
rec := httptest.NewRecorder()

handler.ServeHTTP(rec, req)

if rec.Code != http.StatusUnauthorized {
t.Fatalf("attendu 401 Unauthorized, obtenu: %d", rec.Code)
}
})

t.Run("Rejet_Payload_XML_Invalide", func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate", bytes.NewBuffer([]byte("not-xml")))
req.Header.Set("Authorization", "Bearer "+validToken)
rec := httptest.NewRecorder()

handler.ServeHTTP(rec, req)

if rec.Code != http.StatusUnprocessableEntity && rec.Code != http.StatusBadRequest {
t.Fatalf("attendu 400 ou 422, obtenu: %d", rec.Code)
}
})
}
