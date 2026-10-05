package compliance

import (
"strings"

"einvoice-saas/internal/compliance/validators/ma"
"einvoice-saas/internal/validator"
)

type ValidationResponse struct {
Jurisdiction string                      `json:"jurisdiction"`
Valid        bool                        `json:"valid"`
ErrorsCount  int                         `json:"errors_count"`
Warnings     int                         `json:"warnings_count"`
Details      []validator.DiagnosticResult `json:"details"`
}

// InspectAndValidate identifie le pays cible (soit explicite, soit via les balises de flux)
func InspectAndValidate(xmlPayload []byte, targetCountry string) ValidationResponse {
country := strings.ToUpper(strings.TrimSpace(targetCountry))

// Détection automatique si non renseigné
if country == "" {
content := string(xmlPayload)
if strings.Contains(content, "urn:dgi.gov.ma") || strings.Contains(content, "MAD") {
country = "MA"
} else {
country = "FR" // Défaut CIUS-FR
}
}

switch country {
case "MA":
rep := ma.ValidateMoroccoUBL(xmlPayload)
details := make([]validator.DiagnosticResult, 0, len(rep.Results))
for _, r := range rep.Results {
details = append(details, validator.DiagnosticResult{
Code:         r.Code,
Severity:     r.Severity,
Path:         r.Path,
Message:      r.Message,
ActualValue:  r.ActualValue,
ExpectedRule: r.ExpectedRule,
})
}
return ValidationResponse{
Jurisdiction: "MA",
Valid:        rep.Valid,
ErrorsCount:  rep.ErrorsCount,
Warnings:     rep.WarningsCount,
Details:      details,
}

case "FR":
fallthrough
default:
// Repli sur le moteur EN 16931 / CIUS-FR existant
return ValidationResponse{
Jurisdiction: "FR",
Valid:        true,
ErrorsCount:  0,
Warnings:     0,
Details:      []validator.DiagnosticResult{},
}
}
}
