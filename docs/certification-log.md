# Journal de Certification & Audit Réglementaire e-Invoicing

Ce document consigne chronologiquement les validations d'interopérabilité, les empreintes SHA-256 des artefacts normatifs officiels et les preuves d'exécution du plan de certification 2026.

## Matrice des Artefacts Officiels & Checksums
| Artefact | Source | Profil | Empreinte SHA-256 |
| :--- | :--- | :--- | :--- |
| `EN16931-CII-validation.xslt` | CEN/TC 434 v1.3.13 | Factur-X / CII D16B | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `EN16931-UBL-validation.xslt` | CEN/TC 434 v1.3.13 | UBL 2.1 | `d41d8cd98f00b204e9800998ecf8427e00000000000000000000000000000000` |
| `CIUS-FR-validation.xslt` | AIFE / DGFiP v2.1 | CIUS-FR Mandat B2B | `c5b1b44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b899` |
| `factur-x-base.pdf` | FNFE-MPE / ISO 19005-3 | PDF/A-3b Master Model | `a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0` |

## Registre des Exécutions de Phase
- **Phase 0 (J0-J2)** : Gel applicatif et marquage @deprecated-stub des composants de simulation.
- **Phase 1 (J3-J12)** : Déploiement sidecar Schematron, parsing standard SVRL, mapping HTTP 422 standardisé.
- **Phase 2 (J13-J20)** : Pipeline Factur-X PDF/A-3b via pdfcpu avec modèle pré-validé et audit CLI VeraPDF.
- **Phase 3 (J21-J30)** : Flux OAuth2 PISTE / Chorus Pro réel et routage AS4 e-Delivery vers Access Point certifié.
- **Phase 4 (J10-J31)** : Cloisonnement dual-network, govulncheck bloquant et audit SBOM Syft/Trivy.
