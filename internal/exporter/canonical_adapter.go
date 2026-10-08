package exporter

import (
	"time"

	"github.com/shopspring/decimal"

	"einvoice-saas/internal/model"
)

// FromCanonical construit les métadonnées Factur-X à partir du modèle canonique.
func FromCanonical(inv *model.CanonicalInvoice, profile FacturXProfile) InvoiceMetadata {
	if inv == nil {
		return InvoiceMetadata{Profile: profile}
	}
	if profile == "" {
		profile = ProfileEN16931
	}

	issueDate := time.Now().UTC()
	if inv.IssueDate != "" {
		if parsed, err := time.Parse("2006-01-02", inv.IssueDate); err == nil {
			issueDate = parsed
		}
	}

	sellerSIREN := inv.Seller.LegalID
	if len(sellerSIREN) == 14 {
		sellerSIREN = sellerSIREN[:9] // SIRET → SIREN
	}

	totalHT, _ := inv.Totals.TaxExclusiveAmount.Float64()
	totalTTC, _ := inv.Totals.TaxInclusiveAmount.Float64()
	totalTax, _ := inv.Totals.TotalTaxAmount.Float64()
	if inv.Totals.TotalTaxAmount.IsZero() {
		// dérive TVA = TTC - HT si TotalTaxAmount non renseigné
		totalTax, _ = inv.Totals.TaxInclusiveAmount.Sub(inv.Totals.TaxExclusiveAmount).Float64()
	}

	currency := inv.DocumentCurrency
	if currency == "" {
		currency = "EUR"
	}

	return InvoiceMetadata{
		InvoiceNumber: inv.ID,
		SellerName:    inv.Seller.Name,
		SellerSIREN:   sellerSIREN,
		SellerVAT:     inv.Seller.VATID,
		BuyerName:     inv.Buyer.Name,
		IssueDate:     issueDate,
		Currency:      currency,
		TotalHT:       totalHT,
		TotalTTC:      totalTTC,
		TotalTax:      totalTax,
		Profile:       profile,
	}
}

// GenerateFromCanonical produit un PDF/A-3 Factur-X à partir du canonique + XML CII embarqué.
func GenerateFromCanonical(inv *model.CanonicalInvoice, ciiXML []byte, profile FacturXProfile) ([]byte, error) {
	meta := FromCanonical(inv, profile)
	return GenerateFacturXPDFA3(meta, ciiXML)
}

// Ensure decimal is referenced if needed by future helpers
var _ = decimal.Zero
