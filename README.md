# ⚡ E-Invoice SaaS Platform — Conformité EN 16931 & Échange PDP / B2B

Plateforme open-source de facturation électronique conforme à la réglementation européenne (**EN 16931**) et aux spécifications françaises **CIUS-FR** (mandat B2B / PPF / PDP).

---

## 🚀 Dernières Avancées (Release v1.2.0-compliance)

* **Validation Sémantique EN 16931 & CIUS-FR** : Moteur d'analyse XML direct (SAX/Token) sans dépendances lourdes retournant les vrais identifiants de règles normatifs (`BR-01` à `BR-09`, `BR-16`, `BR-21`, `BR-CO-10`, `BR-CO-13`, `BR-CO-14`, `BR-CO-15`, `BR-CO-16`, `BR-CO-25`, `CIUS-FR-01` à `CIUS-FR-05`).
* **Générateur CII Complet (UN/CEFACT D16B)** : Couverture intégrale du schéma standard `rsm:CrossIndustryInvoice` : accord commercial (`SellerTradeParty`, `BuyerTradeParty`, identifiants SIREN/SIRET, TVA, adresses, routage Chorus Pro), lignes complètes (`IncludedSupplyChainTradeLineItem`, unités UN/ECE Rec 20, prix nets, ventilation TVA par ligne) et règlement financier (`ApplicableTradeTax`, échéance, IBAN/BIC).
* **Anti-Rejeu & Idempotence Persistants** : Schéma de migration SQL (`migrations/002_persistent_replay_outbox.sql`) assurant l'unicité `(tenant_id, message_id)` en base relationnelle pour fiabiliser le transport AS4 et les webhooks.
* **Hygiène Logicielle & CI/CD** : Élimination stricte des redéclarations de types, `go vet` valide sans avertissement et suite de tests unitaires 100% au vert sur l'ensemble des modules.

---

## 🏗️ Architecture du Pipeline de Conformité

```
CanonicalInvoice
       │
       ▼
[pkg/syntax] GenerateCIIXML()  ──►  Flux XML UN/CEFACT D16B
       │
       ▼
[pkg/validation] SchematronEngine.ValidateXML()
       ├─► Validation syntaxique & balisage
       ├─► Règles arithmétiques & totaux (BR-CO-*)
       ├─► Règles nationales d'identité & TVA (CIUS-FR-*)
       └─► Rapport structuré (ERROR / WARNING)
```

---

## 🛠️ Commandes Utiles

```bash
# Vérification statique du code
go vet ./pkg/...

# Lancement de l'ensemble des tests (100% PASS)
go test -v -p 1 ./pkg/...

# Exécution de la suite d'épreuves Schematron & golden invoices
go test -v -p 1 ./pkg/validation/...
```
