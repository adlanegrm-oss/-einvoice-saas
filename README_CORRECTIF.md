# Correctif expert SaaS — plateforme E-Invoicing / PDP

## Périmètre réellement audité
Les artefacts fournis dans la conversation sont :
- `Dockerfile`
- `docker-compose.yml`
- `.env.example`
- `.gitignore`
- `.dockerignore`
- `document d exploitation.pdf`

Le code applicatif Go (`cmd/server`, handlers API, modèles SQLite, JWT, validation EN 16931, connecteurs AS4/PPF/PDP, stockage documentaire, workers, tests) n'a pas été fourni. Le présent correctif est donc **exécutable côté infrastructure/exploitation**, mais ne prétend pas corriger des vulnérabilités qui ne peuvent être vérifiées sans le code source.

## Ce que contient ce ZIP
- `docker-compose.hardened.yml` : overlay de durcissement pour les 3 environnements.
- `Caddyfile.example` : terminaison TLS en frontal ; l'application reste sur le réseau interne.
- `.env.example.hardened` : variables et garde-fous d'exploitation.
- `scripts/preflight.sh` : contrôle avant déploiement.
- `scripts/sqlite-backup.sh` : sauvegarde SQLite cohérente + vérification.
- `SECURITY_AUDIT.md` : constats métier/connectique/protocoles/sécurité et priorités.
- `OPERATIONS_CHECKLIST.md` : checklist de mise en production/PRA.
- `PATCH_NOTES.md` : changements et limites.

## Utilisation
1. Conserver le projet source original.
2. Copier les fichiers de ce ZIP dans le dépôt.
3. Adapter `.env` à partir de `.env.example.hardened`.
4. Exécuter `scripts/preflight.sh`.
5. Construire l'image.
6. En production, placer un reverse-proxy/TLS devant l'application. Le `Caddyfile.example` fournit un modèle.
7. Ne jamais exposer directement le port applicatif sur Internet.

## Important
Le correctif ne remplace pas un audit du code métier. Pour une validation finale, il faut le dépôt complet, notamment :
- `go.mod`, `go.sum`
- `cmd/`, `internal/`, `pkg/` ou équivalent
- migrations/schema SQLite
- handlers/API/middleware JWT
- code de génération Factur-X/UBL/CII/EDIFACT
- connecteurs PPF/PDP/AS4
- tests unitaires/intégration/e2e
- configuration CI/CD.
