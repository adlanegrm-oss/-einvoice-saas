package ubl

import (
	"testing"
)

func TestParse_UBL(t *testing.T) {
	t.Run("Facture UBL valide avec PartyName", func(t *testing.T) {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
    <ID>INV-2026-001</ID>
    <IssueDate>2026-10-01</IssueDate>
    <DocumentCurrencyCode>EUR</DocumentCurrencyCode>
    <AccountingSupplierParty>
        <Party>
            <PartyName>
                <Name>Societe Vendeur SAS</Name>
            </PartyName>
        </Party>
    </AccountingSupplierParty>
    <AccountingCustomerParty>
        <Party>
            <PartyName>
                <Name>Societe Acheteur SAS</Name>
            </PartyName>
        </Party>
    </AccountingCustomerParty>
</Invoice>`)

		inv, err := Parse(xmlData)
		if err != nil {
			t.Fatalf("Parse() a échoué: %v", err)
		}
		if inv.InvoiceNumber != "INV-2026-001" {
			t.Errorf("attendu InvoiceNumber 'INV-2026-001', obtenu '%s'", inv.InvoiceNumber)
		}
		if inv.Currency != "EUR" {
			t.Errorf("attendu Currency 'EUR', obtenu '%s'", inv.Currency)
		}
		if inv.Seller.Name != "Societe Vendeur SAS" {
			t.Errorf("attendu Seller.Name 'Societe Vendeur SAS', obtenu '%s'", inv.Seller.Name)
		}
		if inv.Buyer.Name != "Societe Acheteur SAS" {
			t.Errorf("attendu Buyer.Name 'Societe Acheteur SAS', obtenu '%s'", inv.Buyer.Name)
		}
	})

	t.Run("Repli sur PartyLegalEntity/RegistrationName si PartyName absent (EN 16931)", func(t *testing.T) {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
    <ID>INV-2026-002</ID>
    <DocumentCurrencyCode>EUR</DocumentCurrencyCode>
    <AccountingSupplierParty>
        <Party>
            <PartyLegalEntity>
                <RegistrationName>Raison Sociale Vendeur</RegistrationName>
            </PartyLegalEntity>
        </Party>
    </AccountingSupplierParty>
    <AccountingCustomerParty>
        <Party>
            <PartyLegalEntity>
                <RegistrationName>Raison Sociale Client</RegistrationName>
            </PartyLegalEntity>
        </Party>
    </AccountingCustomerParty>
</Invoice>`)

		inv, err := Parse(xmlData)
		if err != nil {
			t.Fatalf("Parse() a échoué: %v", err)
		}
		if inv.Seller.Name != "Raison Sociale Vendeur" {
			t.Errorf("attendu Seller.Name 'Raison Sociale Vendeur', obtenu '%s'", inv.Seller.Name)
		}
		if inv.Buyer.Name != "Raison Sociale Client" {
			t.Errorf("attendu Buyer.Name 'Raison Sociale Client', obtenu '%s'", inv.Buyer.Name)
		}
	})

	t.Run("Erreur sur XML malformé", func(t *testing.T) {
		badXML := []byte(`<Invoice><ID>INV-003</ID><DocumentCurrencyCode>EUR`) // Balise non fermée
		_, err := Parse(badXML)
		if err == nil {
			t.Fatal("attendu erreur pour XML malformé, obtenu nil")
		}
	})
}
