package clearance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPISTETokenManager_FetchAndCache(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Unexpected content-type: %s", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"mock_piste_token_12345","token_type":"Bearer","expires_in":3600}`))
	}))
	defer ts.Close()

	mgr := NewPISTETokenManager(PISTEOAuthConfig{
		TokenURL:     ts.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})

	tok1, err := mgr.GetToken(context.Background())
	if err != nil || tok1 != "mock_piste_token_12345" {
		t.Fatalf("Failed to fetch token: %v", err)
	}

	// 2ème appel doit utiliser le cache
	tok2, err := mgr.GetToken(context.Background())
	if err != nil || tok2 != "mock_piste_token_12345" {
		t.Fatalf("Failed cached token: %v", err)
	}

	if calls != 1 {
		t.Errorf("Expected 1 HTTP call, got %d", calls)
	}
}
