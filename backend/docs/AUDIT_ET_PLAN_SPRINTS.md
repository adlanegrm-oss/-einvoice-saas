# Rapport d'Audit & Plan de Sprints d'Évolution — eInvoice SaaS

## 1. Audit Technique Détaillé de l'Existant

### 1.1 `internal/validator` vs `pkg/validation`
- **Existant** : Règles Go codées en dur avec détection élémentaire. Pas de pipeline Schematron officiel exécutant les règles CEN EN 16931 complètes, ni d'émulation de validation SVRL conforme. Les erreurs ne retournaient pas un code standardisé 422 Unprocessable Entity avec la granularité attendue par les PDP/PEPPOL.
- **Correction apportée** : Implémentation du moteur normatif `NormativeEngine` dans `internal/validator/normative.go` et `pkg/validation/schematron.go`. Analyse intégrale des règles EN 16931 et PEPPOL BIS v3 (BR-01 à BR-65, BR-CO-*, BR-CL-*, PEPPOL-EN16931-*). Retour structuré en HTTP 422 avec les balises XPath, RuleID et sévérité (`fatal`/`warning`).

### 1.2 `internal/exporter` vs `pkg/facturx`
- **Existant** : Injection binaire brute ou mock PDF minimaliste avec startxref et dictionnaires incomplets. Absence de conformité ISO 19005-3 (PDF/A-3b), absence d'extension XMP `pdfaExtension` namespace `fx` (`urn:factur-x.eu:1p0:`), pas de gestion de profil (MINIMUM, BASIC, EN16931).
- **Correction apportée** : Moteur Factur-X `internal/exporter/facturx_pdfa3.go` en pure Go. Intégration de l'arborescence PDF/A-3b complète : `/EmbeddedFiles` avec `/AFRelationship /Alternative`, nommage standard `factur-x.xml`, métadonnées XMP conformes aux spécifications FNFE-MPE / ZUGFeRD, et support des profils MINIMUM, BASIC, EN16931.

### 1.3 `internal/clearance` & `internal/routing` (AS4 & PDP / KSeF)
- **Existant** : Connecteurs HTTP fictifs et anti-rejeu en mémoire volatile sans persistance ni garantie transactionnelle ACID.
- **Correction apportée** : 
  - Connecteur AS4 sécurisé (`internal/routing/as4/as4_client.go`) avec enveloppe ebMS3, corrélation `MessageId` / `RefToMessageId`, stratégie d'exponential backoff avec Dead Letter Queue (DLQ), mTLS et validation de certificat X.509.
  - Connecteur PDP France / PPF (`internal/clearance/ppf_connector.go`) avec authentification PISTE (OAuth2 client credentials), gestion des statuts de cycle de vie (200 Déposée, 201 Rejetée, etc.).
  - Connecteur KSeF Pologne (`internal/clearance/ksef_connector.go`) avec session d'autorisation, validation FA(2) et vérification de traitement asynchrone.

### 1.4 `internal/auth` & `internal/repository`
- **Existant** : Stockage en mémoire ou utilisateurs volatils non persistés dans SQLite.
- **Correction apportée** : Migration SQL versionnée `000004_users_multitenant.up.sql`, modèle persistant `User` avec hachage Argon2id / bcrypt, contrôle d'accès basé sur les rôles (RBAC : `admin`, `operator`, `auditor`), validation JWT stricte avec whitelist d'algorithmes (HMAC-SHA256), vérification obligatoire `iss`, `aud`, `exp`, `nbf`.

---

## 2. Plan de Sprints d'Implémentation

| Sprint | Domaine | Objectif Principal | Effort Estimé |
| :--- | :--- | :--- | :--- |
| **Sprint 1** | **P0 - Conformité & Validation** | Moteur EN16931 + PEPPOL BIS Schematron/XSD + HTTP 422 | 5 j/h |
| **Sprint 2** | **P0 - Factur-X PDF/A-3** | Moteur PDF/A-3b pure Go, embedded XML, XMP namespaces | 4 j/h |
| **Sprint 3** | **P0 - Connecteurs Réels & AS4** | Client ebMS3/AS4 + DLQ, Connecteurs PPF (PISTE) & KSeF | 5 j/h |
| **Sprint 4** | **P0 - Persistance & Sécurité** | Table `users`, RBAC, migrations SQL, JWT durci, CSP | 4 j/h |
| **Sprint 5** | **P1 - Observabilité & PRA** | Prometheus metrics tenant-safe, script de backup chiffré | 3 j/h |
| **Sprint 6** | **P2 - Documentation & Finition**| Revue globale, synchronisation documentation et tests e2e | 2 j/h |
