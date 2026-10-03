package handler_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
)

func withAuthContext(next http.Handler, tenant string, role auth.Role) http.Handler {
	secret := []byte("secret-tres-long-pour-les-tests-qualification-32b")
	tm, err := auth.NewTokenManager(secret, time.Hour)
	if err != nil {
		panic(err)
	}

	tokenStr, _, err := tm.Issue("test@example.com", role, tenant)
	if err != nil {
		panic(err)
	}

	authMW := middleware.Authenticate(tm)
	protected := authMW(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+tokenStr)
		protected.ServeHTTP(w, r)
	})
}

func TestIntegration_Deposit_MultiTenant_And_Purge(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatalf("échec création dossier temporaire: %v", err)
	}
	defer os.RemoveAll(tempDir)

	arcHandler, err := handler.NewArchiveHandler(tempDir)
	if err != nil {
		t.Fatalf("échec instanciation ArchiveHandler: %v", err)
	}

	validXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
    <ID>INV-QUALIF-001</ID>
    <IssueDate>2026-10-01</IssueDate>
    <DocumentCurrencyCode>EUR</DocumentCurrencyCode>
    <AccountingSupplierParty><Party><PartyName><Name>Fournisseur A</Name></PartyName></Party></AccountingSupplierParty>
    <AccountingCustomerParty><Party><PartyName><Name>Client A</Name></PartyName></Party></AccountingCustomerParty>
</Invoice>`)

	t.Run("Dépôt conforme Tenant-Alpha -> ACCEPTE", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("files", "facture_alpha.xml")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write(validXML)
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/deposit", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		h := withAuthContext(http.HandlerFunc(arcHandler.Deposit), "tenant-alpha", auth.RoleClient)
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu HTTP 200, obtenu %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["status"] != "success" {
			t.Errorf("attendu status 'success', obtenu '%v'", resp["status"])
		}
	})

	t.Run("Isolation Multi-Tenant : Tenant-Beta ne voit pas les documents de Tenant-Alpha", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/list", nil)

		w := httptest.NewRecorder()
		h := withAuthContext(http.HandlerFunc(arcHandler.List), "tenant-beta", auth.RoleClient)
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu HTTP 200, obtenu %d", w.Code)
		}

		var list []handler.ArchiveEntry
		_ = json.Unmarshal(w.Body.Bytes(), &list)
		if len(list) != 0 {
			t.Errorf("fuite multi-tenant: tenant-beta voit %d documents de tenant-alpha", len(list))
		}
	})

	t.Run("Isolation Multi-Tenant : Tentative de téléchargement direct inter-tenant -> 404", func(t *testing.T) {
		files, _ := os.ReadDir(filepath.Join(tempDir, "tenant-alpha", handler.FolderReady))
		if len(files) == 0 {
			t.Fatal("aucun fichier archivé trouvé pour tenant-alpha")
		}
		storedFilename := files[0].Name()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/download?folder=factures&file="+storedFilename, nil)

		w := httptest.NewRecorder()
		h := withAuthContext(http.HandlerFunc(arcHandler.Download), "tenant-beta", auth.RoleClient)
		h.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("sécurité compromise: attendu 404 pour téléchargement inter-tenant, obtenu %d", w.Code)
		}
	})

	t.Run("Dépôt Draft (Brouillon temporaire 72h) et cycle de purge", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("target", "draft")
		part, _ := writer.CreateFormFile("files", "draft_invoice.xml")
		_, _ = part.Write([]byte("<Invoice>Incomplet</Invoice>"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/deposit", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		h := withAuthContext(http.HandlerFunc(arcHandler.Deposit), "tenant-alpha", auth.RoleClient)
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu HTTP 200 pour draft, obtenu %d: %s", w.Code, w.Body.String())
		}

		draftDir := filepath.Join(tempDir, "tenant-alpha", handler.FolderDraft)
		draftFiles, _ := os.ReadDir(draftDir)
		if len(draftFiles) != 1 {
			t.Fatalf("attendu 1 fichier draft, trouvé %d", len(draftFiles))
		}

		removed := arcHandler.PurgeExpiredDrafts(time.Now(), handler.DraftTTL)
		if removed != 0 {
			t.Errorf("attendu 0 fichier purgé immédiatement, purgé: %d", removed)
		}

		futureTime := time.Now().Add(73 * time.Hour)
		removedFuture := arcHandler.PurgeExpiredDrafts(futureTime, handler.DraftTTL)
		if removedFuture != 1 {
			t.Errorf("attendu 1 fichier purgé après 73h, purgé: %d", removedFuture)
		}
	})
}