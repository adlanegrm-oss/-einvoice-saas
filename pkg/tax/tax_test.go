package tax_test

import (
	"github.com/adlanegrm-oss/einvoice-saas/pkg/money"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/tax"
	"testing"
)

func TestCalculateVATBreakdown(t *testing.T) {
	// Cas 1 : Standard 20% sur 100.00 EUR
	base := money.New(10000, money.EUR)
	st, err := tax.CalculateVATBreakdown(tax.StandardRate, tax.FromFloatPercent(20.0), base, "", "")
	if err != nil || st.TaxAmount.Cents() != 2000 {
		t.Fatalf("Calcul TVA 20%% incorrect: %v, err: %v", st.TaxAmount, err)
	}

	// Cas 2 : Autoliquidation (AE) sans motif -> doit échouer
	_, err = tax.CalculateVATBreakdown(tax.ReverseCharge, 0, base, "", "")
	if err == nil {
		t.Fatalf("L'autoliquidation sans motif aurait dû échouer (BR-AE-01)")
	}

	// Cas 3 : Autoliquidation (AE) avec motif -> TVA 0 EUR
	stAE, err := tax.CalculateVATBreakdown(tax.ReverseCharge, 0, base, "Autoliquidation - Art. 283 du CGI", "VATEX-EU-AE")
	if err != nil || stAE.TaxAmount.Cents() != 0 {
		t.Fatalf("Erreur calcul autoliquidation: %v", err)
	}
}
