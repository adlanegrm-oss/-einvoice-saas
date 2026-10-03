package currency

import (
"testing"
)

func TestPrecisionAndRounding(t *testing.T) {
// Cas typique où float64 introduit des résidus (ex: 0.1 + 0.2 != 0.3)
price, err := NewAmount("19.99")
if err != nil {
t.Fatalf("Erreur création montant: %v", err)
}

qty, _ := NewAmount("3")
totalHT := price.Mul(qty).RoundTo2()

expectedHT, _ := NewAmount("59.97")
if !totalHT.Equal(expectedHT.Decimal) {
t.Fatalf("Erreur calcul HT: attendu %s, obtenu %s", expectedHT.StringFixed(2), totalHT.StringFixed(2))
}

// Calcul TVA 20%
tvaRate, _ := NewAmount("0.20")
tvaAmount := totalHT.Mul(tvaRate).RoundTo2()

expectedTVA, _ := NewAmount("11.99") // 59.97 * 0.20 = 11.994 -> 11.99
if !tvaAmount.Equal(expectedTVA.Decimal) {
t.Fatalf("Erreur calcul TVA: attendu %s, obtenu %s", expectedTVA.StringFixed(2), tvaAmount.StringFixed(2))
}

totalTTC := totalHT.Add(tvaAmount)
expectedTTC, _ := NewAmount("71.96")
if !totalTTC.Equal(expectedTTC.Decimal) {
t.Fatalf("Erreur calcul TTC: attendu %s, obtenu %s", expectedTTC.StringFixed(2), totalTTC.StringFixed(2))
}
}
