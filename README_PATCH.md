# Patch Correctif & Évolution de Conformité (v2)

Ce patch implémente les briques critiques identifiées pour la transition d'un prototype vers un moteur transactionnel conforme :

1. **Précision Financière Sans `float64` (`internal/currency/money.go`)** :
   - Calculs décimaux à virgule fixe et arrondis bancaires stricts (Half-Even/Half-Up EN 16931).
   - Suppression des tolérances arbitraires (0.05 / 0.02) au profit d'une vérification exacte des sommes (`BR-CO-10`, `BR-CO-13`).

2. **Chaîne de Preuve d'Audit Cryptographique (`internal/audit/chain.go`)** :
   - Implémentation du Linked Log `H_n = SHA256(canonical_payload_n || H_{n-1} || sequence)`.
   - Fonction `VerifyAuditChain(events)` permettant de détecter toute altération manuelle directe en base.

3. **Schéma SQL Transactionnel (`migrations/002_compliance_core.sql`)** :
   - Table `idempotency_keys` avec verrou multi-instances et expiration TTL.
   - Table `audit_events` strictement typée et chaînée.
   - Table `transactional_outbox` pour l'émission asynchrone AS4/PDP avec gestion des retries et DLQ.
   - Stockage du document brut opposable (`raw_payload`) et son hash SHA-256 (`payload_hash`).

4. **Service d'Émission Atomique (`internal/service/invoice_emitter.go`)** :
   - Transaction unique `BEGIN ... COMMIT` regroupant : Idempotence + Persistance Facture + Bloc d'audit `ISSUED` + Tâche Outbox.

5. **Worker Outbox Asynchrone (`cmd/outbox_worker/main.go`)** :
   - Consommation concurrente via `FOR UPDATE SKIP LOCKED`.
   - Émission externe, retry exponentiel et scellement automatique du bloc d'audit `TRANSMITTED`.