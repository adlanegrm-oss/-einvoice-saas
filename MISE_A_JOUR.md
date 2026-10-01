# Mise à jour — corrections de sécurité et de structure

## À faire tout de suite (avant de déployer)

1. **Changer les mots de passe** `hadahowana` et `total2026` partout où ils sont utilisés : ils ont été publiés dans le dépôt public.
2. **Purger l'historique Git** si vous voulez qu'ils disparaissent de GitHub (`git filter-repo` ou BFG), puis forcer le push. Sans cela, les anciennes versions restent consultables.
3. **Renommer le dépôt** en `einvoice-saas` (le tiret initial gêne `git` et `gh`). GitHub redirige l'ancienne adresse ; le module Go (`go.mod`) est déjà au bon nom.
4. Renseigner les secrets (voir `.env.example`) : `JWT_SECRET` et `ADMIN_PASSWORD` sont désormais obligatoires hors `DEV`.

## Appliquer la mise à jour

```bash
# depuis la racine de votre clone, branche dédiée
git checkout -b correctifs-securite
# supprimer les anciens fichiers du dépôt puis copier le contenu du ZIP par-dessus
git rm -r --cached . -q && git clean -fdx -q     # attention : supprime aussi les fichiers non suivis
# (décompresser le ZIP à la racine)
git add -A && git commit -m "Sécurité, structure et tests"
go vet ./... && go test -race ./...
```

Le ZIP a été préparé **sans compilateur Go** : lancez `go vet ./... && go test ./...` avant toute mise en production et corrigez les éventuelles erreurs de compilation restantes.

## Ce qui a changé

**Sécurité**
- Comptes et mots de passe sortis du code ; hachage PBKDF2, jetons JWT signés (`internal/auth`).
- Toutes les routes de données exigent un jeton ; l'en-tête `X-User-Role` n'est plus jamais cru.
- Réinitialisation de mot de passe : jeton à usage unique, jamais renvoyé par l'API, envoyé par e-mail (SMTP) ; réponse identique que le compte existe ou non ; page `reset-password.html` ajoutée (elle n'existait pas).
- Limitation de débit (connexion, mot de passe oublié) ; en-têtes de sécurité.
- Archives isolées par client (`<tenant>/factures|temporaire`) ; l'administrateur voit tout.
- Dépôt : validation réelle du contenu (avant, le message « validé » s'affichait sans contrôle), extensions blanchies, tailles bornées, noms uniques (plus d'écrasement), SHA-256 journalisé.
- Front : plus de mots de passe pré-remplis, échappement HTML (XSS), jeton joint aux requêtes, téléchargement via `fetch`.

**Correctifs fonctionnels**
- Le badge « prêt à envoyer » ne s'affichait jamais (comparaison avec `PRET` sur un texte contenant `PRÊT`).
- Les brouillons « 72 h » n'étaient jamais supprimés : purge horaire ajoutée.
- `PORT`, `DB_NAME`, `APP_ENV` étaient ignorés (port `:8080` codé en dur) : trois environnements possibles.
- Dockerfile en Go 1.22 alors que `go.mod` exige 1.24 : aligné ; conteneur sans root ; volumes nommés au lieu de fichiers `.db` montés (Docker les transformait en dossiers).
- `swagger.yaml` n'était pas servi (hors de `web/`) : déplacé et réécrit.
- Validateur : XML lu en entier (un fichier tronqué après la racine passait), DOCTYPE refusé, EDIFACT vérifié segment par segment (UNH/UNT/UNZ, compteur), PDF : détection de signature corrigée (`/Contents` existe dans tout PDF).
- Factures : numéros uniques par émetteur, identifiant attribué par le serveur, TVA bornée, arrondi au centime, rapport journalier indépendant du format de date SQLite.
- Pool de tâches : plus de panique après `Stop()`, un job qui plante ne tue plus le worker.
- `clearance` ne simule plus un succès « CLEARED » par défaut.

**Hygiène**
- ~85 fichiers vides à la racine supprimés (`func`, `import`, `.gitignoreid`…), ainsi que `inject.go`.
- `archives/` et `test-results/` retirés du dépôt et ajoutés au `.gitignore` ; le PDF d'exemple est dans `testdata/`.
- `internal/clearance` (fichier sans extension) et `sqlite_test.go` (caractère invisible dans le nom) : renommés, donc enfin compilés et testés.
- Deux tables `invoices` en conflit : le registre historique utilise désormais `registry_*`.
- Docs (`README`, `MODE_D_EMPLOI`, `DOCUMENT_D_EXPLOITATION`) réécrites : suppression du texte d'assistant et des fonctions annoncées mais absentes.
- CI : `go vet`, `go test -race`, `govulncheck`.

