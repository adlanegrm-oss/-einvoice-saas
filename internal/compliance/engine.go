package compliance

import (
	"fmt"
	"strings"

	"einvoice-saas/internal/compliance/validators/ma"
	"einvoice-saas/internal/model"
)

// Dispatcher gère l'enregistrement et l'exécution des validateurs de juridictions
type Dispatcher struct {
	validators map[string]model.JurisdictionValidator
}

// NewDispatcher initialise le répartiteur avec les validateurs disponibles
func NewDispatcher() *Dispatcher {
	d := &Dispatcher{
		validators: make(map[string]model.JurisdictionValidator),
	}
	// Enregistrement du profil Maroc canonique
	d.Register(ma.NewMoroccoCanonicalValidator())
	return d
}

// Register ajoute ou remplace un validateur de juridiction
func (d *Dispatcher) Register(v model.JurisdictionValidator) {
	d.validators[strings.ToUpper(v.JurisdictionCode())] = v
}

// Validate traite une facture canonique selon sa juridiction cible
func (d *Dispatcher) Validate(inv *model.CanonicalInvoice) (model.ValidationReport, error) {
	if inv == nil {
		return model.ValidationReport{}, fmt.Errorf("facture canonique nulle")
	}

	jurisdiction := strings.ToUpper(strings.TrimSpace(inv.TargetJurisdiction))
	if jurisdiction == "" {
		jurisdiction = "FR" // Juridiction par défaut
	}

	validator, exists := d.validators[jurisdiction]
	if !exists {
		return model.ValidationReport{
			Jurisdiction: jurisdiction,
			Valid:        false,
			Issues: []model.ValidationIssue{
				{
					RuleID:      "SYS-JURISDICTION-UNSUPPORTED",
					Description: fmt.Sprintf("Aucun validateur actif configuré pour la juridiction: %s", jurisdiction),
					Severity:    model.SeverityError,
					Field:       "TargetJurisdiction",
					Remediation: "Vérifier le code pays cible ou activer le CountryProfile correspondant.",
				},
			},
		}, nil
	}

	return validator.Validate(inv), nil
}