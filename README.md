# eInvoice SaaS

Plateforme SaaS de **validation et d'archivage de factures électroniques** (Factur-X / UBL / EDIFACT / PDF), écrite en Go (bibliothèque standard + SQLite sans CGO).

> **Statut : prototype.** L'authentification, l'isolation par client, l'archivage validé et l'API de factures fonctionnent. La validation normative complète (XSD / Schematron EN 16931, PEPPOL BIS), la génération de vrais Factur-X PDF/A-3 et la transmission à une PDP/PPF ne sont **pas encore** implémentées (voir [Feuille de route](#feuille-de-route)).

## Démarrage rapide (développement)

Prérequis : Go 1.24+.

```bash
go run ./cmd/server
```

En `APP_ENV=DEV` (défaut), le serveur génère un secret de session et un mot de passe administrateur temporaires, affichés **une seule fois** dans les journaux. Pour des identifiants stables :

```bash
export ADMIN_EMAIL=admin@example.com
export ADMIN_PASSWORD='un-mot-de-passe-solide'
export CLIENT_EMAIL=client@example.com        # optionnel : compte client de test
export CLIENT_PASSWORD='un-autre-mot-de-passe'
go run ./cmd/server
```

Puis ouvrez <http://localhost:8080/login.html>. Documentation d'API : <http://localhost:8080/docs.html>.

## Configuration (variables d'environnement)

| Variable | Défaut | Rôle |
|---|---|---|
| `APP_ENV` | `DEV` | `DEV`, `RECETTE`, `PREPROD` ou `PROD`. Hors `DEV`, les secrets sont **obligatoires**. |
| `PORT` | `8080` | Port d'écoute. |
| `DATA_DIR` | `data` | Dossier de la base SQLite et des archives. |
| `DB_NAME` | `invoices.db` | Nom du fichier de base (dans `DATA_DIR`). |
| `ARCHIVE_DIR` | `$DATA_DIR/archives` | Archives : `<tenant>/factures`, `<tenant>/temporaire`. |
| `JWT_SECRET` | — | Clé de signature des jetons, **32 caractères minimum** (obligatoire hors DEV). |
| `TOKEN_TTL` | `8h` | Durée de vie d'une session. |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | `admin@example.com` / — | Compte administrateur (mot de passe ≥ 10 caractères, obligatoire hors DEV). |
| `CLIENT_EMAIL` / `CLIENT_PASSWORD` | — | Compte client optionnel. |
| `PUBLIC_BASE_URL` | `http://localhost:$PORT` | Base des liens de réinitialisation envoyés par e-mail. |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` | — | Envoi des e-mails de réinitialisation. Sans SMTP hors DEV, la réinitialisation est désactivée. |

Modèle prêt à remplir : [`.env.example`](.env.example).

## Docker

```bash
cp .env.example .env      # renseigner tous les secrets
docker compose up -d --build
```

Trois environnements isolés (recette `:8081`, pré-production `:8082`, production `:8080`), chacun avec son volume de données et ses secrets. L'image tourne sans droits root et expose un `HEALTHCHECK` sur `/health`.

> Compose contrôle toutes les variables du fichier : `.env` doit définir les secrets des trois environnements même si vous n'en lancez qu'un.

## Sécurité

- Mots de passe hachés (PBKDF2-HMAC-SHA256, 600 000 itérations, sel aléatoire) ; **aucun identifiant dans le code**.
- Jetons de session JWT HS256 signés côté serveur ; les en-têtes de rôle envoyés par le client sont ignorés.
- Réinitialisation de mot de passe : jeton à usage unique de 15 minutes, transmis **uniquement par e-mail**, réponse identique que le compte existe ou non.
- Limitation de débit sur la connexion et la réinitialisation.
- Isolation par client : chacun n'accède qu'à ses archives et à ses factures ; l'administrateur voit tout.
- Dépôt de fichiers : extensions blanchies (`.pdf`, `.xml`, `.edi`), 10 Mo par fichier, contrôle de contenu, noms nettoyés et uniques, empreinte SHA-256 journalisée, téléchargement forcé (`attachment`, `nosniff`).

**Si un ancien mot de passe de ce projet a été publié** (il l'était dans l'historique Git), considérez-le comme compromis partout où il est réutilisé.

## Tests

```bash
go vet ./...
go test -race ./...
```

La CI (`.github/workflows/ci.yml`) exécute vet, build et tests à chaque push.

## Architecture

```
cmd/server            point d'entrée : configuration, routes, arrêt propre
internal/auth         mots de passe, JWT, comptes, réinitialisation, limiteur
internal/config       lecture et contrôle des variables d'environnement
internal/middleware   authentification, rôles, limitation de débit, en-têtes
internal/handler      API HTTP (auth, archives, factures, validation)
internal/validator    détection et contrôle Factur-X / UBL / EDIFACT / PDF
internal/invoice      modèle de facture et calcul des totaux (arrondi au centime)
internal/repository   SQLite (factures par client, rapports) + registre historique
internal/exporter     export XML CII (profil MINIMUM, incomplet)
internal/worker       pool de tâches asynchrones
internal/clearance    connecteur de clearance (non configuré : refuse par défaut)
web/                  portail (login, console admin, espace client, docs API)
```

Non branchés pour l'instant : `internal/parser/*`, `internal/service`, `internal/jobs`.

## Feuille de route

1. Persister les comptes en base (aujourd'hui recréés au démarrage depuis la configuration).
2. Validation XSD + Schematron (EN 16931, Factur-X, PEPPOL BIS) ; extraction du XML embarqué des PDF.
3. Export Factur-X complet : vendeur (SIREN/TVA), devise, montant à payer, intégration dans un PDF/A-3.
4. Montants en entiers (centimes) plutôt qu'en `float64`.
5. Connecteur PDP/PPF et PEPPOL réels derrière `internal/clearance`.
6. Politique CSP stricte (déplacer les scripts inline hors des pages HTML).

## Licence

MIT — voir [LICENSE](LICENSE).
