# Correction Package pour einvoice-saas (Chantier A & Chantier B)

Ce package contient les corrections et les modules validés pour le dépôt `einvoice-saas` (modules Go, normalizers, validateurs et gestion des moteurs de conformité EN 16931).

## Sommaire des Corrections (Chantier A)
1. **Modèle Canonique en Decimal (`internal/model/canonical.go`)** :
   - Remplacement de tous les types `float64` monétaires par `decimal.Decimal` (bibliothèque `github.com/shopspring/decimal`).
   - Parsing rigoureux via `decimal.NewFromString` sur les données brutes XML (zéro tolérance, exactitude au centime).
   - Sérialisation JSON propre avec gestion correcte des décimaux.
2. **Champs et Références Étendus (`internal/model/`)** :
   - Ajout des champs normatifs (BT/BG) : `Sellers.VATID`, `Seller.TaxRegistrationID`, `Seller.LegalID`, `PrecedingInvoices` (BT-25/BT-26), `DocumentAllowancesCharges` (BG-20/BG-21), Totaux (`AllowanceTotal`, `ChargeTotal`, `PrepaidAmount`, `RoundingAmount`), `ExemptionReason` / `ExemptionReasonCode` (BT-120/121), `PaymentMeans` (IBAN BT-84), `BusinessProcessID`, et `GuidelineID` (BT-24).
3. **Normalizers CII & UBL (`internal/model/cii_normalizer.go`, `ubl_normalizer.go`)** :
   - CII : Correction de la lecture des immatriculations (`SpecifiedTaxRegistration` en slice pour capturer `VA` et `FC`), lecture des remises documentaires, des taxes et des références de factures d'avoir.
   - UBL : Lecture correcte du pays fournisseur et acheteur depuis `cac:PostalAddress/cac:Country/cbc:IdentificationCode` (fin du forçage de juridiction), lecture de `TaxExemptionReason`, `AllowanceCharge`, et des identifiants multiples.
4. **Détection de Syntaxe Robuste (`internal/app/router.go`)** :
   - Remplacement de `bytes.Contains` fragile par `DetectSyntax` basé sur l'élément racine XML (`Invoice`, `CreditNote`, `CrossIndustryInvoice`).
5. **Validation et Règles Arithmétiques (`internal/validator/`, `pkg/compliance/en16931/`)** :
   - Correction de l'étiquetage et des formules des règles EN 16931 (BR-CO-10 à BR-CO-17).
   - `FranceCanonicalValidator` : Contrôle SIREN/SIRET appliqué uniquement aux acheteurs français (via `Buyer.CountryCode == "FR"`).
6. **Tests de Caractérisation et Fixtures (`internal/app/fixtures_pipeline_test.go`, `factures_test_lots/`)** :
   - Suite de tests validant l'ensemble du lot de factures (`FACT_2026_003_CONFORME_EXO.xml` corrigé pour l'acheteur belge, tests d'avoirs 381 avec/sans référence, autoliquidation AE, export hors UE, etc.).

## Sommaire des Architectures (Chantier B)
- **Décision d'Architecture Schematron** : Comparaison chiffrée entre intégration Saxon-HE (Java bridge) vs Sidecar Go/WASM (Saxon-JS) vs Service Externe Python/Node. Choix recommandé et implémentation du wrapper SVRL normalisé.
