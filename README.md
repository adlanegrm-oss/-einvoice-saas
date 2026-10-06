# 🚀 Implémentation & évolution de la plateforme

Cette journée a été consacrée à une phase importante de structuration, de durcissement et d'industrialisation de la plateforme SaaS d'e-invoicing. Les travaux ont porté à la fois sur l'évolution fonctionnelle du moteur de facturation électronique, la sécurisation de l'application, la préparation de l'exploitation et la formalisation de la roadmap technique.

---

## Phase 1 — Évolution fonctionnelle de la facturation électronique

### Multi-format e-invoicing

Le moteur de facturation électronique a été étendu afin de préparer la prise en charge de plusieurs standards et formats de facturation électronique.

Les principaux travaux réalisés comprennent :

* ajout d'une architecture de traitement multi-format ;
* séparation des modèles métier et des formats d'échange ;
* intégration de modèles monétaires dédiés ;
* ajout de parseurs et générateurs spécialisés ;
* prise en charge des formats européens et nationaux ciblés ;
* amélioration de la validation des documents ;
* centralisation du traitement dans un service multi-format ;
* préparation de l'extension future vers de nouveaux standards.

### Formats et standards ciblés

L'architecture actuelle prépare notamment l'intégration de :

* **Facturae** ;
* **FatturaPA** ;
* **KSeF** ;
* **UBL / CII** dans le cadre des évolutions réglementaires ;
* **EN 16931** et profils nationaux à intégrer progressivement.

Cette architecture permet de faire évoluer le moteur sans coupler le domaine métier à un format de facture particulier.

---

## Phase 2 — Restructuration et documentation du projet

Une phase de restructuration documentaire a également été réalisée afin de rendre le projet plus lisible et maintenable.

Travaux réalisés :

* mise à jour du `README.md` ;
* formalisation de l'architecture cible ;
* documentation des objectifs de conformité ;
* création d'une roadmap d'évolution ;
* définition des différents sprints techniques ;
* formalisation des priorités de sécurité, conformité et industrialisation ;
* ajout de tickets et axes d'implémentation destinés au suivi GitHub.

La roadmap est désormais organisée autour de quatre grands axes :

1. **Socle et durcissement de la plateforme**
2. **Sécurité, authentification et multi-tenancy**
3. **Conformité et industrialisation du moteur e-invoicing**
4. **Connectivité, observabilité et exploitation industrielle**

---

## Phase 3 — Durcissement de l'infrastructure

Une première couche de sécurisation et d'industrialisation de l'environnement d'exécution a été intégrée.

### Infrastructure

Les éléments suivants ont été ajoutés ou préparés :

* configuration Docker renforcée ;
* `docker-compose.hardened.yml` ;
* configuration Caddy pour la terminaison TLS ;
* fichiers de configuration dédiés aux environnements sécurisés ;
* scripts d'exploitation et de pré-vérification ;
* préparation de la gestion des secrets ;
* documentation des procédures d'exploitation.

L'objectif est de disposer d'un socle permettant de passer progressivement d'un environnement de développement vers une architecture exploitable en recette, préproduction puis production.

---

## Phase 4 — Sécurisation de l'application Web

Une première phase de sécurisation applicative a été réalisée sur le serveur et les interfaces Web.

Travaux effectués :

* correction du routage statique ;
* sécurisation de la logique d'authentification côté Web ;
* amélioration des mécanismes de redirection ;
* préparation du cloisonnement multi-tenant ;
* amélioration de la structure de configuration ;
* préparation du serveur HTTP à une configuration de production.

La configuration du serveur est également en cours de durcissement avec la mise en place de timeouts explicites pour les lectures, écritures et connexions inactives.

> **État :** le durcissement du serveur HTTP doit encore être validé par `gofmt`, les tests et `go vet` avant d'être considéré comme terminé.

---

## Phase 5 — Préparation du multi-tenancy

Le projet a été structuré pour évoluer vers un véritable modèle SaaS multi-tenant.

Les travaux réalisés ont permis de préparer :

* l'identification des organisations clientes ;
* le cloisonnement des données ;
* la séparation des espaces applicatifs ;
* la gestion future des utilisateurs par organisation ;
* l'application de politiques d'accès selon le contexte utilisateur.

La prochaine étape consiste à remplacer progressivement le stockage utilisateur actuellement en mémoire par une persistance SQLite complète et à appliquer systématiquement les contrôles d'isolation sur les endpoints métier.

