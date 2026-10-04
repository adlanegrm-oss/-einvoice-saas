package rulesets

import (
"context"
"fmt"

"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

// BusinessRule représente une fonction de contrôle unitaire sur une facture canonique.
type BusinessRule func(inv *canonical.Invoice) error

// CustomValidator définit le contrat d'extension pour les règles de gestion spécifiques aux clients/tenants.
type CustomValidator interface {
Validate(ctx context.Context, tenantID string, inv *canonical.Invoice) []error
}

// NoopValidator est l'implémentation neutre par défaut (aucun contrôle spécifique supplémentaire).
type NoopValidator struct{}

// Validate renvoie toujours nil.
func (n *NoopValidator) Validate(_ context.Context, _ string, _ *canonical.Invoice) []error {
return nil
}

// RegistryValidator permet d'enregistrer et d'exécuter des règles par tenant à la volée.
type RegistryValidator struct {
rules map[string][]BusinessRule
}

// NewRegistryValidator instancie un validateur basé sur un registre de règles spécifiques.
func NewRegistryValidator() *RegistryValidator {
return &RegistryValidator{
rules: make(map[string][]BusinessRule),
}
}

// RegisterRule associe une règle de validation personnalisée à un tenantId précis.
func (r *RegistryValidator) RegisterRule(tenantID string, rule BusinessRule) {
if r.rules == nil {
r.rules = make(map[string][]BusinessRule)
}
r.rules[tenantID] = append(r.rules[tenantID], rule)
}

// Validate exécute toutes les règles enregistrées pour le tenant spécifié.
func (r *RegistryValidator) Validate(_ context.Context, tenantID string, inv *canonical.Invoice) []error {
if inv == nil {
return []error{fmt.Errorf("facture inexistante")}
}
registered, ok := r.rules[tenantID]
if !ok || len(registered) == 0 {
return nil
}

var errs []error
for _, rule := range registered {
if err := rule(inv); err != nil {
errs = append(errs, err)
}
}
return errs
}
