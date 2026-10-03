# Notes du correctif

## Modifications infrastructure
1. Suppression de l'exposition directe des ports applicatifs dans le compose durci.
2. Ajout d'un réseau interne.
3. Ajout de `cap_drop: ALL`.
4. Ajout de `init: true`.
5. Ajout de limites CPU/mémoire/PID.
6. Conservation de `read_only`, tmpfs et `no-new-privileges`.
7. Variables de production rendues explicitement obligatoires dans l'exemple durci.
8. Modèle de reverse proxy TLS fourni.
9. Script de préflight.
10. Script de backup SQLite avec `PRAGMA integrity_check`.

## Ce qui n'est volontairement pas "inventé"
Le code Go n'étant pas fourni, aucun handler, middleware JWT, schéma SQL ou connecteur AS4/PPF/PDP n'a été réécrit artificiellement.

## Attention
`docker-compose.hardened.yml` est un modèle de déploiement durci. Il suppose que le binaire écoute sur `0.0.0.0:8080` dans le réseau Docker et que le reverse proxy lui transmet les requêtes.
