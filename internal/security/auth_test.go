package security_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"einvoice-saas/internal/middleware"
	"einvoice-saas/internal/security"
)

func TestInMemoryKeyStore_ProductionGuard(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("attente d'un panic quand APP_ENV=production, aucun leve")
		}
	}()

	_ = security.NewInMemoryKeyStore()
}

func TestInMemoryKeyStore_ResolutionEtStatuts(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	defer os.Unsetenv("APP_ENV")

	store := security.NewInMemoryKeyStore()
	validToken := "sk_test_token_live_12345678901234567890"
	revokedToken := "sk_test_token_revoked_12345678901234567"
	inactiveToken := "sk_test_token_inactive_1234567890123456"

	store.AddKey(validToken, "tenant_corp_a", true)
	store.AddKey(revokedToken, "tenant_corp_b", true)
	store.RevokeKey(revokedToken)
	store.AddKey(inactiveToken, "tenant_corp_c", false)

	hashOf := func(tok string) string {
		h := sha256.Sum256([]byte(tok))
		return hex.EncodeToString(h[:])
	}

	rec, err := store.FindTenantByKeyHash(context.Background(), hashOf(validToken))
	if err != nil || rec == nil || rec.ID != "tenant_corp_a" {
		t.Fatalf("echec resolution cle valide: %v", err)
	}

	_, err = store.FindTenantByKeyHash(context.Background(), hashOf("sk_unknown_key_9999999999999999999"))
	if err == nil {
		t.Fatalf("une cle inconnue aurait du retourner une erreur")
	}

	_, err = store.FindTenantByKeyHash(context.Background(), hashOf(revokedToken))
	if err == nil {
		t.Fatalf("une cle revoquee aurait du retourner une erreur")
	}

	_, err = store.FindTenantByKeyHash(context.Background(), hashOf(inactiveToken))
	if err == nil {
		t.Fatalf("une cle inactive aurait du retourner une erreur")
	}
}

func TestTenantIsolation(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	defer os.Unsetenv("APP_ENV")

	store := security.NewInMemoryKeyStore()
	tokenA := "sk_live_tenant_alpha_0123456789abcdef"
	tokenB := "sk_live_tenant_bravo_fedcba9876543210"

	store.AddKey(tokenA, "tenant_alpha", true)
	store.AddKey(tokenB, "tenant_bravo", true)

	hashOf := func(tok string) string {
		h := sha256.Sum256([]byte(tok))
		return hex.EncodeToString(h[:])
	}

	recA, errA := store.FindTenantByKeyHash(context.Background(), hashOf(tokenA))
	if errA != nil || recA.ID != "tenant_alpha" {
		t.Fatalf("erreur resolution tenant_alpha: %v", errA)
	}

	recB, errB := store.FindTenantByKeyHash(context.Background(), hashOf(tokenB))
	if errB != nil || recB.ID != "tenant_bravo" {
		t.Fatalf("erreur resolution tenant_bravo: %v", errB)
	}

	if recA.ID == recB.ID {
		t.Fatalf("violation d'isolation: deux cles differentes resolvent le meme tenant")
	}
}

func TestRequireAPIKey_MiddlewareIntegration(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	defer os.Unsetenv("APP_ENV")

	store := security.NewInMemoryKeyStore()
	validToken := "sk_live_authorized_user_1234567890123"
	store.AddKey(validToken, "tenant_prod_1", true)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tID, err := middleware.GetTenantID(r.Context())
		if err != nil || tID != "tenant_prod_1" {
			t.Errorf("tenant ID inattendu dans contexte: %s, err: %v", tID, err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ACCEPTED"))
	})

	handler := middleware.RequireAPIKey(store)(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("attente status 200, recu %d", rec.Code)
	}

	reqNoAuth := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate", nil)
	recNoAuth := httptest.NewRecorder()
	handler.ServeHTTP(recNoAuth, reqNoAuth)

	if recNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("attente status 401 pour header manquant, recu %d", recNoAuth.Code)
	}
}
