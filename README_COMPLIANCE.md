# EN 16931 & UN/CEFACT D16B Compliance Implementation

Cette implémentation concrétise l'intégralité des directives du ticket :
`feat(validation): integrate official EN16931 Schematron validation`.

## Composants intégrés

1. **Générateur CII UN/CEFACT D16B (`pkg/syntax/cii_generator.go`)**
   - Implémente la structure standard complète `rsm:CrossIndustryInvoice`.
   - Vendeur (`SellerTradeParty`) et Acheteur (`BuyerTradeParty`) avec SIREN/SIRET, TVA, adresses et identifiants électroniques.
   - Lignes de facture (`IncludedSupplyChainTradeLineItem`) avec quantités, codes unités UN/ECE Rec 20, prix nets, ventilation TVA par ligne.
   - Totaux monétaires complets (`LineTotalAmount`, `TaxBasisTotalAmount`, `TaxTotalAmount`, `GrandTotalAmount`, `DuePayableAmount`).
   - Échéance et conditions de règlement (`ApplicableHeaderTradeSettlement`).

2. **Moteur Normatif Schematron EN 16931 & CIUS-FR (`pkg/validation/schematron.go`)**
   - Validation en flux direct du XML généré.
   - Retourne les identifiants de règles officiels :
     - `BR-01` à `BR-09` : Balises d'entête obligatoires, parties, pays.
     - `BR-16` : Présence obligatoire de lignes de facturation.
     - `BR-21` : Identifiant de ligne.
     - `BR-CO-10` : Cohérence somme des lignes = total net HT.
     - `BR-CO-13` : Somme des bases TVA = assiette totale HT.
     - `BR-CO-14` : Somme des montants TVA = montant total TVA.
     - `BR-CO-15` : Total TTC = Total HT + Total TVA.
     - `BR-CO-16` : Reste à payer <= Total TTC.
     - `BR-CO-25` : Date d'échéance postérieure ou égale à la date d'émission.
     - `CIUS-FR-01` : Contrôle strict SIREN (9 chiffres) ou SIRET (14 chiffres) pour les entités françaises.
     - `CIUS-FR-02` : Format regex TVA intracommunautaire française (`FR[0-9A-Z]{2}[0-9]{9}`).
     - `CIUS-FR-03` : Recommandation Code Routage / Réf Acheteur.
   - Niveaux de sévérité stricts (`ERROR` vs `WARNING`).

3. **Jeux d'épreuves (Golden Invoices) (`pkg/validation/schematron_test.go`)**
   - Facture valide 100% conforme.
   - Facture avec incohérence monétaire (déclenchement garanti de `BR-CO-15`).
   - Facture avec date d'échéance antérieure (déclenchement garanti de `BR-CO-25`).
   - Facture avec SIRET français corrompu (déclenchement garanti de `CIUS-FR-01`).

4. **Anti-Rejeu Persistant & Outbox (`migrations/002_persistent_replay_outbox.sql` & `pkg/as4/`)**
   - Table `inbound_messages` avec contrainte unique `PRIMARY KEY (tenant_id, message_id)` pour remplacer le stockage en mémoire.
   - Table `webhook_events` pour la déduplication des webhooks.
   - Table `outbox_dispatches` pour la persistance des envois transactionnels.

5. **CI/CD Automatisée (`.github/workflows/ci.yml`)**
   - Bloque tout merge si `gofmt`, `go vet` ou un test Schematron échoue.
