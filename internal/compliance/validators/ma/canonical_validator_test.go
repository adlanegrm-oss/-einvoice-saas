package ma_test

import (
	"os"
	"testing"

	"einvoice-saas/internal/compliance/validators/ma"
	"einvoice-saas/internal/model"
)

func TestMoroccoCanonicalValidator_FromUBL(t *testing.T) {
	data, err := os.ReadFile("../../../../factures_test_lots/ma/test_facture_maroc_ok.xml")
	if err != nil {
		t.Fatalf("Erreur lecture lot UBL: %v", err)
	}

	cinv, err := model.NormalizeUBLToCanonical(data)
	if err != nil {
		t.Fatalf("Erreur normalisation pivot: %v", err)
	}

	validator := ma.NewMoroccoCanonicalValidator()
	report := validator.Validate(cinv)

	if !report.Valid {
		t.Fatalf("Facture attendue valide, erreurs trouvées: %+v", report.Issues)
	}

	if report.Jurisdiction != "MA" {
		t.Errorf("Juridiction attendue MA, obtenu: %s", report.Jurisdiction)
	}
}
