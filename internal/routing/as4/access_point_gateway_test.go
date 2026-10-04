package as4

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAccessPointGateway_Dispatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message_id": "MSG-PEPPOL-999",
			"transmission_id": "TX-AS4-888",
			"status": "DELIVERED",
			"delivered_at": "2026-10-04T02:00:00Z",
			"nrr_receipt": "<ebms3:NonRepudiationInformation/>"
		}`))
	}))
	defer ts.Close()

	gateway := NewAccessPointGateway(ts.URL, nil)
	receipt, err := gateway.DispatchPeppolInvoice(context.Background(), PeppolTransmissionRequest{
		SenderID:   "0002:73204903600045",
		ReceiverID: "0002:80214589000012",
		PayloadXML: "<Invoice/>",
	})

	if err != nil {
		t.Fatalf("Unexpected gateway failure: %v", err)
	}
	if receipt.Status != "DELIVERED" {
		t.Errorf("Expected DELIVERED, got %s", receipt.Status)
	}
	if receipt.MessageID != "MSG-PEPPOL-999" {
		t.Errorf("Expected MSG-PEPPOL-999, got %s", receipt.MessageID)
	}
}
