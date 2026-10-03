package en16931_test

import (
	"context"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
)

func TestParseSVRL_WithFailures(t *testing.T) {
	mockSVRL := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svrl:schematron-output xmlns:svrl="http://purl.oclc.org/dsdl/svrl" title="EN16931 validation">
    <svrl:fired-rule context="ubl:Invoice"/>
    <svrl:failed-assert id="BR-01" flag="fatal" location="/*:Invoice[1]/*:CustomizationID[1]">
        <svrl:text>[BR-01]-An Invoice that contains specification identifier shall be compliant.</svrl:text>
    </svrl:failed-assert>
    <svrl:failed-assert id="BR-CO-10" flag="warning" location="/*:Invoice[1]/*:LegalMonetaryTotal[1]">
        <svrl:text>[BR-CO-10]-Line extension amount warning.</svrl:text>
    </svrl:failed-assert>
</svrl:schematron-output>`)

	res, err := en16931.ParseSVRL(mockSVRL)
	if err != nil {
		t.Fatalf("échec inattendu ParseSVRL: %v", err)
	}
	if res.IsValid {
		t.Errorf("attendu IsValid=false, obtenu true")
	}
	if len(res.Diagnostics) != 2 {
		t.Fatalf("attendu 2 diagnostics, obtenu %d", len(res.Diagnostics))
	}
	if res.Diagnostics[0].RuleID != "BR-01" || res.Diagnostics[0].Severity != "FATAL" {
		t.Errorf("diagnostic 0 invalide: %+v", res.Diagnostics[0])
	}
	if res.Diagnostics[1].Severity != "WARNING" {
		t.Errorf("diagnostic 1 sévérité attendue WARNING, obtenu %s", res.Diagnostics[1].Severity)
	}
}

func TestQuickValidateProfile_Success(t *testing.T) {
	v := en16931.NewSchematronValidator()
	validXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
    <cbc:CustomizationID>urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0</cbc:CustomizationID>
    <cbc:ID>INV-2026-0001</cbc:ID>
    <cbc:IssueDate>2026-10-02</cbc:IssueDate>
    <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
</Invoice>`)

	res, err := v.QuickValidateProfile(context.Background(), validXML)
	if err != nil {
		t.Fatalf("validation inattendue en échec: %v", err)
	}
	if !res.IsValid {
		t.Errorf("attendu IsValid=true, obtenu false")
	}
	if len(res.Diagnostics) != 0 {
		t.Errorf("attendu 0 diagnostics, obtenu %d", len(res.Diagnostics))
	}
}

func TestQuickValidateProfile_MissingMandatoryFields(t *testing.T) {
	v := en16931.NewSchematronValidator()
	invalidXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
    <cbc:ID></cbc:ID>
</Invoice>`)

	res, err := v.QuickValidateProfile(context.Background(), invalidXML)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if res.IsValid {
		t.Errorf("attendu IsValid=false, obtenu true")
	}
	if len(res.Diagnostics) == 0 {
		t.Errorf("attendu diagnostics non vides")
	}
}
