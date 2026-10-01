package validator

import (
	"os"
	"strings"
	"testing"
)

func validate(s string) FileValidationResult {
	return NewFileValidator().ValidateFileContent([]byte(s))
}

func hasError(r FileValidationResult, part string) bool {
	for _, e := range r.Errors {
		if strings.Contains(e, part) {
			return true
		}
	}
	return false
}

func TestEmptyAndUnknown(t *testing.T) {
	if r := validate(""); r.IsValid || r.Format != FormatUnknown {
		t.Error("un fichier vide doit être invalide")
	}
	if r := validate("   \n\t "); r.IsValid {
		t.Error("un fichier d'espaces doit être invalide")
	}
	if r := validate("juste du texte"); r.IsValid || r.Format != FormatUnknown {
		t.Error("un texte libre doit être inconnu")
	}
}

const ublOK = `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>F-2026-001</cbc:ID>
  <cbc:IssueDate>2026-09-24</cbc:IssueDate>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Vendeur SA</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Acheteur SARL</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
</Invoice>`

func TestUBL(t *testing.T) {
	r := validate(ublOK)
	if !r.IsValid || r.Format != FormatUBL {
		t.Fatalf("UBL valide refusé : %+v", r)
	}
	if r.Meta["invoice_number"] != "F-2026-001" {
		t.Errorf("numéro non extrait : %v", r.Meta)
	}

	noID := strings.Replace(ublOK, "<cbc:ID>F-2026-001</cbc:ID>", "", 1)
	if r := validate(noID); r.IsValid || !hasError(r, "Numéro de facture") {
		t.Errorf("UBL sans numéro accepté : %+v", r)
	}

	// document tronqué après la racine : l'ancien validateur l'acceptait
	broken := strings.Replace(ublOK, "</Invoice>", "", 1)
	if r := validate(broken); r.IsValid {
		t.Error("UBL tronqué accepté")
	}
}

// Peppol BIS / EN 16931 : la raison sociale est dans PartyLegalEntity/RegistrationName,
// PartyName est facultatif. Un tel document est valide.
func TestUBLLegalEntityOnly(t *testing.T) {
	doc := strings.NewReplacer(
		"<cac:PartyName><cbc:Name>Vendeur SA</cbc:Name></cac:PartyName>",
		"<cac:PartyLegalEntity><cbc:RegistrationName>Vendeur SA</cbc:RegistrationName></cac:PartyLegalEntity>",
		"<cac:PartyName><cbc:Name>Acheteur SARL</cbc:Name></cac:PartyName>",
		"<cac:PartyLegalEntity><cbc:RegistrationName>Acheteur SARL</cbc:RegistrationName></cac:PartyLegalEntity>",
	).Replace(ublOK)
	if r := validate(doc); !r.IsValid || r.Format != FormatUBL {
		t.Fatalf("UBL avec PartyLegalEntity seul refusé : %+v", r)
	}
}

func TestXMLRejections(t *testing.T) {
	if r := validate(`<?xml version="1.0"?><!DOCTYPE Invoice><Invoice/>`); r.IsValid || !hasError(r, "DOCTYPE") {
		t.Errorf("DOCTYPE accepté : %+v", r)
	}
	if r := validate(`<racine><a></racine>`); r.IsValid {
		t.Error("XML mal formé accepté")
	}
	if r := validate(`<?xml version="1.0"?><Autre/>`); r.IsValid || r.Format != FormatUnknown {
		t.Errorf("schéma inconnu accepté : %+v", r)
	}
}

func TestCII(t *testing.T) {
	ok := `<?xml version="1.0"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
  xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
  <rsm:ExchangedDocument><ram:ID>FX-1</ram:ID></rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:SellerTradeParty><ram:Name>Vendeur</ram:Name></ram:SellerTradeParty>
      <ram:BuyerTradeParty><ram:Name>Acheteur</ram:Name></ram:BuyerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeSettlement><ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode></ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`
	if r := validate(ok); !r.IsValid || r.Format != FormatFacturX || len(r.Warnings) != 0 {
		t.Fatalf("CII valide refusé : %+v", r)
	}

	noID := strings.Replace(ok, "<ram:ID>FX-1</ram:ID>", "", 1)
	if r := validate(noID); r.IsValid || !hasError(r, "ExchangedDocument/ID") {
		t.Errorf("CII sans numéro accepté : %+v", r)
	}
}