## Constats à connaître

- `testdata/facture_test_hybride.pdf` est un PDF 1.4 **sans XML embarqué** : ce n'est pas un vrai Factur-X.
- Les comptes sont en mémoire : ils sont recréés au démarrage depuis la configuration (un mot de passe changé par réinitialisation est perdu au redémarrage).
- Le XML Factur-X exporté reste incomplet (pas de vendeur, ni de devise, ni de montant à payer).
- Les tests Playwright (`test/buttons.spec.ts`) visent l'ancienne interface et ne passent plus tels quels : la page d'accueil redirige vers la connexion.
- Non branchés : `internal/parser`, `internal/service`, `internal/jobs`.

Feuille de route : voir `README.md`.

## Seconde passe de revue (relecture complète, sans compilateur)

Corrigé :
- **Rapport journalier** : la date par défaut utilisait l'heure locale alors que la base regroupe par jour UTC ; le rapport du soir pouvait tomber sur le mauvais jour. Désormais UTC partout.
- **Enregistrement d'une facture** : la date insérée était `inv.IssueDate` et non la date effective calculée ; lecture tolérante aux dates NULL (bases anciennes).
- **Validation UBL / Peppol** : un UBL conforme EN 16931 n'ayant que `PartyLegalEntity/RegistrationName` (sans `PartyName`) était rejeté. Les deux sont maintenant lus (+ test).
- **`index.html`** : le lien « Télécharger » visait une route inexistante (`/api/v1/archives/download`) et ne pouvait de toute façon pas envoyer le jeton ; remplacé par un téléchargement authentifié. Valeurs du tableau échappées (XSS). La supervision interrogeait `http://localhost:<port>/api/v1/health` en dur : elle utilise maintenant l'hôte de la page et `/health`.
- **`login.html`** : les e-mails en dur (`admin.super@einvoice.int`, `tresorerie@totalenergies.com`) du « mot de passe oublié » sont remplacés par l'e-mail saisi ; `#` parasite en fin de fichier supprimé.
- **`.dockerignore`** ajouté (absent : `COPY . .` embarquait `.env`, bases et archives dans le contexte de build).
- `sharepoint.md` : noms de tables alignés (`registry_*`). Mise en forme gofmt de quelques structures.

Limites connues (non traitées) :
- Aucun `go build` / `go test` n'a pu être exécuté (pas de compilateur Go dans l'environnement de revue) : la relecture est manuelle. **Lancez `go vet ./... && go test -race ./...` avant tout déploiement.**
- XML encodé en ISO-8859-1 : refusé (le décodeur standard n'a pas de convertisseur de jeu de caractères).
- Détection du Factur-X dans un PDF par recherche du nom de pièce jointe : ne voit pas un nom placé dans un flux compressé.
- Les secrets `recette` / `preprod` / `prod` sont tous exigés par `docker compose` même pour lancer un seul service (interpolation globale du fichier).

## Jalons validés - Cycle E2E complet (01/10/2026)

- **Handlers & Compilation** : Résolution du doublon `writeJSON` et correction syntaxique dans `internal/handler/task_handler.go`.
- **Authentification & Session** :
  - `POST /api/v1/auth/login` : émission du token JWT avec claims RBAC (`role: ADMIN`, `tenant`).
  - `GET /api/v1/auth/me` : validation du contexte connecté et du middleware Bearer.
- **Gestion des factures (Core API)** :
  - `POST /api/v1/invoices` : validation des règles métier (`number`, `customer`, `items`, TVA 0-100%).
  - Calcul automatique et arrondi au centime (`CalculateTotals`) des montants HT, TVA et TTC.
  - `GET /api/v1/invoices` : persistance vérifiée sous SQLite WAL.
  - `GET /api/v1/reports/daily` : agrégation journalière synchrone opérationnelle.
- **Conformité & Formats** :
  - `GET /api/v1/invoices/export?id=...` : export XML conforme Factur-X / CII (`CrossIndustryInvoice`).
- **Archivage & GED** :
  - `POST /api/v1/invoices/deposit` : ingestion multipart via le champ `files`, rejet des schémas invalides (`422`), contrôle de conformité et calcul d'empreinte SHA-256 scellée.
  - `GET /api/v1/invoices/list` : listing isolé par tenant (`VALIDE_PRET_A_ENVOYER`).
  - `GET /api/v1/invoices/download?folder=factures&file=...` : restitution sécurisée des pièces archivées (`200 OK`, `1452 octets`).
- **Traitement asynchrone** :
  - `POST /api/v1/jobs/daily-report` : pool de workers en arrière-plan réceptif (`accepted`).
