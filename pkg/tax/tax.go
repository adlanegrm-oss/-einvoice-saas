package tax

import (
	"fmt"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/money"
)

// Category represents UNCL5305 tax category codes.
type Category string

const (
	StandardRate    Category = "S"  // Standard rate
	ZeroRated       Category = "Z"  // Zero rated goods
	Exempt          Category = "E"  // Exempt from tax
	ReverseCharge   Category = "AE" // VAT Reverse charge
	VATExemptEEA    Category = "K"  // VAT exempt for EEA intra-community supply
	FreeExport      Category = "G"  // Free export item, tax not charged
	ServicesOutside Category = "O"  // Services outside scope of tax
)

// TaxPercent represents VAT rate in basis points (e.g. 2000 = 20.00%, 550 = 5.50%).
type TaxPercent uint16

func NewTaxPercent(basisPoints uint16) TaxPercent {
	return TaxPercent(basisPoints)
}

func FromFloatPercent(pct float64) TaxPercent {
	return TaxPercent(money.RoundHalfEven(pct * 100.0))
}

func (tp TaxPercent) Float64() float64 {
	return float64(tp) / 100.0
}

func (tp TaxPercent) String() string {
	return fmt.Sprintf("%.2f%%", tp.Float64())
}

// Subtotal represents EN 16931 VAT breakdown item (BG-23).
type Subtotal struct {
	Category        Category    // BT-118
	Rate            TaxPercent  // BT-119 (0% for E, AE, Z, etc.)
	TaxableAmount   money.Money // BT-116
	TaxAmount       money.Money // BT-117
	ExemptionReason string      // BT-120 (mandatory if E or AE)
	ExemptionCode   string      // BT-121
}

// CalculateVATBreakdown calculates the exact VAT amount using half-even rounding.
func CalculateVATBreakdown(cat Category, rate TaxPercent, taxable money.Money, reason, code string) (Subtotal, error) {
	if (cat == Exempt || cat == ReverseCharge) && reason == "" && code == "" {
		return Subtotal{}, fmt.Errorf("rule BR-E-01/BR-AE-01: exemption reason or code is mandatory for category %s", cat)
	}

	var vatAmount money.Money
	if cat == Exempt || cat == ReverseCharge || cat == ZeroRated || rate == 0 {
		vatAmount = money.New(0, taxable.Currency())
	} else {
		rawVAT := float64(taxable.Cents()) * (float64(rate) / 10000.0)
		vatAmount = money.New(money.RoundHalfEven(rawVAT), taxable.Currency())
	}

	return Subtotal{
		Category:        cat,
		Rate:            rate,
		TaxableAmount:   taxable,
		TaxAmount:       vatAmount,
		ExemptionReason: reason,
		ExemptionCode:   code,
	}, nil
}
