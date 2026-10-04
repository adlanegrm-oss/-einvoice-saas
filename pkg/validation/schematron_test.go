package validation

import (
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/syntax"
)

func createGoldenCanonicalInvoice() *canonical.CanonicalInvoice {
	return &canonical.CanonicalInvoice{
		InvoiceNumber:  "INV-2026-FR-001",
		IssueDate:      time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC),
		DueDate:        time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC),
		TypeCode:       "380",
		Currency:       "EUR",
		BuyerReference: "DEP-FINANCE-99",
		ProfileURN:     "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:en16931",
		Seller: canonical.Party{
			Name:           "Cegedim Integration SA",
			LegalEntityID:  "73204903600045",
			TaxID:          "FR12732049036",
			CountryCode:    "FR",
			AddressLine1:   "137 Rue d'Aguesseau",
			PostalCode:     "92100",
			CityName:       "Boulogne-Billancourt",
			EndpointScheme: "0002",
			ElectronicAddr: "73204903600045",
		},
		Buyer: canonical.Party{
			Name:           "Airwaves Client SAS",
			LegalEntityID:  "80214589000012",
			TaxID:          "FR89802145890",
			CountryCode:    "FR",
			AddressLine1:   "10 Boulevard Haussmann",
			PostalCode:     "75009",
			CityName:       "Paris",
			EndpointScheme: "0002",
			ElectronicAddr: "80214589000012",
		},
		PaymentMeans: canonical.PaymentMeans{
			TypeCode: "30",
			IBAN:     "FR7630006000011234567890189",
		},
		Lines: []canonical.InvoiceLine{
			{
				ID:               "1",
				Name:             "Prestation EDI Chorus Pro",
				Quantity:         1.0,
				UnitCode:         "C62",
				UnitPriceNet:     100000,
				LineExtensionNet: 100000,
				TaxCategory:      "S",
				TaxRatePercent:   20.0,
			},
			{
				ID:               "2",
				Name:             "Support technique B2B",
				Quantity:         2.0,
				UnitCode:         "HUR",
				UnitPriceNet:     25000,
				LineExtensionNet: 50000,
				TaxCategory:      "S",
				TaxRatePercent:   20.0,
			},
		},
		TaxBreakdowns: []canonical.TaxBreakdown{
			{
				TaxCategory:    "S",
				TaxRatePercent: 20.0,
				BaseHT:         150000,
				TaxAmount:      30000,
			},
		},
		MonetaryTotals: canonical.MonetaryTotals{
			LineExtensionTotal: 150000,
			NetHT:              150000,
			TaxAmount:          30000,
			GrossTTC:           180000,
			PayableDue:         180000,
		},
	}
}

func TestSchematron_GoldenValidInvoice(t *testing.T) {
	inv := createGoldenCanonicalInvoice()
	xmlBytes, err := syntax.GenerateCIIXML(inv)
	if err != nil {
		t.Fatalf("Failed to generate CII XML: %v", err)
	}

	engine := NewSchematronEngine(true)
	report, err := engine.ValidateXML(xmlBytes)
	if err != nil {
		t.Fatalf("Engine validation error: %v", err)
	}

	if !report.Valid {
		for _, d := range report.Diagnostics {
			t.Errorf("Unexpected diagnostic: [%s] %s (%s)", d.RuleID, d.Message, d.Path)
		}
		t.Fatalf("Golden invoice should be valid according to EN16931 and CIUS-FR")
	}
}

func TestSchematron_InvalidMathAmounts(t *testing.T) {
	inv := createGoldenCanonicalInvoice()
	inv.MonetaryTotals.GrossTTC = 999999

	xmlBytes, err := syntax.GenerateCIIXML(inv)
	if err != nil {
		t.Fatalf("Failed to generate XML: %v", err)
	}

	engine := NewSchematronEngine(true)
	report, err := engine.ValidateXML(xmlBytes)
	if err != nil {
		t.Fatalf("Engine error: %v", err)
	}

	if report.Valid {
		t.Fatal("Expected validation to fail on corrupt amounts (BR-CO-15), but it passed")
	}

	foundBRCO15 := false
	for _, d := range report.Diagnostics {
		if d.RuleID == "BR-CO-15" {
			foundBRCO15 = true
			if d.Severity != SchematronError {
				t.Errorf("Expected Severity ERROR for BR-CO-15, got %s", d.Severity)
			}
		}
	}

	if !foundBRCO15 {
		t.Errorf("Rule BR-CO-15 was not triggered in diagnostics: %+v", report.Diagnostics)
	}
}

func TestSchematron_InvalidDueDate(t *testing.T) {
	inv := createGoldenCanonicalInvoice()
	inv.DueDate = inv.IssueDate.AddDate(0, 0, -5)

	xmlBytes, err := syntax.GenerateCIIXML(inv)
	if err != nil {
		t.Fatalf("Failed to generate XML: %v", err)
	}

	engine := NewSchematronEngine(true)
	report, _ := engine.ValidateXML(xmlBytes)

	if report.Valid {
		t.Fatal("Expected validation to fail on past due date, but passed")
	}

	foundRule := false
	for _, d := range report.Diagnostics {
		if d.RuleID == "BR-CO-25" {
			foundRule = true
		}
	}
	if !foundRule {
		t.Errorf("Rule BR-CO-25 was expected, got %+v", report.Diagnostics)
	}
}

func TestSchematron_InvalidFrenchSIRET(t *testing.T) {
	inv := createGoldenCanonicalInvoice()
	inv.Seller.LegalEntityID = "INVALID_SIRET_ABC"

	xmlBytes, err := syntax.GenerateCIIXML(inv)
	if err != nil {
		t.Fatalf("Failed to generate XML: %v", err)
	}

	engine := NewSchematronEngine(true)
	report, _ := engine.ValidateXML(xmlBytes)

	if report.Valid {
		t.Fatal("Expected validation to fail on invalid French SIRET (CIUS-FR-01)")
	}

	foundCIUS := false
	for _, d := range report.Diagnostics {
		if d.RuleID == "CIUS-FR-01" {
			foundCIUS = true
		}
	}
	if !foundCIUS {
		t.Errorf("Expected CIUS-FR-01 error, got %+v", report.Diagnostics)
	}
}
