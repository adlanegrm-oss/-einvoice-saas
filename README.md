# E-Invoice Compliance Gateway (Prototype Technique Avancé)

Passerelle API et moteur de normalisation, validation et sécurisation des flux de facturation électronique (UBL 2.1, CII, Factur-X) pour ERP et applications de gestion d'entreprise.

> **Avertissement de conformité :** Ce logiciel est un socle technologique en pré-production technique. Il n'est ni une Plateforme de Dématérialisation Partenaire (PDP) immatriculée, ni qualifié officiellement pour l'émission fiscale directe sans plateforme intermédiaire agréée.

Consultez [`CAPABILITY_MATRIX.md`](CAPABILITY_MATRIX.md) pour la grille de maturité composant par composant.

## Cycle d'Intégrité d'une Facture
1. **Réception & Empreinte :** Calcul SHA-256 du document brut.
2. **Idempotence :** Verrouillage strict par clé et détection des rejeux concurrents.
3. **Contrôles Sémantiques :** Validation des règles fondamentales EN 16931 (Ruleset 2026.1).
4. **Scellement Cryptographique :** Chaînage d'événements Merkle irréversible (`SHA-256(event + prev_hash)`).
5. **Outbox Durable :** Persistance avant émission réseau avec reprise sur panne (DLQ).
6. **Passerelle Réseau :** Interface d'acheminement standardisée avec reçu normé.
7. **Evidence Ledger :** Exportation d'un dossier de preuve d'audit consolidé.
