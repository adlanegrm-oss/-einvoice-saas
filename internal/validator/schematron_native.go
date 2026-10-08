package validator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// NativeRule représente une règle métier exécutable sans XSLT externe.
type NativeRule struct {
	ID       string
	Flag     Severity // FATAL / ERROR / WARNING
	Message  string
	Applies  func(ctx *NativeValidationContext) bool // filtre profil / syntaxe
	Evaluate func(ctx *NativeValidationContext) (ok bool, location string)
}

// NativeValidationContext agrège les faits extraits du XML.
type NativeValidationContext struct {
	Profile     ValidationProfile
	IsUBL       bool
	IsCII       bool
	CustomID    string
	SellerTaxID string
	SellerSiret string
	BuyerSiret  string
	TaxItems    []taxBreakdownItem
	RawXML      []byte
}

// NativeSchematronEngine exécute un jeu de règles Go (pas de dépendance xsltproc).
type NativeSchematronEngine struct {
	rules []NativeRule
}

func NewNativeSchematronEngine() *NativeSchematronEngine {
	e := &NativeSchematronEngine{}
	e.rules = defaultNativeRules()
	return e
}

func defaultNativeRules() []NativeRule {
	siretRe := regexp.MustCompile(`^\d{14}$`)

	return []NativeRule{
		{
			ID:      "BR-01",
			Flag:    SeverityFatal,
			Message: "[BR-01]-An Invoice that claims compliance with EN16931 must specify the specification identifier (CustomizationID).",
			Applies: func(ctx *NativeValidationContext) bool { return true },
			Evaluate: func(ctx *NativeValidationContext) (bool, string) {
				return strings.TrimSpace(ctx.CustomID) != "", "/*[1]"
			},
		},
		{
			ID:      "BR-CO-09",
			Flag:    SeverityFatal,
			Message: "[BR-CO-09]-The Seller VAT identifier (BT-31), the Seller tax representative VAT identifier (BT-63) or the Seller tax registration identifier (BT-32) shall be present.",
			Applies: func(ctx *NativeValidationContext) bool { return true },
			Evaluate: func(ctx *NativeValidationContext) (bool, string) {
				return strings.TrimSpace(ctx.SellerTaxID) != "" || siretRe.MatchString(ctx.SellerSiret), "/*[1]"
			},
		},
		{
			ID:      "BR-FR-01",
			Flag:    SeverityFatal,
			Message: "[BR-FR-01]-The Seller identifier (BT-29) shall be present and must be a SIRET number under CIUS-FR.",
			Applies: func(ctx *NativeValidationContext) bool {
				return ctx.Profile == ProfileCIUSFR
			},
			Evaluate: func(ctx *NativeValidationContext) (bool, string) {
				return siretRe.MatchString(strings.TrimSpace(ctx.SellerSiret)), "/*[1]"
			},
		},
		{
			ID:      "BR-FR-03",
			Flag:    SeverityFatal,
			Message: "[BR-FR-03]-The Buyer identifier (BT-46) shall be present and must be a SIRET number under CIUS-FR.",
			Applies: func(ctx *NativeValidationContext) bool {
				return ctx.Profile == ProfileCIUSFR
			},
			Evaluate: func(ctx *NativeValidationContext) (bool, string) {
				return siretRe.MatchString(strings.TrimSpace(ctx.BuyerSiret)), "/*[1]"
			},
		},
		{
			ID:      "BR-CO-17",
			Flag:    SeverityFatal,
			Message: "[BR-CO-17]-VAT category tax amount must equal taxable amount × rate (tolerance 0.02).",
			Applies: func(ctx *NativeValidationContext) bool { return true },
			Evaluate: func(ctx *NativeValidationContext) (bool, string) {
				for _, item := range ctx.TaxItems {
					expected := math.Round(item.Basis*(item.Rate/100.0)*100) / 100
					if math.Abs(item.Tax-expected) > 0.02 {
						return false, item.Xpath
					}
				}
				return true, ""
			},
		},
	}
}

// ValidateProfile exécute les règles natives pour un profil donné.
func (e *NativeSchematronEngine) ValidateProfile(xmlData []byte, profile ValidationProfile) (*SchematronReport, error) {
	ctx, err := extractNativeContext(xmlData, profile)
	if err != nil {
		return nil, err
	}

	report := &SchematronReport{
		Profile: profile,
		Valid:   true,
		Issues:  make([]SchematronIssue, 0),
	}

	for _, rule := range e.rules {
		if !rule.Applies(ctx) {
			continue
		}
		ok, loc := rule.Evaluate(ctx)
		if ok {
			continue
		}
		if loc == "" {
			loc = "/*[1]"
		}
		report.Valid = false
		report.Issues = append(report.Issues, SchematronIssue{
			RuleID:   rule.ID,
			Severity: rule.Flag,
			Message:  rule.Message,
			XPath:    loc,
		})
	}

	return report, nil
}

