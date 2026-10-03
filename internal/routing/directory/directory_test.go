package directory

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestCachedDirectory_Lookup(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("echec ouverture db memoire : %v", err)
	}
	defer db.Close()

	serverCalls := 0
	mockPPF := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalls++
		if r.URL.Path == "/api/v1/directory/99999999999999" {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"receiver_id": "12345678901234",
			"platform_name": "PDP_PARTENAIRE_X",
			"as4_endpoint": "https://as4.partenaire.fr/inbound"
		}`))
	}))
	defer mockPPF.Close()

	cd, err := NewCachedDirectory(db, mockPPF.URL, "token-secret", 1*time.Hour)
	if err != nil {
		t.Fatalf("erreur NewCachedDirectory : %v", err)
	}

	ctx := context.Background()

	// 1. Premier lookup (appel distant)
	ep, err := cd.Lookup(ctx, "12345678901234")
	if err != nil {
		t.Fatalf("erreur premier lookup : %v", err)
	}
	if ep.PlatformName != "PDP_PARTENAIRE_X" {
		t.Errorf("attendu PDP_PARTENAIRE_X, obtenu %s", ep.PlatformName)
	}
	if serverCalls != 1 {
		t.Errorf("attendu 1 appel serveur, obtenu %d", serverCalls)
	}

	// 2. Second lookup (doit provenir du cache SQLite, sans nouvel appel HTTP)
	epCached, err := cd.Lookup(ctx, "12345678901234")
	if err != nil {
		t.Fatalf("erreur second lookup : %v", err)
	}
	if epCached.AS4Endpoint != ep.AS4Endpoint {
		t.Errorf("endpoint different dans le cache")
	}
	if serverCalls != 1 {
		t.Errorf("le cache aurait du eviter l'appel distant, total appels = %d", serverCalls)
	}

	// 3. SIRET inexistant
	_, err = cd.Lookup(ctx, "99999999999999")
	if err == nil {
		t.Errorf("attendu erreur pour siret inexistant, eu nil")
	}
}
