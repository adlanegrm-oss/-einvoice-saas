# Matrice des Capacités & Conformité Plateforme (Octobre 2026)

Ce document établit l'état de conformité réel, audité et technique des modules de l'API E-Invoice Infrastructure.

| Composant / Fonction | État Réel | Preuve / Validation | Cible de Déploiement |
| :--- | :--- | :--- | :--- |
| **Parsing UBL 2.1 / CII** | ✅ Production-ready | Tests unitaires & golden fixtures | Multi-tenant SaaS |
| **Modèle Monétaire Exact** | ✅ Production-ready | Arithmétique entière (Fixed-point 4 décimales) | Multi-tenant SaaS |
| **Validation Syntaxique XML** | ✅ Production-ready | Parser durci anti-XXE, sans DTD | Multi-tenant SaaS |
| **Validation Règles EN 16931** | 🟡 En cours (BR-01 à BR-65) | Moteur de règles internes typées | Validation Sandbox / API |
| **Schematron XSLT / ISO** | 🔴 Roadmap Q1 2027 | Moteur natif en cours d'évaluation | Sandbox |
| **Factur-X (ZUGFeRD 2.2)** | 🟡 Pilote | Extraction & injection PDF/A-3 | Déploiement pilote |
| **Transactional Outbox & DLQ** | ✅ Production-ready | Workers concurrents, backoff exponentiel | Multi-tenant SaaS |
| **Idempotence Persistante** | ✅ Production-ready | DB backing table, détection 409 | Multi-tenant SaaS |
| **Isolation Multi-Tenant** | ✅ Production-ready | Row-level `tenant_id` & scopes API Keys | Multi-tenant SaaS |
| **Connecteur AS4 / ebMS3** | 🟡 Adapter Mock & TLS PoC | Client testé en environnement simulé | Pilote connecté |
| **Routage PDP / Réseau** | 🟡 Interface abstraite | `TransportAdapter` pluggable | Partenariats PDP |
| **Audit Trail Cryptographique** | ✅ Production-ready | Chaîne Merkle / SHA-256 scellée | Multi-tenant SaaS |
| **Statut Réglementaire DGFiP** | ℹ️ Solution Compatible | Non immatriculée PDP (Orchestration & Validation) | B2B Middleware |

> **Avertissement de positionnement juridique :** La solution est une infrastructure logicielle de conformité et d'interfaçage B2B ("Solution Compatible"). Elle ne prétend pas au statut de Plateforme de Dématérialisation Partenaire (PDP) au sens de l'article 290 B du CGI.
