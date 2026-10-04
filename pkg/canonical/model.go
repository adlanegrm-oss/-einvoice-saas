package canonical

import "time"

type CanonicalInvoice struct {
ID             string         `json:"id"`
TenantID       string         `json:"tenant_id"`
InvoiceNumber  string         `json:"invoice_number"`
IssueDate      time.Time      `json:"issue_date"`
DueDate        time.Time      `json:"due_date"`
Currency       string         `json:"currency"`
Seller         Party          `json:"seller"`
Buyer          Party          `json:"buyer"`
Lines          []LineItem     `json:"lines"`
TaxBreakdowns  []TaxBreakdown `json:"tax_breakdowns"`
MonetaryTotals MonetaryTotals `json:"monetary_totals"`
PaymentTerms   PaymentTerms   `json:"payment_terms"`
}

type Party struct {
ID             string `json:"id"`
TaxID          string `json:"tax_id"`
Name           string `json:"name"`
CountryCode    string `json:"country_code"`
AddressLine    string `json:"address_line"`
PostalCode     string `json:"postal_code"`
City           string `json:"city"`
RoutingAddress string `json:"routing_address"`
}

type LineItem struct {
LineID       string `json:"line_id"`
Description  string `json:"description"`
Quantity     int64  `json:"quantity"`
UnitPriceHT  int64  `json:"unit_price_ht"`
LineHT       int64  `json:"line_ht"`
VATRateBasis int64  `json:"vat_rate"`
VATCategory  string `json:"vat_category"`
}

type TaxBreakdown struct {
VATCategory string `json:"vat_category"`
VATRate     int64  `json:"vat_rate"`
BaseHT      int64  `json:"base_ht"`
TaxAmount   int64  `json:"tax_amount"`
}

type MonetaryTotals struct {
NetHT      int64 `json:"net_ht"`
TaxAmount  int64 `json:"tax_amount"`
GrossTTC   int64 `json:"gross_ttc"`
PayableDue int64 `json:"payable_due"`
}

type PaymentTerms struct {
PaymentMeansCode string `json:"payment_means_code"`
IBAN             string `json:"iban,omitempty"`
BIC              string `json:"bic,omitempty"`
}