### 🚀 Release : Module Front Client - Dépôt de Factures & Moteur de Conformité EN 16931

#### 📦 Fonctionnalités livrées :
- **Composant DropZone unifié & multi-formats** :
  - Support des factures aux formats UBL (.xml), CII / Factur-X (.pdf/.xml) et EDIFACT D01B (.edi).
  - Gestion du dépôt unitaire ou par lot (bouton d'upload multiple et glisser-déposer).
- **Moteur d'inspection et de règles légales & fiscales (CGI Art. 242 nonies A)** :
  - Analyse du contenu des flux XML (contrôle des balises `CustomizationID` CIUS-FR, `CompanyID`, `PartyTaxScheme` et équilibre TVA `TaxTotal`).
  - Détection de la conformité du conteneur PDF/A-3 et vérification de la présence de la charge structurée `factur-x.xml`.
  - Contrôle syntaxique des interchanges EDIFACT (UNB, UNH, BGM, DTM).
- **Restitution métier & Aide à la remédiation** :
  - Affichage modulaire par facture avec badges de conformité instantanés.
  - Volet correctif détaillant les anomalies fiscales/techniques et les actions recommandées pour corriger la facture.
- **Reporting & Traçabilité Front-Office (Audit Trail)** :
  - Accusé de réception global du lot avec référence d'enregistrement (`DEP-XXXXXX`).
  - Téléchargement du rapport d'audit au format JSON structuré et export de la synthèse au format CSV.
  - Journal d'audit chronologique en direct (horodatage, événement, empreinte SHA-256 simulée, statut).

#### 🛠️ Correctifs techniques & Architecture :
- Correction de l'orchestration dans `main.js` et `GenericView.js` (alignement des paramètres et initialisation post-rendu DOM de `DropZone`).
- Résolution des exceptions bloquantes sur `baseRoute` et nettoyage des traces de cache navigateur.
- Ajout d'un jeu de tests de référence (`factures_test_lots`) validant les scénarios conformes et rejetés.
