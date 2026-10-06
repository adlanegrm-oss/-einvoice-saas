package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/service"
	"einvoice-saas/internal/validator"
)

func loadTestXML(t *testing.T) []byte {
	paths := []string{
		"../../factures_test_lots/01_UBL_EN16931_CONFORME.xml",
		"../factures_test_lots/01_UBL_EN16931_CONFORME.xml",
		"../../factures_test_lots/FACT_2026_001_CONFORME.xml",
		"../factures_test_lots/FACT_2026_001_CONFORME.xml",
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil && len(data) > 200 {
			return data
		}
	}
	t.Skip("Fixture XML conforme introuvable dans factures_test_lots")
	return nil
}

func buildTestRouter() http.Handler {
	keyStore := &InMemoryKeyStore{}
	schematronEngine := validator.NewSchematronEngine(nil)
	normativeValidator := validator.NewNormativeValidator(true)
	frFiscalValidator := fr.NewFranceCanonicalValidator()

	valFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		resp, err := executeValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
		if err != nil {
			return "", nil, false, resp, err
		}
		return resp.Syntax, resp.CanonicalInvoice, resp.Valid, resp, nil
	}

	invoiceSvc := service.NewInvoiceService(
		newInMemInvoiceRepo(),
		newInMemEventRepo(),
		newInMemIdemRepo(),
		valFn,
	)

	return setupRouter(keyStore, invoiceSvc)
}

func TestGatewayEndpoints_Integration(t *testing.T) {
	handler := buildTestRouter()
	validToken := "sk_test_demo_live_gateway_token_123456789"

	t.Run("Validate_Rejet_Sans_Cle_API", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate", bytes.NewBuffer([]byte("<dummy/>")))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attendu 401 Unauthorized, obtenu: %d", rec.Code)
		}
	})

	t.Run("Validate_Rejet_Payload_XML_Invalide", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate", bytes.NewBuffer([]byte("not-xml")))
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("attendu 400 ou 422, obtenu: %d", rec.Code)
		}
	})

	t.Run("Ingest_Rejet_Facture_Invalide", func(t *testing.T) {
		invalidXML := `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
    <CustomizationID>urn:cen.eu:en16931:2017</CustomizationID>
    <ID>INV-FAIL-01</ID>
</Invoice>`

		req := httptest.NewRequest(http.MethodPost, "/v1/invoices", bytes.NewBufferString(invalidXML))
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("attendu 422 Unprocessable Entity, obtenu: %d", rec.Code)
		}

		var resp service.IngestionResult
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("JSON invalide: %v", err)
		}
		if resp.Status != "rejected" {
			t.Fatalf("attendu status 'rejected', obtenu: %v", resp.Status)
		}
	})

	t.Run("HappyPath_Validate_UBL_Conforme", func(t *testing.T) {
		validXML := loadTestXML(t)

		req := httptest.NewRequest(http.MethodPost, "/v1/invoices/validate?profile=EN16931", bytes.NewReader(validXML))
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("attendu 200 OK, obtenu %d: %s", rec.Code, rec.Body.String())
		}

		var resp UnifiedValidationResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("erreur deserialisation JSON: %v", err)
		}

		if !resp.Valid {
			t.Fatalf("la facture conforme a ete marquee invalide")
		}
		if resp.Syntax != "UBL-2.1" {
			t.Errorf("syntaxe attendue UBL-2.1, obtenu: %s", resp.Syntax)
		}
		if resp.CanonicalInvoice == nil {
			t.Fatalf("modele canonique manquant dans la reponse")
		}
	})

	t.Run("HappyPath_Ingest_UBL_SHA256_Exact", func(t *testing.T) {
		validXML := loadTestXML(t)

		h := sha256.Sum256(validXML)
		expectedSHA256 := hex.EncodeToString(h[:])

		req := httptest.NewRequest(http.MethodPost, "/v1/invoices?profile=EN16931", bytes.NewReader(validXML))
		req.Header.Set("Authorization", "Bearer "+validToken)
		req.Header.Set("Idempotency-Key", "test_idempotency_key_ubl_001")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("attendu 202 Accepted, obtenu %d: %s", rec.Code, rec.Body.String())
		}

		var resp service.IngestionResult
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("erreur deserialisation JSON: %v", err)
		}

		if resp.Status != "accepted" {
			t.Errorf("statut attendu 'accepted', obtenu: %s", resp.Status)
		}
		if resp.DocumentSHA256 != expectedSHA256 {
			t.Errorf("SHA-256 divergent !\nattendu : %s\nobtenu  : %s", expectedSHA256, resp.DocumentSHA256)
		}
		if resp.AuditHash == "" {
			t.Errorf("audit_hash manquant dans la reponse d'ingestion")
		}
	})
}
