package validator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSchematronSidecarClient_ParseSVRL(t *testing.T) {
	svrlMock := `<schematron-output xmlns:svrl="http://purl.oclc.org/dsdl/svrl">
		<failed-assert id="BR-CO-15" flag="fatal" location="/rsm:CrossIndustryInvoice/Totals">
			<text>Le montant total TTC doit égaler Total HT + Taxe.</text>
		</failed-assert>
		<failed-assert id="CIUS-FR-01" flag="fatal" location="/rsm:CrossIndustryInvoice/Seller/ID">
			<text>SIREN ou SIRET obligatoire pour le vendeur français.</text>
		</failed-assert>
	</schematron-output>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(svrlMock))
	}))
	defer ts.Close()

	client := NewSchematronSidecarClient(ts.URL)
	errReport, err := client.Validate(context.Background(), "EN16931", []byte("<test/>"))
	if err != nil {
		t.Fatalf("Unexpected client error: %v", err)
	}
	if errReport == nil {
		t.Fatal("Expected validation violations, got nil")
	}
	if len(errReport.Violations) != 2 {
		t.Fatalf("Expected 2 violations, got %d", len(errReport.Violations))
	}
	if errReport.Violations[0].RuleID != "BR-CO-15" {
		t.Errorf("Expected BR-CO-15, got %s", errReport.Violations[0].RuleID)
	}
	if errReport.Violations[1].RuleID != "CIUS-FR-01" {
		t.Errorf("Expected CIUS-FR-01, got %s", errReport.Violations[1].RuleID)
	}
}
