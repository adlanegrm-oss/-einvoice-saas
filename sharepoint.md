Voici une version corrigée, enrichie et formalisée de la documentation, accompagnée de schémas textuels clairs (architecture et flux de données) pour illustrer parfaitement ton projet avant de l'exporter en PDF pour ton SharePoint.
📄 Documentation Technique & Fonctionnelle : SaaS d'E-Invoicing & Auditabilité
1. Vue d'ensemble du Projet
Ce SaaS d'e-invoicing est développé en Go pour répondre à des exigences de haute performance, de faible empreinte mémoire et de stricte traçabilité (audit-ready). Il automatise la réception, la validation, la structuration, le hachage cryptographique et l'archivage de factures électroniques.
2. Architecture Globale du Système
Le schéma ci-dessous illustre la séparation des couches, depuis le point d'entrée de l'application jusqu'à la base de données SQLite :
[ Point d'Entrée / Client ] (cmd/server/main.go)
           │
           ▼
[ Logique Métier & Validation ]
           │
           ├──────────────────────────────┐
           ▼                              ▼
  [ Moteur de Hachage ]          [ Couche de Persistance ]
  (internal/repository/hash.go)  (internal/repository/sqlite.go)
           │                              │
           └──────────────┬───────────────┘
                          ▼
             [ Base de Données SQLite ]
             (Mode WAL activé & Transactions)

3. Schéma Relationnel de la Base de Données (registre historique ; les factures structurées de l'API sont dans la table `invoices`, distincte)
La persistance repose sur une structure relationnelle optimisée à 3 tables pour garantir l'intégrité et l'historique des opérations :
┌───────────────────────────             ┌───────────────────────────
│   registry_invoices                    │       registry_items       
├───────────────────────────             ├───────────────────────────
│ PK  id                    │◄───────────│ PK  id                    │
│     invoice_number (UQ)   │            │ FK  invoice_id            │
│     created_at            │            │     description           │
└───────────────────────────             │     quantity              │
                                         │     unit_price            │
                                         └───────────────────────────
┌───────────────────────────
│        invoice_logs                    
├───────────────────────────
│ PK  id                    │
│     invoice_number        │
│     status                │
│     invoice_hash          │
│     processed_at          │
└───────────────────────────

Description des tables :
 * registry_invoices : En-tête de la facture (contrainte d'unicité sur le numéro de facture).
 * registry_items : Lignes de détails liées à la facture (suppression en cascade ON DELETE CASCADE).
 * invoice_logs : Journal d'audit immuable consignant chaque tentative de traitement, le statut (ACCEPTED / REJECTED) et le hash cryptographique.
4. Sécurité et Intégrité : Processus de Hachage SHA-256
Pour empêcher toute altération des données, chaque facture subit une empreinte cryptographique avant d'être journalisée :
 * Sérialisation : Transformation de la structure de la facture en un flux JSON déterministe via encoding/json.
 * Hachage : Application de l'algorithme SHA-256 (crypto/sha256) sur le flux obtenu.
 * Traçabilité : Le hash hexadécimal généré est stocké dans la table invoice_logs. Le moindre changement d'un montant ou d'un libellé modifie radicalement l'empreinte.
5. Flux de Traitement (Workflow d'Exécution)
Le diagramme séquentiel ci-dessous décrit le parcours d'une facture au sein du programme principal (main.go) :
Main            Repository (SQLite)       Hash Module        Audit Logs
 │                       │                     │                   │
 │── 1. Init DB (WAL) ──>│                     │                   │
 │                       │                     │                   │
 │── 2. Build Invoice ──>│                     │                   │
 │                       │                     │                   │
 │── 3. Generate Hash ────────────────────────>│                   │
 │<── Return Hash ─────────────────────────────│                   │
 │                       │                     │                   │
 │── 4. Insert Invoice ->│                     │                   │
 │                       │                     │                   │
 │── 5. Log Processing ─>│────────────────────────────────────────>│
 │                       │                     │                   │

6. Qualité, Tests et Intégration Continue (CI/CD)
 * Automatisation : Utilisation de GitHub Actions pour exécuter à chaque commit la commande de vérification de compilation (go build -v ./...).
 * Concurrence SQLite : Activation dynamique du mode WAL (Write-Ahead Logging) permettant d'optimiser les performances d'écriture sous forte charge concurrente.
💡 Recommandation pour votre SharePoint :
Tu peux copier l'intégralité de ce texte corrigé et structuré dans un traitement de texte (Word, Google Docs), l'enregistrer sous le format .pdf, puis l'importer directement dans le répertoire documentaire de ton espace SharePoint.
