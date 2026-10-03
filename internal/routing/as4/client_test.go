package as4

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
)

func TestAS4Client_SendPayload(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/soap+xml") {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<eb3:SignalMessage xmlns:eb3="http://docs.oasis-open.org/ebxml-msg/ebms/v3.0/ns/core/200704/"><eb3:Receipt/></eb3:SignalMessage>`))
	}))
	defer mockServer.Close()

	client := NewAS4Client()
	ep := &dispatcher.TargetEndpoint{
		ReceiverID:   "11111111111111",
		PlatformName: "PDP_DEST_TEST",
		AS4Endpoint:  mockServer.URL,
	}

	receipt, err := client.SendPayload(context.Background(), ep, []byte("<Invoice>Test AS4 Payload</Invoice>"))
	if err != nil {
		t.Fatalf("echec envoi AS4 : %v", err)
	}

	if !strings.HasPrefix(receipt, "NRR-") {
		t.Errorf("attendu prefixe de recepisse NRR-, obtenu : %s", receipt)
	}
}

func TestAS4Client_SendPayload_Simulated(t *testing.T) {
	client := NewAS4Client()
	ep := &dispatcher.TargetEndpoint{
		ReceiverID:   "11111111111111",
		PlatformName: "PDP_SIMULATED",
		AS4Endpoint:  "mock://as4-test",
	}

	receipt, err := client.SendPayload(context.Background(), ep, []byte("<Invoice>Test</Invoice>"))
	if err != nil {
		t.Fatalf("echec simulation AS4 : %v", err)
	}

	if !strings.HasPrefix(receipt, "NRR-SIMULATED-ACK-") {
		t.Errorf("attendu format simulation, obtenu : %s", receipt)
	}
}