# 📋 Rapport d'Exécution des Tests & Historique de Release

* **Release** : `v1.2.0-compliance`
* **Commit** : `feat(validation): integrate official EN16931 Schematron validation & full D16B CII mapping`
* **Statut Global** : 🟢 PASS (10 packages vérifiés, 0 échec)

### Résultats de la Suite de Tests (`go test -v -p 1 ./pkg/...`)

| Package Go | Nom du Test Unitaire | Statut | Couverture / Règle validée |
| :--- | :--- | :--- | :--- |
| `pkg/validation` | `TestSchematron_GoldenValidInvoice` | 🟢 PASS | Facture de référence valide 100% conforme EN 16931 & CIUS-FR |
| `pkg/validation` | `TestSchematron_InvalidMathAmounts` | 🟢 PASS | Rejet ciblé sur incohérence arithmétique (Règle `BR-CO-15`) |
| `pkg/validation` | `TestSchematron_InvalidDueDate` | 🟢 PASS | Rejet si date d'échéance antérieure à émission (Règle `BR-CO-25`) |
| `pkg/validation` | `TestSchematron_InvalidFrenchSIRET` | 🟢 PASS | Rejet si identifiant légal vendeur FR non conforme (`CIUS-FR-01`) |
| `pkg/validation` | `TestValidateStrictEN16931` | 🟢 PASS | Seuil de tolérance strict et règles internes de sécurité |
| `pkg/validation` | `TestToleranceStrictRejection` | 🟢 PASS | Rejet des écarts d'arrondis supérieurs à 0.01 EUR |
| `pkg/syntax` | `TestValidateCoreBusinessRules` | 🟢 PASS | Intégrité de pré-sérialisation XML UN/CEFACT D16B |
| `pkg/as4` | `TestAS4ReplayAndReceipt` | 🟢 PASS | Protection anti-rejeu par empreinte et émission du receipt |
| `pkg/facturx` | `TestFacturXCoherence` | 🟢 PASS | Cohérence bilatérale des totaux PDF vs XML |
| `pkg/lifecycle` | `TestStateTransitions` | 🟢 PASS | Machine à états du cycle de vie facture |
| `pkg/money` | `TestRoundHalfEven` / `TestMoneyOperations` | 🟢 PASS | Arrondi banquier et opérations en centimes sans flottants |
| `pkg/pdp` | `TestPDPClientSubmit` | 🟢 PASS | Soumission HTTP sécurisée, jeton Bearer et clé d'idempotence |
| `pkg/tax` | `TestCalculateVATBreakdown` | 🟢 PASS | Calcul des assiettes et ventilation fiscale multi-taux |
| `pkg/webhook` | `TestVerifyHMACWebhook` | 🟢 PASS | Signature HMAC-SHA256 et rejet des replays de plus de 300s |
| `pkg/xmlsec` | `TestBlockXXE` / `TestValidXML` | 🟢 PASS | Blocage strict des attaques par injection XML (XXE) |
