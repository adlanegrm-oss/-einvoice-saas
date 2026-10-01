# Mode d'emploi

## 1. Espace client

1. Ouvrez `/login.html` et connectez-vous avec le compte fourni par l'administrateur.
2. **Déposer des factures** : bouton de dépôt du tableau de bord, jusqu'à 50 fichiers (`.pdf`, `.xml`, `.edi`, 10 Mo maximum chacun).
   - **Valider (prêt à envoyer)** : chaque fichier est contrôlé (format, structure, champs essentiels). Un fichier invalide est **rejeté** avec le détail des erreurs, il n'est pas archivé.
   - **Brouillon (72 h)** : le fichier est conservé sans exigence de validité, puis supprimé automatiquement au bout de 72 heures.
3. **Consulter les factures** : liste de vos archives avec leur statut ; « Ouvrir / Télécharger » récupère le document (les PDF s'ouvrent dans le navigateur, les XML et EDI sont téléchargés).
4. **Mot de passe oublié** : depuis la page de connexion, saisissez votre e-mail ; si un compte correspond, un lien valable 15 minutes est envoyé. Le message affiché est le même que le compte existe ou non.

Vous ne voyez jamais les documents des autres clients.

## 2. Console d'administration

Réservée au rôle `ADMIN` : elle voit les archives de tous les clients et peut déclencher le rapport global (`POST /api/v1/jobs/daily-report`).

## 3. Utiliser l'API

Toutes les routes, sauf `/health`, `/api/v1/health` et `/api/v1/auth/*`, exigent un jeton `Bearer`.

```bash
# 1. Se connecter
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"client@example.com","password":"…"}' | jq -r .token)

# 2. Valider un fichier sans l'enregistrer
curl -X POST http://localhost:8080/api/v1/validate \
  -H "Authorization: Bearer $TOKEN" --data-binary @facture.xml

# 3. Enregistrer une facture structurée
curl -X POST http://localhost:8080/api/v1/invoices \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"number":"FAC-2026-001","customer":"ACME","items":[{"description":"Prestation","quantity":2,"unit_price":100,"vat_rate":20}]}'

# 4. Rapport du jour, export Factur-X d'une facture
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/reports/daily?date=2026-09-24"
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/invoices/export?id=inv-…"
```

Contrat complet : `/docs.html` (Swagger UI) ou `/swagger.yaml`.

## 4. Journaux

Les journaux sont émis en JSON sur la sortie standard (`docker compose logs -f`). Chaque dépôt est aussi consigné dans `archives/<tenant>/traitement.log` avec l'empreinte SHA-256 du fichier.
