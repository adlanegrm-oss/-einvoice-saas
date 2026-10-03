package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var docs = map[string]string{
	// -------------------------------------------------------------------------
	// 1. README.md
	// -------------------------------------------------------------------------
	"README.md": `# eInvoice SaaS Engine

> Moteur d'orchestration, de pré-validation et de distribution de factures électroniques multi-formats (UBL 2.1, CII Factur-X, EDIFACT, SAP IDoc, Facturae, FatturaPA, KSeF).

---

## Vue d'ensemble

**eInvoice SaaS Engine** est une passerelle middleware haute performance développée en Go. Conçue pour répondre aux exigences des réformes réglementaires de facturation électronique B2B/B2G, la plateforme assure :

- **Ingestion & Détection Automatique :** Identification instantanée du standard documentaire (XML UBL/CII, EDIFACT D01B, SAP IDoc XML, XML Facturae 3.2, FatturaPA 1.2, FA_VAT KSeF).
- **Pré-validation Sémantique EN 16931 :** Contrôle des règles métier (BR), montants HT/TVA/TTC, cohérence des identifiants fiscaux (SIRET, TVA intracommunautaire) et devises.
- **Piste d'Audit Fiable (PAF) Immuable :** Chaînage cryptographique SHA-256 persistant (Genesis -> Bloc N) garantissant l'intégrité et la non-répudiation des transitions d'état.
- **Isolation Multi-Tenant Robuste :** Ségrégation stricte des données par ` + "`" + `OrganizationID` + "`" + ` au niveau repository et gestion des verrous d'idempotence.
- **Transport & Routage eDelivery :** Connecteurs d'aiguillage AS4 (ebMS3 / SOAP 1.2) avec queues de reprise outbox et dispatch vers plateformes de clearance (PPF, KSeF).

---

## Architecture du Pipeline

` + "```" + `
[ Facture Brute (JSON / XML / EDI) ]
                │
                ▼
        [ Ingestion API ] ─── (Idempotency Key & Rate Limiting)
                │
                ▼
     [ Détection de Format ] ─── UBL 2.1 / CII / EDIFACT / IDoc / FatturaPA / KSeF
                │
                ▼
  [ Modèle Canonique Pivot ] ─── Normalisation monétaire (invoice.Money)
                │
                ▼
   [ Pré-validation EN 16931 ] ─── Contrôle des règles métier BR-xx
                │
                ▼
 [ Scellement Cryptographique ] ─── Calcul C14N / SHA-256 du document
                │
                ▼
    [ Piste d'Audit Fiable ] ─── Chaînage immuable (event_hash = SHA256(event || prev_hash))
                │
                ▼
   [ Routage & Clearance ] ─── PPF / Peppol AS4 Dispatcher / KSeF Outbox
` + "```" + `

---

## Démarrage Rapide

### Prérequis
- Go 1.22+ (recommandé 1.24+)
- SQLite3 ou conteneur Docker compatible

### Installation locale

` + "```bash" + `
# Cloner le dépôt
git clone [https://github.com/adlanegrm-oss/-einvoice-saas.git](https://github.com/adlanegrm-oss/-einvoice-saas.git)
cd -einvoice-saas

# Télécharger les dépendances
go mod download

# Compiler le binaire serveur
go build -o bin/server.exe ./cmd/server
` + "```" + `

### Lancement

` + "```bash" + `
# Variables minimales de développement
export APP_ENV=development
export ADMIN_EMAIL=admin@mondomaine.com
export ADMIN_PASSWORD=ChangeMeWithAStrongPassword2026!
export DATABASE_PATH=./data/einvoice.db

./bin/server.exe
` + "```" + `

---

## Documentation Complète

Pour aller plus loin dans l'intégration et l'exploitation :
- **[PRODUIT_SAAS.md](PRODUIT_SAAS.md)** : Spécifications fonctionnelles, matrice de conformité et guides de mise en route pas-à-pas.
- **[docs/DOSSIER_EXPLOITATION.md](docs/DOSSIER_EXPLOITATION.md)** : Manuel d'exploitation opérationnel, observabilité, sauvegardes et plan de reprise d'activité (PRA).
`,

	// -------------------------------------------------------------------------
	// 2. docs/DOSSIER_EXPLOITATION.md
	// -------------------------------------------------------------------------
	"docs/DOSSIER_EXPLOITATION.md": `# Dossier d'Exploitation Opérationnel (DEX)

**Produit :** eInvoice SaaS Engine  
**Version :** 1.0.0-Release  
**Statut :** Production Ready  

---

## 1. Architecture Système & Déploiement

### 1.1 Composants d'infrastructure
Le service fonctionne sous forme d'un conteneur binaire autonome Go orchestré derrière un reverse-proxy HTTPS obligatoire :
- **Reverse Proxy / Terminaison TLS :** Caddy ou NGINX (TLS 1.3 forcé, HSTS activé, redirection HTTP -> HTTPS).
- **Service Applicatif :** ` + "`" + `einvoice-saas` + "`" + ` (port interne ` + "`" + `8080` + "`" + ` non exposé sur le réseau public).
- **Moteur de Stockage :** SQLite avec mode WAL activé (` + "`" + `PRAGMA journal_mode=WAL` + "`" + `) sur volume persistant sécurisé.

### 1.2 Schéma de Déploiement Sécurisé

` + "```" + `
[ Internet (Clients / ERP) ]
             │  (HTTPS - TLS 1.3 / Port 443)
             ▼
      [ Caddy / Nginx ] ─── (Terminaison TLS, Rate Limiting Edge, WAF)
             │  (Réseau Docker Interne / 127.0.0.1:8080)
             ▼
    [ eInvoice SaaS Engine ]
             │