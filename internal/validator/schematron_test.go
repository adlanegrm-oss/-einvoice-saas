package validator

import (
	"testing"
)

func TestParseSVRL_ValidDocument(t *testing.T) {
	svrl := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svrl:schematron-output xmlns:svrl="http://purl.oclc.org/dsdl/svrl" title="Validation EN 16931">
<svrl:active-pattern id="en16931-pattern" name="EN 16931 Core Rules"/>
</svrl:schematron-output>`)

	report, err := ParseSVRL(svrl)
	if err != nil {
		t.Fatalf("erreur inattendue sur SVRL valide: %v", err)
	}
	if !report.Valid {
		t.Errorf("attendu: valide, reçu: non valide avec %d issues", len(report.Issues))
	}
	if len(report.Issues) != 0 {
		t.Errorf("attendu: 0 issue, reçu: %d", len(report.Issues))
	}
}

func TestParseSVRL_WithFailedAsserts(t *testing.T) {
	svrl := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svrl:schematron-output xmlns:svrl="http://purl.oclc.org/dsdl/svrl">
<svrl:failed-assert id="BR-FR-01" flag="fatal" location="/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction[1]/ram:ApplicableHeaderTradeAgreement[1]/ram:SellerTradeParty[1]">
<svrl:text>[BR-FR-01] Le numéro SIRET du vendeur est obligatoire pour les factures B2B françaises.</svrl:text>
</svrl:failed-assert>
<svrl:failed-assert id="BR-CO-15" flag="error" location="/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction[1]/ram:ApplicableHeaderTradeSettlement[1]/ram:SpecifiedTradeSettlementHeaderMonetarySummation[1]">
<svrl:text>[BR-CO-15] Invoice total amount with VAT (BT-112) must equal the sum of amount without VAT and total VAT amount.</svrl:text>
</svrl:failed-assert>
<svrl:failed-assert id="BR-WARN-01" flag="warning" location="/rsm:CrossIndustryInvoice/rsm:ExchangedDocument[1]">
<svrl:text>Recommandation de libellé de référence.</svrl:text>
</svrl:failed-assert>
</svrl:schematron-output>`)

	report, err := ParseSVRL(svrl)
	if err != nil {
		t.Fatalf("erreur de parsing SVRL: %v", err)
	}

	if report.Valid {
		t.Fatal("attendu: document non valide suite aux asserts fatal et error")
	}

	if len(report.Issues) != 3 {
		t.Fatalf("attendu: 3 issues, obtenu: %d", len(report.Issues))
	}

	// Contrôle de la règle CIUS-FR
	issue1 := report.Issues[0]
	if issue1.RuleID != "BR-FR-01" {
		t.Errorf("RuleID attendu 'BR-FR-01', obtenu '%s'", issue1.RuleID)
	}
	if issue1.Severity != SeverityFatal {
		t.Errorf("Severity attendue FATAL, obtenue %s", issue1.Severity)
	}
	if issue1.XPath == "" {
		t.Error("XPath ne doit pas être vide")
	}

	// Contrôle de la règle EN 16931
	issue2 := report.Issues[1]
	if issue2.RuleID != "BR-CO-15" {
		t.Errorf("RuleID attendu 'BR-CO-15', obtenu '%s'", issue2.RuleID)
	}
	if issue2.Severity != SeverityError {
		t.Errorf("Severity attendue ERROR, obtenue %s", issue2.Severity)
	}
}

func TestSchematronEngine_EndToEnd(t *testing.T) {
	mockSVRL := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svrl:schematron-output xmlns:svrl="http://purl.oclc.org/dsdl/svrl">
<svrl:failed-assert id="BR-CIUS-FR-TVA" flag="error" location="/CrossIndustryInvoice/ApplicableHeaderTradeSettlement">
<svrl:text>Taux de TVA non conforme pour CIUS-FR v2.0</svrl:text>
</svrl:failed-assert>
</svrl:schematron-output>`)

	executor := &MockXSLTExecutor{MockOutput: mockSVRL}
	engine := NewSchematronEngine(executor)

	report, err := engine.ValidateSchematron([]byte("<xml/>"), []byte("<xsl/>"), ProfileCIUSFR)
	if err != nil {
		t.Fatalf("erreur d'exécution Schematron: %v", err)
	}

	if report.Profile != ProfileCIUSFR {
		t.Errorf("profil attendu %s, obtenu %s", ProfileCIUSFR, report.Profile)
	}
	if report.Valid {
		t.Error("attendu: rapport invalide")
	}
	if len(report.Issues) != 1 || report.Issues[0].RuleID != "BR-CIUS-FR-TVA" {
		t.Errorf("problème de détection de l'erreur: %+v", report.Issues)
	}
}
