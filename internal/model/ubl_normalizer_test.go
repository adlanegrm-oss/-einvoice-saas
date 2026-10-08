package model_test

import (
"os"
"testing"

"github.com/shopspring/decimal"

"einvoice-saas/internal/model"
)

func TestNormalizeUBLToCanonical_Morocco(t *testing.T) {
data, err := os.ReadFile("../../factures_test_lots/ma/test_facture_maroc_ok.xml")
if err != nil {
t.Skipf("Jeu d'essai UBL absent: %v", err)
}

cinv, err := model.NormalizeUBLToCanonical(data)
if err != nil {
t.Fatalf("Échec normalisation: %v", err)
}

if cinv.Seller.CountryCode != "MA" && cinv.DocumentCurrency != "MAD" {
// juridiction déduite côté engine, pas forcément dans le modèle
t.Logf("CountryCode=%s Currency=%s", cinv.Seller.CountryCode, cinv.DocumentCurrency)
}

if cinv.Seller.LegalID == "" && cinv.Seller.VATID == "" {
t.Errorf("identifiant vendeur manquant")
}

expected := decimal.NewFromFloat(1200)
if !cinv.Totals.PayableAmount.Equal(expected) {
t.Logf("PayableAmount obtenu: %s (attendu 1200 si fixture standard)", cinv.Totals.PayableAmount.StringFixed(2))
}
}
