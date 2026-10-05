# Changelog

## [1.2.0] - 2026-10-05

### Added
- **Normalisation CII (`internal/model`)** : Prise en charge de la syntaxe CII-D16B vers le modèle canonique avec extraction de `DueDateDateTime` et conservation de la TVA par ligne (`VatPercent`, `VatCategory`).
- **Validation Normative EN 16931 (`internal/validator`)** :
  - Contrôle arithmétique de ligne : `Quantité × Prix = Montant HT` (`BR-LINE-NET-AMOUNT`).
  - Validation des sous-totaux de TVA (`BR-TAX-CALCULATION`, `BR-TAX-NONNEGATIVE`).
  - Cohérence stricte des totaux : somme des lignes (`BR-CO-10`) et cohérence TTC (`BR-CO-15`).
- **Validateur XSD (`internal/validator/xsd.go`)** : Interface de contrôle de schéma amont et arborescence `schemas/cii/` et `schemas/facturx/`.
- **Validation externe PDF/A (`internal/exporter/pdfa_validator.go`)** : Intégration de l'interface VeraPDF pour certification CI.
- **Pipeline d'intégration (`internal/integration`)** : Suite de tests bout en bout couvrant CII -> Canonical -> Validation EN 16931 -> Export Factur-X PDF/A-3.

### Security & Hardening
- **Générateur Factur-X (`internal/exporter/facturx_pdfa3.go`)** : Validation préalable du flux XML (rejet des flux non-CII, vides ou mal formés).
- Clarification du statut de `VerifyFacturXContainer` en tant que smoke test structurel (séparation avec la validation externe PDF/A-3b).
