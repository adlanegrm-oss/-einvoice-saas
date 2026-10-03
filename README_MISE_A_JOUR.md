# Package de Mise à Jour — eInvoice SaaS
**Date** : 02/10/2026  
**Version cible** : 1.1.0 (Socle Sécurisé + Multi-tenant + Conformité renforcée)  
**Périmètre** : Sprints 0 → 3

Ce package contient les livrables d’infrastructure, de configuration, de scripts et de documentation correspondant au plan d’action suivant :

| Sprint | Période       | Objectif principal                          |
|--------|---------------|---------------------------------------------|
| 0      | Semaine 1     | Sécurisation immédiate (réseau + secrets)   |
| 1      | Semaines 2-3  | Sécurité applicative + Multi-tenant         |
| 2      | Semaines 4-5  | Conformité EN 16931 + Formats              |
| 3      | Semaines 6-8  | Industrialisation + Connecteurs + PRA       |

> **Important** : Ce package est un **correctif d’exploitation et de configuration**.  
> Il ne réécrit pas l’intégralité du code Go métier (handlers, pipeline, parsers).  
> Les modifications de code applicatif devront être réalisées dans le dépôt source en suivant le plan détaillé ci-dessous.

---

## Contenu du package

```
einvoice-saas-update/
├── README_MISE_A_JOUR.md          ← ce fichier
├── CHANGELOG.md
├── docker-compose.hardened.yml   ← Compose sans exposition de ports
├── Caddyfile.example             ← Reverse-proxy TLS
├── .env.example.hardened         ← Secrets obligatoires par environnement
├── config/
│   └── security-headers.md       ← Headers recommandés
├── scripts/
│   ├── preflight.sh              ← Contrôles avant déploiement
│   ├── sqlite-backup.sh          ← Sauvegarde cohérente SQLite + intégrité
│   └── rotate-secrets.sh         ← Aide à la rotation des secrets
└── docs/
    ├── PLAN_SPRINTS.md           ← Plan d’action détaillé
    ├── TICKETS_GITHUB.md         ← Tickets prêts à coller dans GitHub Issues
    ├── OPERATIONS_CHECKLIST.md   ← Checklist mise en production
    ├── SECURITY_HARDENING.md     ← Mesures de durcissement
    └── PRA_RUNBOOK.md            ← Procédure de restauration
```

---

## Installation rapide (Sprint 0)

```bash
# 1. Copier les fichiers dans la racine du projet
cp docker-compose.hardened.yml Caddyfile.example .env.example.hardened /chemin/vers/einvoice-saas/
cp -r scripts/ docs/ /chemin/vers/einvoice-saas/

# 2. Créer un .env réel à partir du modèle
cp .env.example.hardened .env
# → Générer les secrets (voir scripts/rotate-secrets.sh)

# 3. Exécuter les contrôles pré-déploiement
chmod +x scripts/*.sh
./scripts/preflight.sh

# 4. Lancer avec le compose durci + reverse-proxy
docker compose -f docker-compose.hardened.yml up -d --build
# Puis configurer Caddy (ou Traefik/nginx) avec Caddyfile.example
```

---

## Ordre recommandé d’application

1. **Sprint 0** (immédiat) → appliquer ce package
2. **Sprint 1** → modifications code (persistance users, JWT strict, isolation tenant)
3. **Sprint 2** → XSD/Schematron + Factur-X complet + audit trail chaîné
4. **Sprint 3** → connecteurs, métriques, PRA, CI renforcée

Voir `docs/PLAN_SPRINTS.md` et `docs/TICKETS_GITHUB.md` pour le détail des tickets GitHub et des fichiers à modifier dans le code source.
