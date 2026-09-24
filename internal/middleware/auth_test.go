package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	apiKey := "secret-api-key-123"
	protected := AuthMiddleware(dummyHandler, apiKey)

	// 1. Rejet sans en-tête d'authentification
	req1, _ := http.NewRequest("GET", "/protected", nil)
	rr1 := httptest.NewRecorder()
	protected.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusUnauthorized {
		t.Errorf("Attendu 401 Unauthorized, reçu %d", rr1.Code)
	}

	// 2. Succès avec en-tête X-API-Key
	req2, _ := http.NewRequest("GET", "/protected", nil)
	req2.Header.Set("X-API-Key", apiKey)
	rr2 := httptest.NewRecorder()
	protected.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Attendu 200 OK, reçu %d", rr2.Code)
	}

	// 3. Succès avec Authorization: Bearer
	req3, _ := http.NewRequest("GET", "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+apiKey)
	rr3 := httptest.NewRecorder()
	protected.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("Attendu 200 OK, reçu %d", rr3.Code)
	}
}
