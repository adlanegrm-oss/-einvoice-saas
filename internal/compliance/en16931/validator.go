package en16931

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	ErrMalformedXML    = errors.New("document XML mal formé")
	ErrUnsupportedRoot = errors.New("élément racine de facture non reconnu")
)

// SchematronValidator gère la validation syntaxique et les contrôles de profil SVRL/EN16931
type SchematronValidator struct{}

func NewSchematronValidator() *SchematronValidator {
	return &SchematronValidator{}
}

// ValidateSyntax contrôle la validité syntaxique XML de base (Well-formedness)
func (v *SchematronValidator) ValidateSyntax(xmlData []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	decoder.Strict = true
	for {
		t, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("%w: %v", ErrMalformedXML, err)
		}
		_ = t
	}
	return nil
}

// QuickValidateProfile applique les assertions clés de l'EN 16931 (PEPPOL / Chorus / DGFIP)
func (v *SchematronValidator) QuickValidateProfile(ctx context.Context, xmlData []byte) (*ValidationResult, error) {
	if err := v.ValidateSyntax(xmlData); err != nil {
		return nil, err
	}

	res := &ValidationResult{
		IsValid:     true,
		Diagnostics: make([]DiagnosticItem, 0),
	}

	content := string(xmlData)

	// Règle BR-01 : Spécification du processus métier (CustomizationID)
	if !strings.Contains(content, "urn:cen.eu:en16931:2017") {
		res.IsValid = false
		res.Diagnostics = append(res.Diagnostics, DiagnosticItem{
			RuleID:   "BR-01",
			Severity: "ERROR",
			Location: "/*:Invoice/*:CustomizationID",
			Message:  "Une facture conforme doit spécifier la spécification EN 16931 (urn:cen.eu:en16931:2017)",
		})
	}

	// Règle BR-02 : Numéro de facture obligatoire
	if match, _ := regexp.MatchString(`<(?:cbc:)?ID>[^<]+</(?:cbc:)?ID>`, content); !match {
		res.IsValid = false
		res.Diagnostics = append(res.Diagnostics, DiagnosticItem{
			RuleID:   "BR-02",
			Severity: "ERROR",
			Location: "/*:Invoice/*:ID",
			Message:  "Le numéro de facture (BT-1) est obligatoire et ne peut être vide",
		})
	}

	// Règle BR-03 : Date d'émission obligatoire
	if match, _ := regexp.MatchString(`<(?:cbc:)?IssueDate>\d{4}-\d{2}-\d{2}</(?:cbc:)?IssueDate>`, content); !match {
		res.IsValid = false
		res.Diagnostics = append(res.Diagnostics, DiagnosticItem{
			RuleID:   "BR-03",
			Severity: "ERROR",
			Location: "/*:Invoice/*:IssueDate",
			Message:  "La date d'émission (BT-2) est obligatoire et doit respecter le format AAAA-MM-JJ",
		})
	}

	// Règle BR-CO-04 : Code devise de la facture obligatoire
	if match, _ := regexp.MatchString(`<(?:cbc:)?DocumentCurrencyCode>[A-Z]{3}</(?:cbc:)?DocumentCurrencyCode>`, content); !match {
		res.IsValid = false
		res.Diagnostics = append(res.Diagnostics, DiagnosticItem{
			RuleID:   "BR-CO-04",
			Severity: "ERROR",
			Location: "/*:Invoice/*:DocumentCurrencyCode",
			Message:  "Le code devise de facturation (BT-5) doit comporter 3 lettres majuscules (ISO 4217)",
		})
	}

	return res, nil
}