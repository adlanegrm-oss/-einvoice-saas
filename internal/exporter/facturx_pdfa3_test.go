package exporter

import (
"bytes"
"testing"
"time"
)

func TestGenerateFacturXPDFA3_RejectsInvalidOrNonCIIXML(t *testing.T) {
meta := InvoiceMetadata{
InvoiceNumber: "INV-001",
Profile:       ProfileEN16931,
}

// 1. XML vide
_, err := GenerateFacturXPDFA3(meta, []byte(""))
if err == nil {
t.Fatal("attendu : rejet d'un XML vide")
}

// 2. XML mal formé
_, err = GenerateFacturXPDFA3(meta, []byte("<rsm:CrossIndustryInvoice><bad>"))
if err == nil {
t.Fatal("attendu : rejet d'un XML mal formé")
}

// 3. XML valide mais racine UBL au lieu de CII
_, err = GenerateFacturXPDFA3(meta, []byte(`<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><ID>INV-1</ID></Invoice>`))
if err == nil {
t.Fatal("attendu : rejet d'une racine non-CII")
}
}

func TestGenerateFacturXPDFA3_EmbedsExactXML(t *testing.T) {
meta := InvoiceMetadata{
InvoiceNumber: "FX-2026-001",
SellerName:    "ACME CORP",
IssueDate:     time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
Currency:      "EUR",
Profile:       ProfileEN16931,
}

cii := []byte(`<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"><rsm:ExchangedDocument><ram:ID xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">FX-2026-001</ram:ID></rsm:ExchangedDocument></rsm:CrossIndustryInvoice>`)

pdf, err := GenerateFacturXPDFA3(meta, cii)
if err != nil {
t.Fatalf("erreur génération : %v", err)
}

if err := VerifyFacturXContainer(pdf); err != nil {
t.Fatalf("échec vérification conteneur : %v", err)
}

// Vérification de compatibilité de l'ancien alias
if err := VerifyPDFA3Conformance(pdf); err != nil {
t.Fatalf("échec VerifyPDFA3Conformance : %v", err)
}

if !bytes.Contains(pdf, cii) {
t.Fatal("le flux XML injecté dans le PDF ne correspond pas à l'entrée fournie")
}
}
