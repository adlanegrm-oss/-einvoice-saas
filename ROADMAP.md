# Roadmap & Statut de Qualification - einvoice-saas

## Statut Global d'Implémentation

| Périmètre | Phase | Statut | Preuves & Validation |
| :--- | :--- | :--- | :--- |
| **Infra & Hardening** | Sprint 0 | **Livré & Durci** | Docker non-root, read-only/tmpfs, Caddy TLS/HSTS, scripts d'exploitation (`scripts/`). |
| **Sécurité & Multi-Tenant** | Sprint 1 | **Validé (Code & Tests)** | Hash scrypt, JWT, isolation stricte `owner`, anti-bruteforce (`internal/auth`, `internal/repository`). |
| **Conformité & Formats** | Sprint 2 | **Validé (Code & Tests)** | Norme EN 16931, arithmétique financière exacte, parsers UBL/EDIFACT/Factur-X (`internal/compliance`, `internal/validator`). |
| **Routage & Cycle de Vie** | Sprint 3 | **Validé (Code & Tests)** | Machine à états, piste d'audit, client d'acheminement AS4 simulé & direct (`internal/routing/as4`, `internal/service`). |
| **Exposition & Réseau Prod** | Sprint 4 | **En cours / Spécifié** | Déploiement dual-network pour flux sortants AS4/SMTP réels et interconnexion PDP/PEPPOL live. |

---

## Synthèse des Tests & Qualification

- **Tests unitaires et d'intégration** : 100 % passants (`go test -v ./...` consigné dans `test_results.log`).
  - Validation du cloisonnement inter-tenants (404 sur les tentatives d'accès croisées).
  - Validation de l'idempotence et audit trail (`invoice_status_history`).
  - Zéro dérive binaire sur les montants financiers.
- **Résilience sous charge (SQLite WAL)** : Validée via `test_load_qualification.go`.
  - Concurrence soutenue : 2 500 requêtes / 50 workers.
  - Débit mesuré : ~750 à 950 req/s.
  - Erreurs serveur (5xx) : 0.

---

## Prochaines étapes

1. Déploiement du schéma réseau conteneurisé dual-network (accès sortant TLS contrôlé).
2. Interconnexion réelle aux passerelles AS4 certifiées / endpoints PDP de test.
3. Remplacement définitif des scripts de mise à jour manuelle par un gestionnaire de migrations SQL versionné.
