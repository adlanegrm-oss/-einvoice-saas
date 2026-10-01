package edifact

import (
	"fmt"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
)

func Parse(data []byte) (*models.Invoice, error) {
	content := strings.TrimSpace(string(data))

	if content == "" {
		return nil, fmt.Errorf("empty EDIFACT document")
	}

	segments := strings.Split(content, "'")

	invoice := &models.Invoice{}

	for _, segment := range segments {
		segment = strings.TrimSpace(segment)

		if segment == "" {
			continue
		}

		fields := strings.Split(segment, "+")

		switch fields[0] {
		case "BGM":
			if len(fields) > 2 {
				invoice.InvoiceNumber = fields[2]
			}

		case "DTM":
			if len(fields) > 1 {
				invoice.Status = "parsed"
			}
		}
	}

	if invoice.InvoiceNumber == "" {
		invoice.InvoiceNumber = "UNKNOWN"
	}

	return invoice, nil
}