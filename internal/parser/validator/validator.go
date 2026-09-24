package validator

import (
	"fmt"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
)

func Validate(invoice *models.Invoice) []error {
	var errors []error

	if invoice == nil {
		return []error{
			fmt.Errorf("invoice is nil"),
		}
	}

	if strings.TrimSpace(invoice.InvoiceNumber) == "" {
		errors = append(errors, fmt.Errorf("invoice number is required"))
	}

	if strings.TrimSpace(invoice.Seller.Name) == "" {
		errors = append(errors, fmt.Errorf("seller name is required"))
	}

	if strings.TrimSpace(invoice.Buyer.Name) == "" {
		errors = append(errors, fmt.Errorf("buyer name is required"))
	}

	if strings.TrimSpace(invoice.Currency) == "" {
		errors = append(errors, fmt.Errorf("currency is required"))
	}

	if invoice.Amounts.Total < 0 {
		errors = append(errors, fmt.Errorf("total amount cannot be negative"))
	}

	return errors
}