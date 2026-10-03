# État d'Avancement du Projet & Jalons Techniques (Octobre 2026)

## 📊 Résumé de la Couverture des Tests
* **Statut global :** 100 % des tests unitaires, d'intégration et de non-régression validés.
* **Détail d'exécution :** Voir le journal complet dans [`TEST_REPORT.md`](TEST_REPORT.md).

---

## 🎯 Avancement des Chantiers Prioritaires (P0)

| Chantier / Fonctionnalité | État | Détails techniques |
| :--- | :---: | :--- |
| **Modèle Monétaire Exact** | ✅ Terminé | Remplacement de `float64` par `domain.Amount` (centimes), arrondi bancaire Half-Even (EN 16931). |
| **Isolation Multi-Tenant** | ✅ Terminé | Propagation du `tenant_id` via `context.Context`, contrainte `(tenant_id, invoice_number)`. |
| **Idempotence Persistante** | ✅ Terminé | Stockage SQL persistant, validation SHA-256 du payload, détection de conflit `409`. |
| **Machine d'États Découplée** | ✅ Terminé | Séparation orthogonale : Conformité, Acheminement (Transport) et Paiement. |
| **Transactional Outbox & DLQ** | ✅ Terminé | Émission atomique DB (Facture + Audit + Outbox), dépilage asynchrone avec backoff exponentiel. |
| **Chaîne d'Audit Cryptographique** | ✅ Terminé | Enchaînement scellé SHA-256 (`prev_event_hash`, `payload_hash`, `event_hash`). |
| **Moteur EN 16931 & CIUS-FR** | ✅ Terminé | Diagnostics structurés (`BR-xx`, balises `BT-xx`, sévérité `ERROR`/`WARNING`). |
| **Schéma SQL PostgreSQL** | ✅ Prêt | Migration complète (`migrations/000001_p0_infrastructure_core.up.sql`). |
| **Matrice de Capacités** | ✅ Publié | Définition claire du statut "Solution Compatible" dans [`CAPABILITY_MATRIX.md`](CAPABILITY_MATRIX.md). |

---

## 🗓️ Feuille de Route Immédiate (Sprint B / P1)
- [ ] Exposition du point d'entrée `POST /v1/validate` en accès public/sandbox pour audit ERP.
- [ ] Système de dispatch des webhooks sortants signés (HMAC-SHA256).
- [ ] Connecteur direct vers l'annuaire d'adressage et récepteur PDP partenaire.
