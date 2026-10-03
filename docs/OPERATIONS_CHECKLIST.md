# Checklist mise en production

## Avant déploiement
- [ ] Secrets distincts par environnement.
- [ ] Aucun secret réel dans Git.
- [ ] `PUBLIC_BASE_URL_PROD` est HTTPS.
- [ ] Certificat TLS valide.
- [ ] Port applicatif non publié sur Internet.
- [ ] Firewall : seuls 80/443 et accès d'administration nécessaires.
- [ ] Sauvegarde initiale effectuée.
- [ ] Restauration testée sur un environnement isolé.
- [ ] Compte ADMIN configuré et mot de passe temporaire supprimé/rotaté.

## Métier
- [ ] Tests EN 16931 passants.
- [ ] Tests des montants/TVA/arrondis.
- [ ] Tests des statuts et transitions.
- [ ] Test anti-double émission.
- [ ] Test idempotence.
- [ ] Test multi-tenant négatif.
- [ ] Test téléchargement d'un autre tenant refusé.

## Connectique
- [ ] TLS validé.
- [ ] Certificats partenaires validés.
- [ ] Timeouts configurés.
- [ ] Retry borné.
- [ ] Anti-rejeu.
- [ ] Corrélation des messages.
- [ ] Rejeu manuel sécurisé.
- [ ] DLQ ou équivalent opérationnel.

## XML / fichiers
- [ ] XXE/DTD interdits.
- [ ] Taille maximale de fichier.
- [ ] Compression contrôlée.
- [ ] Path traversal testé.
- [ ] Validation contenu réel.
- [ ] Factur-X PDF/A-3 contrôlé.
- [ ] Hash calculé sur l'objet effectivement transmis.

## Observabilité
- [ ] Logs structurés.
- [ ] Pas de JWT, mot de passe, SMTP password ou secret dans les logs.
- [ ] Corrélation request/message/invoice.
- [ ] Alertes sur erreurs 5xx.
- [ ] Alertes sur queue saturée.
- [ ] Alertes sur échec de sauvegarde.
- [ ] Monitoring `/health`.

## PRA
- [ ] RPO défini.
- [ ] RTO défini.
- [ ] Backup SQLite cohérent.
- [ ] Archives sauvegardées.
- [ ] Backup hors machine.
- [ ] Chiffrement backup.
- [ ] Test de restauration périodique.
