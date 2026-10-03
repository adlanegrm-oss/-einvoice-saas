package rulesets

import (
"strings"
)

type Diagnostic struct {
Code     string `json:"code"`
Severity string `json:"severity"`
Path     string `json:"path"`
Message  string `json:"message"`
}

type Report struct {
Valid          bool         `json:"valid"`
Standard       string       `json:"standard"`
Profile        string       `json:"profile"`
RuleVersion    string       `json:"rule_version"`
Diagnostics    []Diagnostic `json:"diagnostics"`
}

type RulePack2026 struct{}

func NewRulePack2026() *RulePack2026 {
return &RulePack2026{}
}

func (r *RulePack2026) Version() string {
return "FR-2026.1"
}

func (r *RulePack2026) Evaluate(invoiceNumber, sellerSIRET, buyerSIRET string, netTotal, taxTotal, grossTotal int64, hasLines bool) Report {
rep := Report{
Valid:       true,
Standard:    "EN16931",
Profile:     "FR-CIUS",
RuleVersion: r.Version(),
Diagnostics: make([]Diagnostic, 0),
}

addError := func(code, path, msg string) {
rep.Valid = false
rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
Code: code, Severity: "ERROR", Path: path, Message: msg,
})
}

if strings.TrimSpace(invoiceNumber) == "" {
addError("BR-01", "/Invoice/ID", "Numéro d'identification unique obligatoire manquant.")
}
if grossTotal != netTotal+taxTotal {
addError("BR-CO-09", "/Invoice/LegalMonetaryTotal/TaxInclusiveAmount", "Le montant TTC doit être strictement égal au Net HT + TVA.")
}
if !hasLines {
addError("BR-16", "/Invoice/InvoiceLine", "La facture doit comporter au moins une ligne.")
}
if len(strings.TrimSpace(sellerSIRET)) != 14 {
addError("FR-R-01", "/Invoice/AccountingSupplierParty", "Le vendeur français doit avoir un numéro SIRET à 14 chiffres.")
}
if strings.TrimSpace(buyerSIRET) == "" {
rep.Diagnostics = append(rep.Diagnostics, Diagnostic{
Code: "FR-W-01", Severity: "WARNING", Path: "/Invoice/AccountingCustomerParty", Message: "Identifiant acheteur manquant : risque de rejet de routage PDP.",
})
}

return rep
}
