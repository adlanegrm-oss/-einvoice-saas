package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
)

func setup(t *testing.T) (*auth.TokenManager, http.Handler, http.Handler) {
	t.Helper()
	tm, err := auth.NewTokenManager([]byte(strings.Repeat("k", 32)), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	shared := Protect(tm, ok, auth.RoleAdmin, auth.RoleClient)
	adminOnly := Protect(tm, ok, auth.RoleAdmin)
	return tm, shared, adminOnly
}

func do(h http.Handler, headers map[string]string) int {
	req := httptest.NewRequest("GET", "/test", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestAuthenticate(t *testing.T) {
	tm, shared, adminOnly := setup(t)
	clientTok, _, _ := tm.Issue("c@example.com", auth.RoleClient, "t-1")
	adminTok, _, _ := tm.Issue("a@example.com", auth.RoleAdmin, "t-0")

	cases := []struct {
		name    string
		handler http.Handler
		headers map[string]string
		want    int
	}{
		{"sans jeton -> 401", shared, nil, 401},
		{"en-tête X-User-Role forgé -> 401", shared, map[string]string{"X-User-Role": "ADMIN"}, 401},
		{"Authorization: ADMIN (ancien schéma) -> 401", shared, map[string]string{"Authorization": "ADMIN"}, 401},
		{"jeton bidon -> 401", shared, map[string]string{"Authorization": "Bearer abc.def.ghi"}, 401},
		{"jeton client sur route partagée -> 200", shared, map[string]string{"Authorization": "Bearer " + clientTok}, 200},
		{"jeton client sur route admin -> 403", adminOnly, map[string]string{"Authorization": "Bearer " + clientTok}, 403},
		{"jeton admin sur route admin -> 200", adminOnly, map[string]string{"Authorization": "Bearer " + adminTok}, 200},
		{"préfixe bearer insensible à la casse -> 200", shared, map[string]string{"Authorization": "bearer " + clientTok}, 200},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := do(c.handler, c.headers); got != c.want {
				t.Errorf("attendu %d, reçu %d", c.want, got)
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	tm, err := auth.NewTokenManager([]byte(strings.Repeat("k", 32)), time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	tok, _, _ := tm.Issue("c@example.com", auth.RoleClient, "t-1")
	time.Sleep(1100 * time.Millisecond) // les jetons ont une précision à la seconde
	h := Protect(tm, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), auth.RoleClient)
	if got := do(h, map[string]string{"Authorization": "Bearer " + tok}); got != 401 {
		t.Errorf("jeton expiré : attendu 401, reçu %d", got)
	}
}

func TestClaimsInContext(t *testing.T) {
	tm, _, _ := setup(t)
	tok, _, _ := tm.Issue("c@example.com", auth.RoleClient, "t-42")
	var seen *auth.Claims
	h := Protect(tm, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = ClaimsFrom(r.Context())
	}), auth.RoleClient)
	do(h, map[string]string{"Authorization": "Bearer " + tok})
	if seen == nil || seen.Tenant != "t-42" {
		t.Errorf("claims absents ou incorrects : %+v", seen)
	}
}

func TestRateLimit(t *testing.T) {
	l := auth.NewLimiter(2, time.Minute)
	h := RateLimit(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if do(h, nil) != 200 || do(h, nil) != 200 {
		t.Fatal("les deux premières requêtes doivent passer")
	}
	if got := do(h, nil); got != 429 {
		t.Errorf("attendu 429, reçu %d", got)
	}
}

func TestRecover(t *testing.T) {
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") }))
	if got := do(h, nil); got != 500 {
		t.Errorf("attendu 500, reçu %d", got)
	}
}
