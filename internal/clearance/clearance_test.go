package clearance_test

import (
	"context"
	"strings"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/clearance"
)

func TestPPFConnector_SubmitAndCheck(t *testing.T) {
	conn := clearance.NewPPFConnector(clearance.PPFConfig{
		BaseURL:  "https://api.sandbox.chorus-pro.gouv.fr",
		ClientID: "test-client",
	})

	ctx := context.Background()
	payload := []byte("<Invoice>Test-PPF</Invoice>")

	resp, err := conn.SubmitInvoice(ctx, "INV-001", payload)
	if err != nil {
		t.Fatalf("échec SubmitInvoice PPF: %v", err)
	}
	if !strings.HasPrefix(resp.ClearanceID, "PPF-FR-") {
		t.Errorf("format ClearanceID invalide: %s", resp.ClearanceID)
	}
	if resp.Status != "CLEARED" {
		t.Errorf("attendu status=CLEARED, obtenu %s", resp.Status)
	}

	// Consultation du statut
	checkResp, err := conn.CheckStatus(ctx, resp.ClearanceID)
	if err != nil {
		t.Fatalf("échec CheckStatus PPF: %v", err)
	}
	if checkResp.ClearanceID != resp.ClearanceID {
		t.Errorf("id non concordant")
	}
}

func TestKSeFConnector_SubmitAndUPO(t *testing.T) {
	conn := clearance.NewKSeFConnector(clearance.KSeFConfig{
		NIP:            "1234567890",
		EnvironmentURL: "https://ksef-test.mf.gov.pl/api",
	})

	ctx := context.Background()
	payload := []byte("<Faktura>Test-KSeF</Faktura>")

	resp, err := conn.SubmitInvoice(ctx, "INV-POL-01", payload)
	if err != nil {
		t.Fatalf("échec SubmitInvoice KSeF: %v", err)
	}
	if !strings.HasPrefix(resp.ClearanceID, "1234567890-") {
		t.Errorf("format ClearanceID KSeF incorrect: %s", resp.ClearanceID)
	}
	if len(resp.UPODocument) == 0 {
		t.Fatalf("document UPO manquant")
	}
	if !strings.Contains(string(resp.UPODocument), "<NumerKSeF>") {
		t.Errorf("format UPO XML non conforme: %s", string(resp.UPODocument))
	}
}

func TestDefaultService_ReturnsNotConfigured(t *testing.T) {
	svc := clearance.NewService()
	_, err := svc.SubmitInvoice(context.Background(), "INV-01", []byte("data"))
	if err != clearance.ErrNotConfigured {
		t.Errorf("attendu ErrNotConfigured, obtenu %v", err)
	}
}