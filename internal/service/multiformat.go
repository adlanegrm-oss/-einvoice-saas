package service

import (
	"context"
	"fmt"

	exp_ksef "github.com/adlanegrm-oss/einvoice-saas/internal/exporter/ksef"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/parser/fatturapa"
	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

type MultiFormatEngine struct{}

func NewMultiFormatEngine() *MultiFormatEngine {
	return &MultiFormatEngine{}
}

func (m *MultiFormatEngine) IngestPayload(ctx context.Context, payload []byte) (*invoice.Invoice, validator.FormatType, error) {
	format := validator.SniffDocumentFormat(payload)

	var inv *invoice.Invoice
	var err error

	switch format {
	case validator.FormatFatturaPA:
		inv, err = fatturapa.Parse(payload)
	default:
		return nil, format, fmt.Errorf("format de document non pris en charge pour ce test: %s", format)
	}

	if err != nil {
		return nil, format, fmt.Errorf("erreur de parsing [%s]: %w", format, err)
	}

	return inv, format, nil
}

func (m *MultiFormatEngine) ConvertToTarget(inv invoice.Invoice, target validator.FormatType) ([]byte, error) {
	switch target {
	case validator.FormatKSeF:
		return exp_ksef.GenerateXML(inv)
	default:
		return nil, fmt.Errorf("format cible non supporté: %s", target)
	}
}