const ediOK = "UNA:+.? 'UNB+UNOC:3+SENDER+RECEIVER+260924:1200+1'UNH+1+INVOIC:D:96A:UN'BGM+380+INV-001+9'DTM+137:20260924:102'UNT+4+1'UNZ+1+1'"

func TestEDIFACT(t *testing.T) {
	r := validate(ediOK)
	if !r.IsValid || r.Format != FormatEDIFACT {
		t.Fatalf("EDIFACT valide refusé : %+v", r)
	}
	if r.Meta["invoice_number"] != "INV-001" {
		t.Errorf("numéro non extrait : %v", r.Meta)
	}

	badCount := strings.Replace(ediOK, "UNT+4+1", "UNT+9+1", 1)
	if r := validate(badCount); r.IsValid || !hasError(r, "UNT") {
		t.Errorf("mauvais compteur UNT accepté : %+v", r)
	}
	noBGM := strings.Replace(ediOK, "BGM+380+INV-001+9'", "", 1)
	if r := validate(noBGM); r.IsValid {
		t.Error("EDIFACT sans BGM accepté")
	}
	noUNZ := strings.Replace(ediOK, "UNZ+1+1'", "", 1)
	if r := validate(noUNZ); r.IsValid || !hasError(r, "UNZ") {
		t.Errorf("EDIFACT sans UNZ accepté : %+v", r)
	}
	// "UNH" et "UNT" cités dans une donnée ne doivent pas suffire
	fake := "UNB+UNOC:3+A+B+260924:1200+1'FOO+UNH UNT'"
	if r := validate(fake); r.IsValid {
		t.Error("un EDIFACT factice contenant seulement les mots UNH/UNT a été accepté")
	}
	if r := validate(strings.TrimSuffix(ediOK, "'")); r.IsValid {
		t.Error("un dernier segment non terminé doit être refusé")
	}
}

func TestPDF(t *testing.T) {
	plain := "%PDF-1.4\n1 0 obj\n<< /Type /Page /Contents 2 0 R >>\nendobj\ntrailer\n%%EOF\n"
	r := validate(plain)
	if !r.IsValid || r.Format != FormatPDF {
		t.Fatalf("PDF simple : %+v", r)
	}
	if r.Meta["is_signed"] != "false" {
		t.Error("un PDF avec /Contents ne doit pas être considéré comme signé")
	}
	if len(r.Warnings) != 2 {
		t.Errorf("deux avertissements attendus (non signé, sans XML) : %v", r.Warnings)
	}

	signed := "%PDF-1.7\n<< /Type /Sig /ByteRange [0 10 20 30] /Contents <00> >>\n%%EOF\n"
	if r := validate(signed); r.Format != FormatSignedPDF || r.Meta["is_signed"] != "true" {
		t.Errorf("PDF signé non détecté : %+v", r)
	}

	hybrid := "%PDF-1.7\n<< /Type /Filespec /F (factur-x.xml) >>\n%%EOF\n"
	if r := validate(hybrid); r.Format != FormatFacturX {
		t.Errorf("PDF Factur-X non détecté : %+v", r)
	}

	if r := validate("%PDF-1.4\n1 0 obj\n<< >>"); r.IsValid || !hasError(r, "EOF") {
		t.Errorf("PDF tronqué accepté : %+v", r)
	}
}

// Le PDF d'exemple du dépôt est un PDF 1.4 simple, sans XML embarqué : il ne
// s'agit donc pas d'un vrai Factur-X. Ce test documente ce constat.
func TestSamplePDF(t *testing.T) {
	data, err := os.ReadFile("../../testdata/facture_test_hybride.pdf")
	if err != nil {
		t.Skip("fichier d'exemple absent")
	}
	r := NewFileValidator().ValidateFileContent(data)
	if !r.IsValid {
		t.Fatalf("le PDF d'exemple doit être lisible : %+v", r)
	}
	if r.Format == FormatFacturX {
		t.Log("le PDF d'exemple embarque désormais un XML Factur-X")
	}
}
