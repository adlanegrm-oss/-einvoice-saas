# Runbook PRA — eInvoice SaaS

## Objectifs

| Indicateur | Valeur cible (recommandée) |
|------------|---------------------------|
| RPO        | ≤ 1 heure                 |
| RTO        | ≤ 30 minutes              |
| Rétention  | 14 jours locaux + 90 jours hors-site |

## Sauvegarde quotidienne

```bash
# Sur l'hôte ou via cron dans le conteneur (si volume monté)
./scripts/sqlite-backup.sh /data/invoices.db /backups/prod
# + rsync / copie chiffrée hors machine
```

- Fréquence : toutes les heures (ou après pic d'activité).
- Inclure : DB SQLite + répertoire `archives/` (ou volume complet `/data`).
- Vérifier `PRAGMA integrity_check` après chaque backup.
- Chiffrer les backups hors-site (gpg / age / S3 SSE).

## Restauration

1. Arrêter le service :
   ```bash
   docker compose -f docker-compose.hardened.yml stop einvoice-prod
   ```
2. Restaurer le fichier DB :
   ```bash
   gunzip -c backups/invoices_YYYYMMDDTHHMMSSZ.db.gz > /data/invoices.db
   # ou sqlite3 /data/invoices.db ".restore 'backup.db'"
   ```
3. Restaurer le volume archives si nécessaire.
4. Vérifier intégrité :
   ```bash
   sqlite3 /data/invoices.db "PRAGMA integrity_check;"
   ```
5. Redémarrer :
   ```bash
   docker compose -f docker-compose.hardened.yml start einvoice-prod
   ```
6. Contrôler `/health` + login + liste factures.

## Tests périodiques

- Restauration complète sur environnement isolé (recette) **au moins 1× / mois**.
- Documenter le temps réel (RTO mesuré).
- Alerter si le backup échoue (cron + monitoring).

## Points d'attention

- Sous WAL, utiliser toujours `.backup` (pas de simple `cp` sous charge).
- Restaurer DB **et** archives ensemble (cohérence documents ↔ métadonnées).
- Après restauration : invalider éventuellement les sessions JWT (rotation secret si compromission).
