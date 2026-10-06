package validator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type XSLTExecutor interface {
	Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error)
}

type CLIExecutor struct {
	BinaryPath string
}

func NewCLIExecutor(binPath string) *CLIExecutor {
	if binPath == "" {
		binPath = "xsltproc"
	}
	return &CLIExecutor{BinaryPath: binPath}
}

func (c *CLIExecutor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
	cmd := exec.Command(c.BinaryPath, "-", "-")
	cmd.Stdin = bytes.NewReader(append(xsltSheet, xmlInput...))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cli xslt: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

type MockXSLTExecutor struct {
	MockOutput []byte
	MockErr    error
}

func (m *MockXSLTExecutor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
	if m.MockErr != nil {
		return nil, m.MockErr
	}
	if len(m.MockOutput) > 0 {
		return m.MockOutput, nil
	}
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<schematron-output xmlns="http://purl.oclc.org/dsdl/svrl" title="Validation EN 16931">
</schematron-output>`), nil
}

type NativeEN16931Executor struct{}

func NewDefaultXSLTExecutor() XSLTExecutor {
	return &NativeEN16931Executor{}
}

type ublPartyID struct {
	SchemeID string `xml:"schemeID,attr"`
	Value    string `xml:",chardata"`
}

type ublInvoiceDoc struct {
	XMLName         xml.Name `xml:"Invoice"`
	CustomizationID struct {
		Value string `xml:",chardata"`
	} `xml:"CustomizationID"`
	AccountingSupplierParty struct {
		Party struct {
			PartyIdentification []ublPartyID `xml:"PartyIdentification>ID"`
			PartyTaxScheme      []struct {
				CompanyID struct {
					Value string `xml:",chardata"`
				} `xml:"CompanyID"`
			} `xml:"PartyTaxScheme"`
		} `xml:"Party"`
	} `xml:"AccountingSupplierParty"`
	TaxTotal []struct {
		TaxSubtotal []struct {
			TaxableAmount struct {
				Value string `xml:",chardata"`
			} `xml:"TaxableAmount"`
			TaxAmount struct {
				Value string `xml:",chardata"`
			} `xml:"TaxAmount"`
			TaxCategory struct {
				Percent struct {
					Value string `xml:",chardata"`
				} `xml:"Percent"`
			} `xml:"TaxCategory"`
		} `xml:"TaxSubtotal"`
	} `xml:"TaxTotal"`
}

type ciiInvoiceDoc struct {
	XMLName                  xml.Name `xml:"CrossIndustryInvoice"`
	ExchangedDocumentContext struct {
		GuidelineSpecifiedDocumentContextParameter struct {
			ID struct {
				Value string `xml:",chardata"`
			} `xml:"ID"`
		} `xml:"GuidelineSpecifiedDocumentContextParameter"`
	} `xml:"ExchangedDocumentContext"`
	SupplyChainTradeTransaction struct {
		ApplicableHeaderTradeAgreement struct {
			SellerTradeParty struct {
				SpecifiedLegalOrganization struct {
					ID struct {
						SchemeID string `xml:"schemeID,attr"`
						Value    string `xml:",chardata"`
					} `xml:"ID"`
				} `xml:"SpecifiedLegalOrganization"`
				SpecifiedTaxRegistration []struct {
					ID struct {
						SchemeID string `xml:"schemeID,attr"`
						Value    string `xml:",chardata"`
					} `xml:"ID"`
				} `xml:"SpecifiedTaxRegistration"`
			} `xml:"SellerTradeParty"`
		} `xml:"ApplicableHeaderTradeAgreement"`
		ApplicableHeaderTradeSettlement struct {
			ApplicableTradeTax []struct {
				CalculatedAmount struct {
					Value string `xml:",chardata"`
				} `xml:"CalculatedAmount"`
				BasisAmount struct {
					Value string `xml:",chardata"`
				} `xml:"BasisAmount"`
				RateApplicablePercent struct {
					Value string `xml:",chardata"`
				} `xml:"RateApplicablePercent"`
			} `xml:"ApplicableTradeTax"`
		} `xml:"ApplicableHeaderTradeSettlement"`
	} `xml:"SupplyChainTradeTransaction"`
}

type taxBreakdownItem struct {
	Basis float64
	Rate  float64
	Tax   float64
	Xpath string
}

func (n *NativeEN16931Executor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
	var failedAsserts []string
	var customID string
	var sellerTaxID string
	var sellerSiret string
	var breakdown []taxBreakdownItem
	var isDocFound bool

	var ublDoc ublInvoiceDoc
	if err := xml.Unmarshal(xmlInput, &ublDoc); err == nil && ublDoc.XMLName.Local == "Invoice" {
		isDocFound = true
		customID = strings.TrimSpace(ublDoc.CustomizationID.Value)

		for _, pi := range ublDoc.AccountingSupplierParty.Party.PartyIdentification {
			val := strings.TrimSpace(pi.Value)
			if val != "" {
				sellerSiret = val
			}
		}
		for _, pts := range ublDoc.AccountingSupplierParty.Party.PartyTaxScheme {
			if val := strings.TrimSpace(pts.CompanyID.Value); val != "" {
				sellerTaxID = val
				break
			}
		}

		for i, tt := range ublDoc.TaxTotal {
			for j, sub := range tt.TaxSubtotal {
				basis, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxableAmount.Value), 64)
				tax, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxAmount.Value), 64)
				rate, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxCategory.Percent.Value), 64)
				loc := fmt.Sprintf("/*[local-name()='Invoice'][1]/*[local-name()='TaxTotal'][%d]/*[local-name()='TaxSubtotal'][%d]", i+1, j+1)
				breakdown = append(breakdown, taxBreakdownItem{Basis: basis, Rate: rate, Tax: tax, Xpath: loc})
			}
		}
	} else {
		var ciiDoc ciiInvoiceDoc
		if errCII := xml.Unmarshal(xmlInput, &ciiDoc); errCII == nil && ciiDoc.XMLName.Local == "CrossIndustryInvoice" {
			isDocFound = true
			customID = strings.TrimSpace(ciiDoc.ExchangedDocumentContext.GuidelineSpecifiedDocumentContextParameter.ID.Value)
			sellerParty := ciiDoc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.SellerTradeParty
			if sVal := strings.TrimSpace(sellerParty.SpecifiedLegalOrganization.ID.Value); len(sVal) == 14 {
				sellerSiret = sVal
			}
			for _, reg := range sellerParty.SpecifiedTaxRegistration {
				val := strings.TrimSpace(reg.ID.Value)
				if reg.ID.SchemeID == "0009" || len(val) == 14 {
					sellerSiret = val
				}
				if reg.ID.SchemeID == "VA" || strings.HasPrefix(val, "FR") {
					sellerTaxID = val
				}
			}

			for i, taxItem := range ciiDoc.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.ApplicableTradeTax {
				basis, _ := strconv.ParseFloat(strings.TrimSpace(taxItem.BasisAmount.Value), 64)
				tax, _ := strconv.ParseFloat(strings.TrimSpace(taxItem.CalculatedAmount.Value), 64)
				rate, _ := strconv.ParseFloat(strings.TrimSpace(taxItem.RateApplicablePercent.Value), 64)
				loc := fmt.Sprintf("/*[local-name()='CrossIndustryInvoice'][1]/*[local-name()='SupplyChainTradeTransaction'][1]/*[local-name()='ApplicableHeaderTradeSettlement'][1]/*[local-name()='ApplicableTradeTax'][%d]", i+1)
				breakdown = append(breakdown, taxBreakdownItem{Basis: basis, Rate: rate, Tax: tax, Xpath: loc})
			}
		}
	}

	if !isDocFound {
		return nil, fmt.Errorf("native validator: unsupported or unparseable XML document")
	}

	// Détection du SIRET dans le bloc fournisseur UBL
	supplierBlockRegex := regexp.MustCompile(`(?s)<cac:AccountingSupplierParty>.*?</cac:AccountingSupplierParty>`)
	if match := supplierBlockRegex.Find(xmlInput); match != nil {
		idRegex := regexp.MustCompile(`<cac:PartyIdentification>\s*<cbc:ID[^>]*>(\d{14})</cbc:ID>\s*</cac:PartyIdentification>`)
		if idMatch := idRegex.FindSubmatch(match); len(idMatch) > 1 {
			sellerSiret = string(idMatch[1])
		} else {
			sellerSiret = ""
		}
	}

	// [BR-01]
	if customID == "" {
		failedAsserts = append(failedAsserts, `<failed-assert id="BR-01" location="/*[1]" flag="fatal"><text>[BR-01]-An Invoice that claims compliance with EN16931 must specify the specification identifier (CustomizationID).</text></failed-assert>`)
	}

	// [BR-CO-09]
	if strings.TrimSpace(sellerTaxID) == "" {
		failedAsserts = append(failedAsserts, `<failed-assert id="BR-CO-09" location="/*[1]" flag="fatal"><text>[BR-CO-09]-The Seller VAT identifier (BT-31), the Seller tax representative VAT identifier (BT-63) or the Seller tax registration identifier (BT-32) shall be present.</text></failed-assert>`)
	}

	// [BR-FR-01] Rejet si absence de SIRET vendeur
	if strings.TrimSpace(sellerSiret) == "" {
		failedAsserts = append(failedAsserts, `<failed-assert id="BR-FR-01" location="/*[1]" flag="fatal"><text>[BR-FR-01]-The Seller identifier (BT-29) shall be present and must be a SIRET number under CIUS-FR.</text></failed-assert>`)
	}

	// [BR-CO-17]
	for _, item := range breakdown {
		expected := math.Round(item.Basis*(item.Rate/100.0)*100) / 100
		diff := math.Abs(item.Tax - expected)
		if diff > 0.02 {
			msg := fmt.Sprintf(`<failed-assert id="BR-CO-17" location="%s" flag="fatal"><text>[BR-CO-17]-The VAT category tax amount (%.2f) in a VAT breakdown must equal the VAT category taxable amount (%.2f) multiplied by the VAT category rate (%.2f%%).</text></failed-assert>`,
				item.Xpath, item.Tax, item.Basis, item.Rate)
			failedAsserts = append(failedAsserts, msg)
		}
	}

	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<schematron-output xmlns=\"http://purl.oclc.org/dsdl/svrl\" title=\"Validation EN 16931\">\n")
	for _, fa := range failedAsserts {
		sb.WriteString("  " + fa + "\n")
	}
	sb.WriteString("</schematron-output>")

	return []byte(sb.String()), nil
}