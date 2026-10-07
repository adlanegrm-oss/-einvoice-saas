package parser

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

const (
	MaxInvoiceXMLSize = 10 * 1024 * 1024
)

var (
	ErrPayloadTooLarge = errors.New("xml payload exceeds maximum allowed size (10MB)")
	ErrSecurityEntity  = errors.New("security error: DTD and external XML entities are forbidden")
)

func HardenedXMLReader(raw io.Reader) ([]byte, error) {
	limitedReader := io.LimitReader(raw, MaxInvoiceXMLSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input stream: %w", err)
	}
	if len(data) > MaxInvoiceXMLSize {
		return nil, ErrPayloadTooLarge
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	decoder.Entity = map[string]string{}

	for {
		token, err := decoder.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("malformed xml token: %w", err)
		}

		switch tok := token.(type) {
		case xml.Directive:
			dStr := string(tok)
			if bytes.Contains(tok, []byte("DOCTYPE")) || bytes.Contains(tok, []byte("ENTITY")) || bytes.Contains(tok, []byte("SYSTEM")) {
				return nil, fmt.Errorf("%w: directive '%s' rejected", ErrSecurityEntity, dStr)
			}
		}
	}

	return data, nil
}
