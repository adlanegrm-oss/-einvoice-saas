# COMPLIANCE CARE and VALIDATION PLATFORM (CCandVP)

Solution SaaS modulaire conçue pour la conformité réglementaire de facturation électronique (EN 16931-1:2017 & profil CIUS-FR v2.0), avec multi-tenancy, matrice RBAC granulaire par scope géographique, et traçabilité immuable (RFC 3161).

---

## 🏛️ Architecture & Périmètres Métier

### 1. Espaces & Rôles Applicatifs (data-role)
* **Front Client (/app)** : Facturation, dépôt de lots, validation Schematron en temps réel et tiroir de preuves cryptographiques.
* **Front Partenaire (/partner)** : Console multi-tenants pour intégrateurs, cabinets comptables et ERP, supervision d'appels API et webhooks.
* **Console Admin (/admin)** : Gouvernance plateforme, gestion des habilitations utilisateurs et audit de conformité.
* **Console Exploitation (/ops)** : Supervision infrastructure, métriques temps réel, latence de la file Outbox et rejeu DLQ.
* **Console MCO (/mco)** : Runbooks d'astreinte N2/N3 et gestion du PRA.
* **Console R&D (/rnd)** : Banc de test Schematron et comparateur Golden Files.
* **Console Déploiement (/deploy)** : Supervision des environnements CI/CD (DEV, RECETTE, PREPROD, PROD) avec procédure de rollback sécurisée.

### 2. Gouvernance & Moteur de Permissions (policyEngine)
L'accès aux ressources est évalué dynamiquement selon trois scopes :
* **ALL_ORGANIZATION** : Visibilité totale sur l'entité légale (Admin, Direction Financière).
* **OWN_ESTABLISHMENT** : Cloisonnement strict au niveau de la succursale (Comptable).
* **OWN_DOCUMENTS** : Restriction aux seuls documents créés par l'utilisateur (Commercial).

### 3. Normes & Spécifications
* **EN 16931-1:2017** : Norme sémantique européenne pour la facturation électronique.
* **CIUS-FR v2.0** : Spécifications nationales françaises (UBL 2.1, CII D16B, Factur-X / PDF/A-3).
* **RFC 3161** : Horodatage électronique qualifié pour archivage à valeur probante.

