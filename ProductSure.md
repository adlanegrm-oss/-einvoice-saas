# PRODUCT SURE — DOSSIER PRODUIT & ERGONOMIE D'EXPLOITATION
**E-Invoice SaaS Platform — Référence Architecture, Ergonomie & Conformité 2026**

---

## 1. VISION ET POSITIONNEMENT DU PRODUIT
E-Invoice SaaS est une infrastructure souveraine et étanche conçue pour automatiser le cycle complet de facturation électronique conforme à la norme européenne EN 16931, sans friction technique pour l'utilisateur final.

### Piliers Fondamentaux
- **Étanchéité Multi-Tenant Absolue** : Aucune colocation de tables ; chaque organisation dispose d'une partition logique et physique scellée.
- **Scellement Cryptographique SHA-256** : Calcul d'empreinte immuable dès la validation de la pièce garantissant la piste d'audit fiable (PAF).
- **Rétention & Clôture Déterministe M+3** : Purge transactionnelle en deux temps (SQL transactionnel puis filesystem) avec contrôle strict des états terminaux.
- **Interopérabilité Européenne** : Prise en charge native de Factur-X (CII/PDF), UBL 2.1, EDIFACT, ainsi que des connecteurs de clearance (PPF France, KSeF Pologne).

---

## 2. ERGONOMIE ET PARCOURS UTILISATEUR (UX/UI)

### 2.1. Système Visuel & Lisibilité
- **Fonds & Contrastes** : Arrière-plans institutionnels sombres (`#0A192F`, `#112240`) assurant un confort visuel prolongé pour les gestionnaires financiers.
- **Actions Prioritaires** : Boutons d'émission et de validation en bleu cobalt (`#2563EB`), offrant un guidage visuel immédiat.
- **Codes d'État Universels** :
  - `VERT` : Facture validée, scellée, prête pour acheminement.
  - `AMBRE` : En attente de clearance ou traitement asynchrone dans le worker pool.
  - `ROUGE` : Rejet Schematron explicité ligne par ligne avec code règle SVRL.

### 2.2. Parcours Clés de l'Application
1. **Dépôt & Contrôle en Moins de 3 Clics** :
   - Glisser-déposer du document comptable (PDF, XML).
   - Contrôle syntaxique et calcul arithmétique instantané (HT + TVA = TTC exact).
   - Restitution du certificat de conformité numérique.
2. **Tableau de Bord de Clôture Mensuelle** :
   - Visualisation groupée des factures par cycle comptable (`YYYY-MM`).
   - Indicateur clair de statut de verrouillage avant application de la politique de rétention.
3. **Piste d'Audit & Historique des Statuts** :
   - Timeline chronologique non modifiable traçant l'auteur, l'horodatage UTC et l'empreinte de charge utile à chaque transition.

---

## 3. ARCHITECTURE TECHNIQUE ET STABILITÉ
- **Moteur Backend** : Écrit en Go, optimisé pour la concurrence et les faibles temps de latence.
- **Base de Données** : SQLite configuré en mode WAL (Write-Ahead Logging) pour des écritures transactionnelles concurrentes sans verrouillage de lecture.
- **Routage Asynchrone** : File d'attente Outbox avec reprise exponentielle (Backoff) et Dead-Letter Queue (DLQ) pour l'expédition AS4/Peppol.
- **Performance** : Worker pool borné empêchant tout épuisement mémoire lors des pics d'activité.

---

## 4. DÉPLOIEMENT & EXPLOITATION
- Exécutable autonome compilable sans dépendances externes complexes.
- API REST sécurisée par jetons d'accès JWT, horodatage d'invalidation et rate-limiting au niveau IP et compte.
- Monitoring natif intégrant métriques Prometheus, journalisation structurée JSON et traces d'exécution.