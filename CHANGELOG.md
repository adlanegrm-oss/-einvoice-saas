# Changelog — Package de Mise à Jour eInvoice SaaS

## [1.1.0] — 2026-10-02 — Package Sprints 0-3

### Sprint 0 — Sécurisation immédiate
- Suppression de l’exposition directe des ports applicatifs (`ports:` → `expose:`)
- Ajout d’un réseau Docker interne (`internal: true`)
- Ajout de `cap_drop: ALL`, `init: true`, limites CPU/mémoire/PID
- Conservation de `read_only`, tmpfs et `no-new-privileges`
- Variables de production rendues explicitement obligatoires
- Modèle de reverse-proxy TLS (Caddy) fourni
- Script de préflight (`preflight.sh`)
- Script de backup SQLite avec `PRAGMA integrity_check`

### Sprint 1 — Sécurité & Multi-tenant (à implémenter dans le code)
- Persistance des comptes utilisateurs (table `users`)
- Validation JWT stricte (iss, aud, exp, nbf)
- Isolation tenant vérifiée sur tous les endpoints
- Rate limiting par tenant
- Protection upload renforcée (magic bytes, XXE)
- Headers de sécurité unifiés

### Sprint 2 — Conformité & Formats (à implémenter dans le code)
- Validation XSD + Schematron EN 16931 / PEPPOL BIS
- Finalisation générateur Factur-X PDF/A-3
- Pipeline multi-format unifié
- Audit trail chaîné (hash_{n} = SHA256(event_n || hash_{n-1}))
- Idempotence sur POST /invoices/emit

### Sprint 3 — Industrialisation
- Observabilité (Prometheus)
- PRA formalisé (RPO/RTO + runbook)
- CI renforcée (govulncheck, SBOM, scan image)
- Documentation d’exploitation mise à jour

### Notes
- Aucun secret réel n’est inclus dans ce package.
- Le code Go métier n’est pas réécrit ici (voir PLAN_SPRINTS.md).
