# Changelog eInvoice SaaS

## [1.3.0] - 2026-10-04
### Added
- **Validation Normative EN16931 & PEPPOL BIS** : Moteur de conformité avec retour HTTP 422 détaillant les codes de règles BR et emplacements XPath.
- **Moteur PDF/A-3 Factur-X** : Génération conforme ISO 19005-3b avec association `/AFRelationship /Alternative` et XMP structuré.
- **Transport AS4 ebMS3 & DLQ** : Client sécurisé TLS 1.2+ avec exponential backoff et mise en file d'attente d'échec (Dead Letter Queue).
- **Connecteurs Fiscaux** : Module PPF / Chorus Pro et connecteur KSeF (Pologne).
- **Persistance Utilisateurs & RBAC** : Migration `000004_users_multitenant.up.sql` et gestion de sessions.
- **Sécurité Web** : En-têtes CSP stricts, protection contre les uploads XXE et validation stricte JWT.
- **PRA** : Scripts de sauvegarde à chaud et restauration chiffrée AES-256.

### Fixed
- Élimination des collisions de redéclaration de types dans `pkg/canonical` et `pkg/validation`.
- Remplacement des structures en mémoire volatiles par des tables SQLite persistantes.
