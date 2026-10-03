package edifact

import (
	"testing"
)

func TestParse_EDIFACT(t *testing.T) {
	t.Run("Message EDIFACT valide avec BGM et DTM", func(t *testing.T) {
		ediData := []byte("UNB+UNOA:1+SENDER+RECEIVER'UNH+1+INVOIC:D:96A:UN'BGM+380+INV-EDI-2026+9'DTM+137:20261001:102'UNT+5+1'UNZ+1+1'")

		inv, err := Parse(ediData)
		if err != nil {
			t.Fatalf("Parse() a échoué: %v", err)
		}
		if inv.InvoiceNumber != "INV-EDI-2026" {
			t.Errorf("attendu InvoiceNumber 'INV-EDI-2026', obtenu '%s'", inv.InvoiceNumber)
		}
		if inv.Status != "parsed" {
			t.Errorf("attendu Status 'parsed', obtenu '%s'", inv.Status)
		}
	})

	t.Run("Message sans numéro BGM utilise le fallback UNKNOWN", func(t *testing.T) {
		ediData := []byte("UNB+UNOA:1+SENDER+RECEIVER'DTM+137:20261001:102'UNZ+1+1'")

		inv, err := Parse(ediData)
		if err != nil {
			t.Fatalf("Parse() a échoué: %v", err)
		}
		if inv.InvoiceNumber != "UNKNOWN" {
			t.Errorf("attendu InvoiceNumber 'UNKNOWN', obtenu '%s'", inv.InvoiceNumber)
		}
	})

	t.Run("Erreur sur document vide", func(t *testing.T) {
		emptyData := []byte("   \n\t  ")
		_, err := Parse(emptyData)
		if err == nil {
			t.Fatal("attendu erreur sur document vide, obtenu nil")
		}
	})
}
