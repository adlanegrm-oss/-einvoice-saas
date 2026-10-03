package en16931

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

var siretRegex = regexp.MustCompile(`^[0-9]{14}$`)

type ValidationError struct {
	RuleID  string `json:"rule_id"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] %s (champ: %s)", e.RuleID, e.Message, e.Field)
	}
	return fmt.Sprintf("[%s] %s", e.RuleID, e.Message)
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateInvoice(inv *invoice.Invoice) []ValidationError {
	var errs []ValidationError

	// BR-01 : Numéro de facture
	if strings.TrimSpace(inv.Number) == "" {
		errs = append(errs, ValidationError{
			RuleID:  "BR-01",
			Message: "Le numéro de facture est obligatoire",
			Field:   "number",
		})
	}

	// BR-02 : Date d'émission
	if inv.IssueDate.IsZero() {
		errs = append(errs, ValidationError{
			RuleID:  "BR-02",
			Message: "La date d'émission est obligatoire",
			Field:   "issue_date",
		})
	}

	// BR-CO-26 : Devise de la facture
	if len(strings.TrimSpace(string(inv.Currency))) != 3 {
		errs = append(errs, ValidationError{
			RuleID:  "BR-CO-26",
			Message: "La devise de la facture doit être un code ISO 4217 à 3 lettres (ex: EUR)",
			Field:   "currency",
		})
	}

	// BR-06 & BR-07 : Identité du vendeur
	if strings.TrimSpace(inv.Seller.Name) == "" {
		errs = append(errs, ValidationError{
			RuleID:  "BR-06",
			Message: "Le nom commercial du vendeur est obligatoire",
			Field:   "seller.name",
		})
	}
	if inv.Seller.Address.CountryCode == "FR" && !siretRegex.MatchString(inv.Seller.SIRET) {
		errs = append(errs, ValidationError{
			RuleID:  "BR-FR-01",
			Message: "Le SIRET du vendeur français doit contenir exactement 14 chiffres",
			Field:   "seller.siret",
		})
	}

	// BR-07 & BR-08 : Identité de l'acheteur
	if strings.TrimSpace(inv.Customer.Name) == "" {
		errs = append(errs, ValidationError{
			RuleID:  "BR-07",
			Message: "Le nom de l'acheteur est obligatoire",
			Field:   "customer.name",
		})
	}

	// BR-CO-10 : Présence d'au moins une ligne
	if len(inv.Items) == 0 {
		errs = append(errs, ValidationError{
			RuleID:  "BR-CO-10",
			Message: "La facture doit comporter au moins une ligne de détail",
			Field:   "items",
		})
		return errs
	}

	var sumLineTotalHT float64
	for i, item := range inv.Items {
		if item.Quantity <= 0 {
			errs = append(errs, ValidationError{
				RuleID:  "BR-22",
				Message: fmt.Sprintf("La quantité de la ligne %d doit être supérieure à 0", i+1),
				Field:   fmt.Sprintf("items[%d].quantity", i),
			})
		}
		if item.UnitPrice.Amount < 0 {
			errs = append(errs, ValidationError{
				RuleID:  "BR-27",
				Message: fmt.Sprintf("Le prix unitaire de la ligne %d ne peut pas être négatif", i+1),
				Field:   fmt.Sprintf("items[%d].unit_price", i),
			})
		}

		lineHT := round2(float64(item.Quantity) * item.UnitPrice.ToFloat())
		sumLineTotalHT += lineHT
	}

	// BR-CO-13 : Somme des lignes vs Total HT
	if math.Abs(sumLineTotalHT-inv.TotalHT.ToFloat()) > 0.05 {
		errs = append(errs, ValidationError{
			RuleID:  "BR-CO-13",
			Message: fmt.Sprintf("La somme des lignes HT (%.2f) diffère du Total HT déclaré (%.2f)", sumLineTotalHT, inv.TotalHT.ToFloat()),
			Field:   "total_ht",
		})
	}

	// BR-CO-15 : Total TTC = HT + TVA
	expectedTTC := round2(inv.TotalHT.ToFloat() + inv.TotalVAT.ToFloat())
	if math.Abs(expectedTTC-inv.TotalTTC.ToFloat()) > 0.02 {
		errs = append(errs, ValidationError{
			RuleID:  "BR-CO-15",
			Message: fmt.Sprintf("Le total TTC (%.2f) ne correspond pas à HT (%.2f) + TVA (%.2f)", inv.TotalTTC.ToFloat(), inv.TotalHT.ToFloat(), inv.TotalVAT.ToFloat()),
			Field:   "total_ttc",
		})
	}

	return errs
}

func round2(val float64) float64 {
	return math.Round(val*100) / 100
}
