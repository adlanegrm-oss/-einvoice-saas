package exporter

import (
"testing"
"time"
)

func TestValidateCIIXMLForEmbedding(t *testing.T) {
t.Run("Rejette XML vide", func(t *testing.T) {
if err := ValidateCIIXMLForEmbedding([]byte("   ")); err == nil {
t.Fatal("attendu: rejet sur payload vide")
}
})

t.Run("Rejette XML mal formé", func(t *testing.T) {
if err := ValidateCIIXMLForEmbedding([]byte("<bad><xml>")); err == nil {
t.Fatal("attendu: rejet sur XML mal formé")
}
})

t.Run("Rejette racine non CrossIndustryInvoice", func(t *testing.T) {
xmlDoc := []byte(`<?xml version="1.0"?><Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><ID>1</ID></Invoice>`)
if err := ValidateCIIXMLForEmbedding(xmlDoc); err == nil {
t.Fatal("attendu: rejet racine non-CII")
}
})

t.Run("Accepte racine CrossIndustryInvoice valide", func(t *testing.T) {
xmlDoc := []byte(`<?xml version="1.0"?><rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"><rsm:ExchangedDocument/></rsm:CrossIndustryInvoice>`)
if err := ValidateCIIXMLForEmbedding(xmlDoc); err != nil {
t.Fatalf("échec inattendu sur racine CII valide: %v", err)
}
})
}

func TestGenerateAndVerifyFacturX(t *testing.T) {
validCII := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100">
<rsm:ExchangedDocument>
<ram:ID xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">INV-2026-TEST</ram:ID>
</rsm:ExchangedDocument>
</rsm:CrossIndustryInvoice>`)

meta := InvoiceMetadata{
InvoiceNumber: "INV-2026-TEST",
SellerName:    "Entreprise Test SAS",
BuyerName:     "Client SARL",
IssueDate:     time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
Currency:      "EUR",
TotalHT:       1000.0,
TotalTTC:      1200.0,
Profile:       ProfileEN16931,
}

pdfData, err := GenerateFacturXPDFA3(meta, validCII)
if err != nil {
t.Fatalf("erreur génération: %v", err)
}

if err := VerifyFacturXContainer(pdfData); err != nil {
t.Fatalf("échec vérification conteneur Factur-X: %v", err)
}
}
