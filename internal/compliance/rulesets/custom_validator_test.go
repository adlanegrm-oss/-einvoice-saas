package rulesets_test

import (
"context"
"errors"
"testing"

"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/rulesets"
"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

func TestNoopValidator(t *testing.T) {
v := &rulesets.NoopValidator{}
inv := &canonical.CanonicalInvoice{}
errs := v.Validate(context.Background(), "TENANT_C", inv)
if len(errs) != 0 {
t.Fatalf("attendu 0 erreur pour NoopValidator, reçu %d", len(errs))
}
}

func TestRegistryValidator_TenantIsolation(t *testing.T) {
reg := rulesets.NewRegistryValidator()

// Règle spécifique pour CLIENT_M : le numéro de facture (ID) ne doit pas être vide
reg.RegisterRule("CLIENT_M", func(inv *canonical.CanonicalInvoice) error {
if inv.ID == "" {
return errors.New("BR-CLIENT-M-01: identifiant de facture obligatoire pour Client M")
}
return nil
})

ctx := context.Background()

// 1. Facture sans ID pour CLIENT_STANDARD -> Ne doit pas échouer
standardInv := &canonical.CanonicalInvoice{ID: ""}
errsStandard := reg.Validate(ctx, "CLIENT_STANDARD", standardInv)
if len(errsStandard) != 0 {
t.Errorf("CLIENT_STANDARD ne doit pas être bloqué par les règles de CLIENT_M")
}

// 2. Facture sans ID pour CLIENT_M -> Doit échouer avec l'erreur précise
mInvWithoutID := &canonical.CanonicalInvoice{ID: ""}
errsM := reg.Validate(ctx, "CLIENT_M", mInvWithoutID)
if len(errsM) != 1 {
t.Fatalf("attendu 1 erreur pour CLIENT_M, reçu %d", len(errsM))
}
if errsM[0].Error() != "BR-CLIENT-M-01: identifiant de facture obligatoire pour Client M" {
t.Errorf("message d'erreur inattendu : %v", errsM[0])
}

// 3. Facture conforme pour CLIENT_M -> Doit passer
mInvWithID := &canonical.CanonicalInvoice{ID: "INV-2026-CLIENT-M"}
errsMValid := reg.Validate(ctx, "CLIENT_M", mInvWithID)
if len(errsMValid) != 0 {
t.Errorf("CLIENT_M avec ID valide doit passer, erreurs: %v", errsMValid)
}
}
