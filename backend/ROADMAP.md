# 🗺️ Roadmap de Certification & Production eInvoice SaaS

## Statut d'Implémentation Actuel (Release v1.3.0 - Octobre 2026)

### ✅ P0 – Bloquants Conformité & Sécurité (100% Terminés)
- [x] **Validation normative complète EN 16931 + PEPPOL BIS v3** : Moteur SAX/Token avec codes officiels d'erreurs SVRL (`BR-01` à `BR-65`, `BR-CO-*`, `CIUS-FR-*`) retournés en HTTP 422 Unprocessable Entity.
- [x] **Génération Factur-X PDF/A-3 (ISO 19005-3)** : Pure Go, métadonnées XMP `pdfaExtension` namespace `fx`, balisage `/AFRelationship /Alternative`, profils MINIMUM, BASIC, EN 16931.
- [x] **Connecteurs de transmission réels** :
  - Client AS4/ebMS3 avec corrélation `MessageId`, retry exponentiel, gestion Dead Letter Queue (DLQ).
  - Connecteur Chorus Pro / PPF (API PISTE) & connecteur KSeF Pologne.
- [x] **Persistance utilisateurs & RBAC** : Table `users` avec migrations SQL versionnées, rôles (`admin`, `operator`, `auditor`).
- [x] **Sécurité durcie** : Validation JWT stricte (HMAC-SHA256, `iss`, `aud`, `exp`, `nbf`), headers CSP stricts, upload guard contre XXE et path traversal.

### 🔄 P1 – Robustesse & Industrialisation
- [x] Observabilité Prometheus avec labels tenant-safe (`tenant_id`, `rule_id`).
- [x] PRA formalisé : Scripts de sauvegarde et restauration chiffrée AES-256 (`scripts/backup_pra.sh`, `scripts/restore_pra.sh`) avec RPO < 1h et RTO < 15 min.
- [x] Machine à états et idempotence verrouillées en base SQLite/PostgreSQL.

### ⏳ P2 – Extensions Multi-pays (En cours)
- [ ] Connecteur direct PEPPOL SMP / Access Point certifié.
- [ ] Validation Schematron FatturaPA (Italie) & Facturae (Espagne).
