package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"einvoice-saas/internal/middleware"
	"einvoice-saas/internal/security"
)

const apiKeyTest = "cle-api-de-test-0123456789-abcdefghij"

func newAuthStore(t *testing.T) *security.InMemoryKeyStore {
	t.Helper()

	t.Setenv("APP_ENV", "test")

	s := security.NewInMemoryKeyStore()

	s.AddKey(
		apiKeyTest,
		"tenant-42",
		true,
	)

	s.AddKey(
		apiKeyTest+"-inactive",
		"tenant-1",
		false,
	)

	s.AddKey(
		apiKeyTest+"-revoked",
		"tenant-2",
		true,
	)

	s.RevokeKey(
		apiKeyTest + "-revoked",
	)

	return s
}

func TestRequireAPIKey(t *testing.T) {
	store := newAuthStore(t)

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantTenant string
	}{
		{
			"header absent",
			"",
			401,
			"",
		},
		{
			"schema Basic",
			"Basic " + apiKeyTest,
			401,
			"",
		},
		{
			"sans espace",
			"Bearer",
			401,
			"",
		},
		{
			"token trop court",
			"Bearer court",
			401,
			"",
		},
		{
			"cle inconnue",
			"Bearer " + strings.Repeat("x", 40),
			401,
			"",
		},
		{
			"cle inactive",
			"Bearer " + apiKeyTest + "-inactive",
			401,
			"",
		},
		{
			"cle revoquee",
			"Bearer " + apiKeyTest + "-revoked",
			401,
			"",
		},
		{
			"cle valide",
			"Bearer " + apiKeyTest,
			200,
			"tenant-42",
		},
		{
			"bearer en minuscules",
			"bearer " + apiKeyTest,
			200,
			"tenant-42",
		},
		{
			"espaces autour du token",
			"Bearer   " + apiKeyTest + "  ",
			200,
			"tenant-42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var gotTenant string

			next := http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					called = true

					gotTenant, _ =
						middleware.GetTenantID(
							r.Context(),
						)

					w.WriteHeader(http.StatusOK)
				},
			)

			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			if tc.header != "" {
				req.Header.Set(
					"Authorization",
					tc.header,
				)
			}

			rr := httptest.NewRecorder()

			middleware.RequireAPIKey(store)(
				next,
			).ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf(
					"status = %d, attendu %d (body=%s)",
					rr.Code,
					tc.wantStatus,
					rr.Body.String(),
				)
			}

			if called != (tc.wantStatus == 200) {
				t.Fatalf(
					"next appele = %v, incoherent avec le status",
					called,
				)
			}

			if gotTenant != tc.wantTenant {
				t.Fatalf(
					"tenant = %q, attendu %q",
					gotTenant,
					tc.wantTenant,
				)
			}
		})
	}
}

type nilTenantStore struct{}

func (nilTenantStore) FindTenantByKeyHash(
	context.Context,
	string,
) (*middleware.TenantRecord, error) {
	return nil, nil
}

func TestRequireAPIKey_StoreReturnsNilTenant(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+apiKeyTest,
	)

	rr := httptest.NewRecorder()

	next := http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			t.Fatal(
				"next ne doit pas etre appele",
			)
		},
	)

	middleware.RequireAPIKey(
		nilTenantStore{},
	)(
		next,
	).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d",
			rr.Code,
		)
	}
}

func TestGetTenantID(t *testing.T) {
	key := middleware.TenantIDContextKey
	bg := context.Background()

	if _, err := middleware.GetTenantID(bg); err == nil {
		t.Error("erreur attendue : tenant absent")
	}

	if _, err := middleware.GetTenantID(
		context.WithValue(bg, key, 42),
	); err == nil {
		t.Error("erreur attendue : mauvais type")
	}

	if _, err := middleware.GetTenantID(
		context.WithValue(bg, key, ""),
	); err == nil {
		t.Error("erreur attendue : tenant vide")
	}

	id, err := middleware.GetTenantID(
		context.WithValue(bg, key, "tenant-9"),
	)

	if err != nil || id != "tenant-9" {
		t.Fatalf(
			"id=%q err=%v",
			id,
			err,
		)
	}
}