---

## Phase 6 — Sécurité de l'authentification

L'architecture d'authentification a été revue dans une logique de durcissement progressif.

Les objectifs désormais formalisés comprennent :

* persistance des utilisateurs en base ;
* stockage sécurisé des mots de passe ;
* PBKDF2 pour le dérivé de mot de passe ;
* validation complète des JWT ;
* contrôle de `iss`, `aud`, `exp` et `nbf` ;
* gestion des rôles et scopes ;
* limitation des tentatives d'authentification ;
* sécurisation des sessions et des secrets.

Une partie de ces éléments constitue désormais le périmètre prioritaire du **Sprint 1**.

---

## Phase 7 — Préparation de l'exploitation et du MCO

La plateforme a également commencé à être structurée pour supporter une exploitation durable.

Les travaux réalisés ou préparés couvrent :

* scripts d'exploitation ;
* configuration d'environnement ;
* documentation de déploiement ;
* procédures de vérification ;
* préparation des environnements DEV / RECETTE / PREPROD / PROD ;
* préparation des mécanismes de rollback ;
* préparation des procédures de sauvegarde et de reprise ;
* formalisation des besoins d'observabilité.

L'objectif est de réduire progressivement la dépendance aux opérations manuelles et de rendre les déploiements reproductibles.

---

## Phase 8 — Roadmap de conformité et d'industrialisation

Une roadmap technique complète a été ajoutée afin de structurer les prochaines évolutions.

### Sprint 0 — Socle

* durcissement Docker ;
* TLS / reverse proxy ;
* gestion sécurisée des secrets ;
* scripts de pré-vérification ;
* documentation ;
* préparation CI/CD.

### Sprint 1 — Sécurité & SaaS

* persistance des utilisateurs ;
* sécurisation des mots de passe ;
* validation complète des JWT ;
* isolation multi-tenant ;
* RBAC et scopes ;
* rate limiting ;
* sécurisation des uploads ;
* protections HTTP.

### Sprint 2 — Conformité e-invoicing

* validation XSD ;
* validation Schematron ;
* EN 16931 ;
* PEPPOL BIS ;
* profils nationaux ;
* Factur-X / PDF-A-3 ;
* pipeline de validation unifié ;
* idempotence ;
* journal d'audit et traçabilité.

### Sprint 3 — Industrialisation

* connecteurs d'échange ;
* AS4 / PDP ;
* observabilité ;
* métriques Prometheus ;
* logs structurés ;
* sauvegarde et PRA ;
* CI/CD renforcée ;
* analyse des vulnérabilités ;
* SBOM ;
* tests de charge ;
* documentation d'exploitation.

---

## Phase 9 — État actuel

La plateforme dispose désormais d'un socle fonctionnel multi-format et d'une première couche d'industrialisation et de sécurisation.

### ✅ Réalisé

* architecture e-invoicing multi-format ;
* prise en charge de plusieurs formats spécialisés ;
* modèles métier monétaires ;
* amélioration de la validation ;
* documentation principale ;
* roadmap technique ;
* configuration Docker renforcée ;
* configuration TLS / reverse proxy ;
* scripts d'exploitation ;
* première sécurisation du Web ;
* préparation du multi-tenancy ;
* structuration des futurs travaux de conformité et d'industrialisation.

### 🔄 En cours

* persistance des utilisateurs ;
* RBAC complet ;
* isolation multi-tenant systématique ;
* durcissement JWT ;
* protection des uploads ;
* sécurisation complète du serveur HTTP ;
* validation réglementaire XSD / Schematron ;
* audit trail complet ;
* observabilité et CI/CD industrielle.

### 🎯 Objectif cible

L'objectif final est de faire évoluer le projet vers une plateforme SaaS d'e-invoicing industrialisée, sécurisée et extensible, capable de gérer plusieurs organisations, plusieurs formats de facturation électronique et plusieurs niveaux de conformité, tout en assurant la traçabilité, l'intégrité des documents et l'exploitation à grande échelle.

---

## 📌 Traçabilité des évolutions

Les évolutions majeures réalisées au cours de cette phase sont historisées dans Git et organisées par commits fonctionnels afin de conserver une traçabilité claire entre :

**fonctionnalité → sécurisation → documentation → industrialisation → conformité → exploitation.**

