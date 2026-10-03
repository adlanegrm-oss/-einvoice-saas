# Matrice de Capacité & Vérité Technique du Dépôt (Octobre 2026)

> **Positionnement officiel :** Prototype technique avancé / Passerelle d'interconnexion & conformité e-invoicing B2B en cours de qualification. **Non agréé PDP / Non certifié à ce stade.**

### Statut de Maturité des Composants

| Capacité / Composant | Code Source | Tests Automatisés | Statut Réel | Qualification DGFiP / Marché |
| :--- | :---: | :---: | :---: | :--- |
| **Authentification (PBKDF2/JWT)** | ✅ Inclus | ✅ Tests unitaires | `PRODUCTION VERIFIED` | Autonome / multi-tenant |
| **Isolation Multi-Tenant** | ✅ Inclus | ✅ Anti-IDOR | `TESTED` | Contexte strict par `tenant_id` |
| **Parsing UBL 2.1 & CII** | ✅ Inclus | ✅ Golden fixtures | `TESTED` | Détection & extraction typée |
| **Idempotence Transactionnelle** | ✅ Inclus | ✅ Tests de concurrence | `TESTED` | Clé avec lock DB & détection 409 |
| **Machine à États Irréversible** | ✅ Inclus | ✅ Transitions scellées | `TESTED` | Aucun retour en arrière possible |
| **Outbox Asynchrone Durable** | ✅ Inclus | ✅ Retry & DLQ | `TESTED` | Survie aux crashs / redémarrages |
| **Validation Règles EN 16931** | 🟡 Partiel | ✅ 15 règles BR/FR | `SANDBOX VERIFIED` | Moteur interne typé (Ruleset 2026.1) |
| **Validation Schematron ISO** | 🔴 Non branché | ❌ Aucun | `NOT READY` | Évaluation moteur XSLT natif Go |
| **Factur-X PDF/A-3 Réel** | 🟡 Extraction XML | ❌ Pas d'injection A-3 | `VERIFY` | Extraction fonctionnelle, génération A-3 non qualifiée |
| **Connecteur PDP / PPF** | 🟡 Gateway Adapter | ✅ Mock incidents | `SANDBOX VERIFIED` | Mock réseau (Timeout, Duplicate, Rejection) |
| **Connecteur AS4 ebMS3** | 🔴 Prototype | 🟡 Mock TLS | `NOT READY` | Nécessite certificats réels & mTLS |
| **Audit Trail Cryptographique** | ✅ Inclus | ✅ Chaîne Merkle SHA-256 | `TESTED` | Dossier de preuve immuable vérifiable |

### Niveaux de maturité utilisés
* `CODED` : Code écrit mais non couvert de bout en bout.
* `TESTED` : Couvert par tests unitaires et intégration en mémoire.
* `VERIFIED` : Validé sur environnement local complet (BDD + fichiers).
* `SANDBOX VERIFIED` : Testé avec succès contre des mocks réalistes d'API partenaires.
* `PARTNER VERIFIED` : Validé en interopérabilité avec une plateforme tierce réelle.
* `PRODUCTION VERIFIED` : Déployé et validé sous charge opérationnelle.
