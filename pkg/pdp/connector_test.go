package pdp

import (
"context"
"net/http"
"net/http/httptest"
"testing"
)

func TestPDPClientSubmit(t *testing.T) {
ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if r.Header.Get("X-Idempotency-Key") != "idem-123" {
w.WriteHeader(http.StatusBadRequest)
return
}
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"transmission_id":"TX-999","lifecycle_code":"200","status":"VALIDE"}`))
}))
defer ts.Close()

client := NewPDPClient(ts.URL, "dummy-key")
res, err := client.SubmitInvoice(context.Background(), SubmissionPayload{
TenantID:       "tenant-1",
InvoiceNumber:  "INV-100",
IdempotencyKey: "idem-123",
})

if err != nil || res.TransmissionID != "TX-999" {
t.Fatalf("echec soumission PDP: %v", err)
}
}