func extractNativeContext(xmlData []byte, profile ValidationProfile) (*NativeValidationContext, error) {
	ctx := &NativeValidationContext{Profile: profile, RawXML: xmlData}

	var ubl ublInvoiceDoc
	if err := xml.Unmarshal(xmlData, &ubl); err == nil && ubl.XMLName.Local == "Invoice" {
		ctx.IsUBL = true
		ctx.CustomID = strings.TrimSpace(ubl.CustomizationID.Value)
		for _, pi := range ubl.AccountingSupplierParty.Party.PartyIdentification {
			if v := strings.TrimSpace(pi.Value); v != "" {
				ctx.SellerSiret = v
			}
		}
		for _, pi := range ubl.AccountingCustomerParty.Party.PartyIdentification {
			if v := strings.TrimSpace(pi.Value); v != "" {
				ctx.BuyerSiret = v
			}
		}
		for _, pts := range ubl.AccountingSupplierParty.Party.PartyTaxScheme {
			if v := strings.TrimSpace(pts.CompanyID.Value); v != "" {
				ctx.SellerTaxID = v
				break
			}
		}
		for i, tt := range ubl.TaxTotal {
			for j, sub := range tt.TaxSubtotal {
				basis, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxableAmount.Value), 64)
				tax, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxAmount.Value), 64)
				rate, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxCategory.Percent.Value), 64)
				loc := fmt.Sprintf("/*[local-name()='Invoice']/*[local-name()='TaxTotal'][%d]/*[local-name()='TaxSubtotal'][%d]", i+1, j+1)
				ctx.TaxItems = append(ctx.TaxItems, taxBreakdownItem{Basis: basis, Rate: rate, Tax: tax, Xpath: loc})
			}
		}
		// SIRET 14 chiffres via regex (namespaces variables)
		if m := regexp.MustCompile(`(?s)<cac:AccountingSupplierParty>.*?</cac:AccountingSupplierParty>`).Find(xmlData); m != nil {
			if id := regexp.MustCompile(`>(\d{14})<`).FindSubmatch(m); len(id) > 1 {
				ctx.SellerSiret = string(id[1])
			}
		}
		if m := regexp.MustCompile(`(?s)<cac:AccountingCustomerParty>.*?</cac:AccountingCustomerParty>`).Find(xmlData); m != nil {
			if id := regexp.MustCompile(`>(\d{14})<`).FindSubmatch(m); len(id) > 1 {
				ctx.BuyerSiret = string(id[1])
			}
		}
		return ctx, nil
	}

	var cii ciiInvoiceDoc
	if err := xml.Unmarshal(xmlData, &cii); err == nil && cii.XMLName.Local == "CrossIndustryInvoice" {
		ctx.IsCII = true
		ctx.CustomID = strings.TrimSpace(cii.ExchangedDocumentContext.GuidelineSpecifiedDocumentContextParameter.ID.Value)
		seller := cii.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.SellerTradeParty
		buyer := cii.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.BuyerTradeParty
		if v := strings.TrimSpace(seller.SpecifiedLegalOrganization.ID.Value); len(v) == 14 {
			ctx.SellerSiret = v
		}
		if v := strings.TrimSpace(buyer.SpecifiedLegalOrganization.ID.Value); len(v) == 14 {
			ctx.BuyerSiret = v
		}
		for _, reg := range seller.SpecifiedTaxRegistration {
			val := strings.TrimSpace(reg.ID.Value)
			if reg.ID.SchemeID == "VA" || strings.HasPrefix(val, "FR") {
				ctx.SellerTaxID = val
			}
			if len(val) == 14 {
				ctx.SellerSiret = val
			}
		}
		for i, taxItem := range cii.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.ApplicableTradeTax {
			basis, _ := strconv.ParseFloat(strings.TrimSpace(taxItem.BasisAmount.Value), 64)
			tax, _ := strconv.ParseFloat(strings.TrimSpace(taxItem.CalculatedAmount.Value), 64)
			rate, _ := strconv.ParseFloat(strings.TrimSpace(taxItem.RateApplicablePercent.Value), 64)
			loc := fmt.Sprintf("/*[local-name()='CrossIndustryInvoice']//*[local-name()='ApplicableTradeTax'][%d]", i+1)
			ctx.TaxItems = append(ctx.TaxItems, taxBreakdownItem{Basis: basis, Rate: rate, Tax: tax, Xpath: loc})
		}
		return ctx, nil
	}

	if bytes.Contains(xmlData, []byte("<")) {
		return nil, fmt.Errorf("native schematron: unsupported or unparseable XML root")
	}
	return nil, fmt.Errorf("native schematron: empty or invalid XML")
}
