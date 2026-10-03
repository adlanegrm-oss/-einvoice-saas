package validator

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

const (
	FormatFatturaPA FormatType = "FATTURAPA"
	FormatFacturae  FormatType = "FACTURAE"
	FormatKSeF      FormatType = "KSEF"
)

func SniffDocumentFormat(data []byte) FormatType {
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "UNA") || strings.HasPrefix(trimmed, "UNB") {
		return FormatEDIFACT
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil || errors.Is(err, io.EOF) {
			break
		}

		if se, ok := token.(xml.StartElement); ok {
			local := se.Name.Local
			space := strings.ToLower(se.Name.Space)

			switch {
			case local == "FatturaElettronica" || strings.Contains(space, "fatturapa"):
				return FormatFatturaPA
			case local == "Facturae" || strings.Contains(space, "facturae"):
				return FormatFacturae
			case local == "Faktura" && (strings.Contains(space, "ksef") || strings.Contains(space, "crd.gov.pl")):
				return FormatKSeF
			case local == "Invoice" && strings.Contains(space, "oasis:names:specification:ubl"):
				return FormatUBL
			case local == "CrossIndustryInvoice":
				return FormatFacturX
			}
		}
	}

	return FormatUnknown
}
