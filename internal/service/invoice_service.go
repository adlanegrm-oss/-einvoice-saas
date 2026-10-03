package service

import (
	"fmt"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

type InvoiceService struct{}

func NewInvoiceService() *InvoiceService {
	return &InvoiceService{}
}

func (s *InvoiceService) ValidateInvoice(invoice *models.Invoice) error {
	if invoice == nil {
		return fmt.Errorf("invoice is nil")
	}

	errors := validator.Validate(invoice)

	if len(errors) > 0 {
		return fmt.Errorf("invoice validation failed: %v", errors)
	}

	return nil
}
