package model_test

import (
	"os"
	"testing"

	"einvoice-saas/internal/model"
)

func TestNormalizeUBLToCanonical_Morocco(t *testing.T) {
	data, err := os.ReadFile("../../factures_test_lots/ma/test_facture_maroc_ok.xml")
	if err != nil {
		t.Fatalf("Impossible de lire le jeu d'essai UBL: %v", err)
	}

	cinv, err := model.NormalizeUBLToCanonical(data)
	if err != nil {
		t.Fatalf("Échec de normalisation vers CanonicalInvoice: %v", err)
	}

	if cinv.TargetJurisdiction != "MA" {
		t.Errorf("Juridiction attendue MA, obtenu: %s", cinv.TargetJurisdiction)
	}

	if cinv.Seller.NationalID != "001524368000045" {
		t.Errorf("ICE Fournisseur attendu 001524368000045, obtenu: %s", cinv.Seller.NationalID)
	}

	if cinv.Totals.PayableAmount != 1200.00 {
		t.Errorf("Montant net à payer attendu 1200.00, obtenu: %f", cinv.Totals.PayableAmount)
	}
}
