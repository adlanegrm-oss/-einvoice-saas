package as4

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAS4Client_RetryAndSuccess(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&calls, 1)
		if current < 2 {
			w.WriteHeader(http.StatusServiceUnavailable) // simule échec transitoire
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<eb3:Receipt/>"))
	}))
	defer server.Close()

	client := NewAS4Client(server.URL, nil, nil)
	client.BaseBackoff = 10 * time.Millisecond

	msg := AS4Message{
		MessageID:   "MSG-AS4-TEST-1",
		SenderParty: "CEGEDIM",
		ReceiverParty: "CHORUS",
		Payload:     []byte("<Invoice>Test</Invoice>"),
	}

	receipt, err := client.SendMessageWithRetry(context.Background(), msg)
	if err != nil {
		t.Fatalf("Expected success on retry, got: %v", err)
	}

	if receipt.RefToMsgID != "MSG-AS4-TEST-1" {
		t.Errorf("Unexpected RefToMsgID: %s", receipt.RefToMsgID)
	}
	if len(client.DLQ) != 0 {
		t.Errorf("DLQ should be empty on success")
	}
}

func TestAS4Client_RoutesToDLQOnFinalFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAS4Client(server.URL, nil, nil)
	client.BaseBackoff = 5 * time.Millisecond
	client.MaxRetries = 2

	msg := AS4Message{
		MessageID: "MSG-FAIL-DLQ",
		Payload:   []byte("<Invoice/>"),
	}

	_, err := client.SendMessageWithRetry(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected transmission to fail")
	}

	if len(client.DLQ) != 1 {
		t.Fatalf("Expected 1 item in DLQ, got %d", len(client.DLQ))
	}
	if client.DLQ[0].Message.MessageID != "MSG-FAIL-DLQ" {
		t.Errorf("DLQ item mismatch: %s", client.DLQ[0].Message.MessageID)
	}
}
