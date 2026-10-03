# Changelog

## Release: Fiabilisation du socle : calcul exact, TVA par catégorie, cycle de vie unifié

### Bloquants corrigés
1. **Moteur monétaire unique :**
   - Remplacement de float64 par un entier int64 (centimes) et arrondi bancaire Half-Even (IEEE 754).
   - Suppression des marges de tolérance arbitraires (0.05€ et 0.02€) sur BR-CO-13 et BR-CO-15.
2. **Modèle de TVA par catégorie (BG-23) :**
   - Introduction du type `TaxPercent` (points de base) distinct de `Money`.
   - Support des catégories S, Z, E, AE, K, G, O avec motif d'exonération/autoliquidation (BR-E-01, BR-AE-01).
3. **Codes de règles et transparence :**
   - Réalignement des diagnostics sur la nomenclature officielle CEN/TC 434 (BR-02, BR-05, BR-CO-13, BR-CO-15, BR-CO-17).
   - Injection du ruleset_version et coverage_scope dans le rapport de validation.
4. **Machine d'états unifiée :**
   - Cycle de vie français unique : BROUILLON -> DEPOSEE -> [REJETEE | REFUSEE | ENCAISSEE].
   - Élimination de l'état fantôme 'CLOSED' dans les batchs de purge au profit des états terminaux réels.
5. **Nettoyage documentaire :**
   - Rétrogradation de l'étiquette 'PRODUCTION VERIFIED' à 'IN DEVELOPMENT / PRE-ALPHA'.
   - Nettoyage des références résiduelles et documents obsolètes.
