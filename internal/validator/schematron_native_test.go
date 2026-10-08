package validator

import (
"testing"
)

func TestNativeSchematronEngine_MissingCustomizationID(t *testing.T) {
xml := []byte(`<?xml version="1.0"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
 xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
 xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>INV-1</cbc:ID>
</Invoice>`)

eng := NewNativeSchematronEngine()
rep, err := eng.ValidateProfile(xml, ProfileEN16931)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if rep.Valid {
t.Fatal("expected invalid (missing CustomizationID / seller tax)")
}
found := false
for _, iss := range rep.Issues {
if iss.RuleID == "BR-01" {
found = true
}
}
if !found {
t.Fatalf("expected BR-01, got %+v", rep.Issues)
}
}

func TestNativeSchematronEngine_CIUSFR_RequiresBuyerSIRET(t *testing.T) {
xml := []byte(`<?xml version="1.0"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
 xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
 xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:CustomizationID>urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0</cbc:CustomizationID>
  <cbc:ID>INV-2</cbc:ID>
  <cac:AccountingSupplierParty>
    <cac:Party>
      <cac:PartyIdentification><cbc:ID>12345678901234</cbc:ID></cac:PartyIdentification>
      <cac:PartyTaxScheme><cbc:CompanyID>FR12345678901</cbc:CompanyID></cac:PartyTaxScheme>
    </cac:Party>
  </cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty>
    <cac:Party></cac:Party>
  </cac:AccountingCustomerParty>
</Invoice>`)

eng := NewNativeSchematronEngine()
rep, err := eng.ValidateProfile(xml, ProfileCIUSFR)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
found := false
for _, iss := range rep.Issues {
if iss.RuleID == "BR-FR-03" {
found = true
}
}
if !found {
t.Fatalf("expected BR-FR-03 for missing buyer SIRET, got %+v", rep.Issues)
}
}
