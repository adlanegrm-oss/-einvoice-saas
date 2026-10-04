package canonical

import (
	"time"
)

type Party struct {
	Name           string `json:"name"`
	LegalEntityID  string `json:"legal_entity_id"` // SIREN (9) or SIRET (14) for FR
	TaxID          string `json:"tax_id"`          // VAT / NIF, e.g. FR12345678901
	CountryCode    string `json:"country_code"`    // ISO 2-letter, e.g. FR
	AddressLine1   string `json:"address_line_1"`
	PostalCode     string `json:"postal_code"`
	CityName       string `json:"city_name"`
	ElectronicAddr string `json:"electronic_addr"` // Peppol Endpoint ID / SIRET scheme
	EndpointScheme string `json:"endpoint_scheme"` // e.g. 0009 (SIREN), 0002 (SIRET), 9957 (FR:VAT)
}

type InvoiceLine struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description,omitempty"`
	Quantity         float64 `json:"quantity"`
	UnitCode         string  `json:"unit_code"`                // UN/ECE Rec 20 (C62, H87, HUR, KGM, etc.)
	UnitPriceNet     int64   `json:"unit_price_net_cents"`     // Base net price in minor units (cents)
	LineExtensionNet int64   `json:"line_extension_net_cents"` // Total line net amount
	TaxCategory      string  `json:"tax_category"`             // 'S', 'Z', 'E', 'AE'
	TaxRatePercent   float64 `json:"tax_rate_percent"`
}

type TaxBreakdown struct {
	TaxCategory     string  `json:"tax_category"`
	TaxRatePercent  float64 `json:"tax_rate_percent"`
	BaseHT          int64   `json:"base_ht_cents"`
	TaxAmount       int64   `json:"tax_amount_cents"`
	ExemptionCode   string  `json:"exemption_code,omitempty"`
	ExemptionReason string  `json:"exemption_reason,omitempty"`
}

type MonetaryTotals struct {
	LineExtensionTotal int64 `json:"line_extension_total_cents"`
	NetHT              int64 `json:"net_ht_cents"`
	TaxAmount          int64 `json:"tax_amount_cents"`
	GrossTTC           int64 `json:"gross_ttc_cents"`
	PrepaidAmount      int64 `json:"prepaid_amount_cents"`
	PayableDue         int64 `json:"payable_due_cents"`
}

type PaymentMeans struct {
	TypeCode string `json:"type_code"` // UNTDID 4461
	IBAN     string `json:"iban,omitempty"`
	BIC      string `json:"bic,omitempty"`
}

type CanonicalInvoice struct {
	ID               string    `json:"id"`
	InvoiceNumber    string    `json:"invoice_number"`
	IssueDate        time.Time `json:"issue_date"`
	DueDate          time.Time `json:"due_date"`
	TypeCode         string    `json:"type_code"` // UNTDID 1001: 380, 381
	Currency         string    `json:"currency"`  // ISO 4217, e.g. EUR
	BuyerReference   string    `json:"buyer_reference"`
	PurchaseOrderRef string    `json:"purchase_order_ref,omitempty"`
	PaymentTerms     string    `json:"payment_terms,omitempty"`
	ProfileURN       string    `json:"profile_urn"`

	Seller         Party          `json:"seller"`
	Buyer          Party          `json:"buyer"`
	PaymentMeans   PaymentMeans   `json:"payment_means"`
	TaxBreakdowns  []TaxBreakdown `json:"tax_breakdowns"`
	Lines          []InvoiceLine  `json:"lines"`
	MonetaryTotals MonetaryTotals `json:"monetary_totals"`
}
