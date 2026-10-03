# Tickets GitHub prêts à coller — eInvoice SaaS

## Sprint 0 (immédiat — infra)

### [P0] Utiliser docker-compose.hardened.yml + reverse-proxy TLS
**Description**  
Remplacer le compose standard par le hardened. Aucun port applicatif publié. Configurer Caddy (ou équivalent) avec Caddyfile.example.

**Critères d'acceptation**
- `docker compose -f docker-compose.hardened.yml config` valide
- Aucun `ports:` dans le compose
- Accès uniquement via HTTPS du proxy
- preflight.sh passe

---

### [P0] Régénérer tous les secrets + purger Git
**Description**  
Générer de nouveaux JWT_SECRET_* et ADMIN_PASSWORD_* (rotate-secrets.sh).  
Si d'anciens mots de passe ont été commités, purger avec `git filter-repo`.

**Critères**
- Nouveaux secrets ≥ 32 / 12 caractères
- Aucun secret réel dans l'historique Git
- .env non versionné

---

## Sprint 1 (sécurité + multi-tenant)

### [P0] Persistance des comptes utilisateurs en SQLite
**Fichiers** : `internal/auth/`, `internal/repository/`

Créer table `users` (id, email, password_hash, role, tenant_id, created_at…).  
Migration + hash PBKDF2-HMAC-SHA256 (600k itérations).  
Les comptes survivent au redémarrage.

### [P0] Validation JWT stricte
Valider explicitement `iss`, `aud`, `exp`, `nbf`.  
Refuser algorithme non autorisé.  
Le tenant_id ne doit **jamais** provenir d'un header client.

### [P0] Isolation tenant sur tous les endpoints
Ajouter tests négatifs inter-tenant (accès à une facture / archive d'un autre tenant → 404).  
Jobs asynchrones transportent le contexte tenant.

### [P1] Rate limiting par tenant + IP
Token bucket sur routes sensibles (emit, deposit, login, forgot-password).

### [P0] Protection upload renforcée
Magic bytes, taille max, refus DOCTYPE/XXE, path traversal.

---

## Sprint 2 (conformité)

### [P0] Validation XSD + Schematron EN 16931 / PEPPOL BIS
Intégrer un validateur XSD (libxml2 via pure Go ou service externe).  
Retourner le code BR exact en 422.

### [P0] Factur-X PDF/A-3 complet
Enrichir vendeur (SIREN/TVA), devise, montant à payer, XMP complet.  
Vérifier correspondance PDF ↔ XML embarqué.

### [P0] Audit trail chaîné
`hash_n = SHA256(canonical_event_n || hash_{n-1})`  
Stockage append-only + vérification d'intégrité périodique.

### [P0] Idempotence réelle sur POST /invoices/emit
Clé d'idempotence + gestion des doublons (retourner l'existant si hash identique).

---

## Sprint 3 (industrialisation)

### [P1] Connecteur AS4 / PDP réel
Retry + backoff + jitter, DLQ, corrélation message-id, TLS strict, anti-rejeu.

### [P1] Observabilité
`/metrics` Prometheus (labels tenant sans fuite), logs structurés, alertes.

### [P0] PRA formalisé
Script backup + test restauration documenté (RPO/RTO). Voir PRA_RUNBOOK.md.

### [P1] CI renforcée
govulncheck + SBOM + scan image + tests de charge automatiques.
