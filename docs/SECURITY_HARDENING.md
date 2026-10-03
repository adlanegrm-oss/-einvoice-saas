# Mesures de durcissement — eInvoice SaaS

## 1. Réseau & exposition (P0)

- Utiliser **uniquement** `docker-compose.hardened.yml`.
- Réseau Docker `internal: true` → l'application n'est accessible que depuis le reverse-proxy.
- Aucun `ports:` publié. Uniquement `expose: "8080"`.
- Reverse-proxy (Caddy / Traefik / nginx) termine le TLS (ports 80/443 uniquement).
- Firewall hôte : seuls 80/443 + accès admin (VPN/SSH) ouverts.

## 2. Secrets (P0)

- Secrets distincts par environnement (JWT + ADMIN_PASSWORD).
- Génération : `./scripts/rotate-secrets.sh` ou `openssl rand -base64 48`.
- Jamais de secret réel dans Git (`.env` dans `.gitignore`).
- Purger l'historique Git si d'anciens mots de passe ont fuité (`git filter-repo`).
- Rotation planifiée (90 jours recommandés).

## 3. Conteneur

Déjà présent dans le compose hardened :
- `read_only: true`
- `tmpfs: /tmp` (noexec,nosuid,nodev)
- `security_opt: no-new-privileges:true`
- `cap_drop: ALL`
- `init: true`
- `pids_limit: 256`
- `mem_limit: 512m` / `cpus: "1.0"`
- USER non-root (Dockerfile)

## 4. Headers de sécurité recommandés (middleware Go ou proxy)

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

## 5. JWT (à implémenter dans le code)

- Algorithme strict : HS256 uniquement (refuser "none" ou algos asymétriques non attendus).
- Validation obligatoire : `iss`, `aud`, `exp`, `nbf`.
- TTL court (ex: 8h max pour access token).
- Pas de confiance dans les headers client (`X-Tenant-ID`, `X-User-Role`).
- Le tenant_id doit venir exclusivement du claim JWT signé.

## 6. Multi-tenant

- Toutes les requêtes DB : `WHERE owner = ?` (ou tenant_id).
- Téléchargements : vérifier tenant + autorisation avant ouverture fichier.
- Jobs asynchrones : contexte tenant explicite.
- Tests négatifs inter-tenant obligatoires (404 systématique).

## 7. Upload / XML

- Taille max 10 Mo.
- Magic bytes + extension blanche.
- Refus DOCTYPE / entités externes (XXE).
- Désactivation résolution externe.
- Path traversal testé.
- Content-Disposition: attachment ; nosniff.
