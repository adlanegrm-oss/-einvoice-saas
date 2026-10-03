# Plan d'Action & Roadmap — eInvoice SaaS
> Trajectoire de qualification : Du socle fonctionnel à l'état prêt pour production exposée et pré-certification PDP.

Le plan est découpé en **4 sprints de 1 à 2 semaines** chacun (environ 6-8 semaines au total).

---

## Sprint 0 — Sécurisation immédiate (3-5 jours)
**Objectif** : Rendre le système non exposable de façon dangereuse dès maintenant.

| Tâche | Priorité | Effort | Livrable |
|-------|----------|--------|----------|
| Utiliser uniquement \docker-compose.hardened.yml\ | P0 | 0,5 j | Compose sans ports exposés |
| Mettre un reverse-proxy TLS (Caddy recommandé) | P0 | 1 j | Domaines HTTPS fonctionnels |
| Régénérer **tous** les secrets (JWT + mots de passe) | P0 | 0,5 j | Nouveaux secrets forts |
| Purger l’historique Git des anciens mots de passe | P0 | 1 j | Historique nettoyé (\git filter-repo\) |
| Rendre \ADMIN_EMAIL_*\ et mots de passe obligatoires | P0 | 0,5 j | \.env.example.hardened\ à jour |
| Script \preflight.sh\ exécuté systématiquement | P0 | 0,5 j | Check automatique avant déploiement |

**Critère de sortie** : Aucun port applicatif n’est accessible depuis Internet. TLS terminé au proxy. Secrets ne sont plus dans Git.

---

## Sprint 1 — Fondations de sécurité & multi-tenant (1-1,5 semaine)

| Tâche | Priorité | Effort | Description |
|-------|----------|--------|-------------|
| Persistance des comptes utilisateurs en SQLite | P0 | 2 j | Table \users\ + migration + hash PBKDF2 |
| Validation JWT complète (\iss\, \ud\, \exp\, \
bf\) | P0 | 1 j | Refus strict des tokens mal formés |
| Isolation tenant vérifiée sur **tous** les endpoints | P0 | 1,5 j | Tests négatifs inter-tenant (404 systématique) |
| Rate limiting par tenant + par IP | P1 | 1 j | Token bucket sur les routes sensibles |
| Protection upload renforcée (taille, magic bytes, XXE) | P0 | 1 j | Refus DOCTYPE + limite mémoire |
| Headers de sécurité (CSP, HSTS, X-Frame-Options…) | P1 | 0,5 j | Middleware unifié |

**Critère de sortie** :
- Un utilisateur CLIENT ne peut plus accéder aux données d’un autre tenant.
- Les comptes survivent au redémarrage.
- Tests de sécurité automatisés passent (401/403/422 attendus).

---

## Sprint 2 — Conformité réglementaire & formats (1,5-2 semaines)

| Tâche | Priorité | Effort | Description |
|-------|----------|--------|-------------|
| Validation XSD + Schematron EN 16931 / PEPPOL BIS | P0 | 3 j | Intégration d’un validateur XSD (ex: libxml2 via pure Go ou service) |
| Finalisation générateur Factur-X PDF/A-3 | P0 | 2,5 j | Vendeur, devise, montant à payer, XMP complet |
| Pipeline multi-format unifié (FatturaPA, Facturae, KSeF) | P1 | 2 j | \MultiFormatEngine\ réellement branché de bout en bout |
| Audit trail chaîné (\hash_n = SHA256(event \|\| hash_{n-1})\) | P0 | 1,5 j | Preuve d’immutabilité renforcée |
| Idempotence réelle sur \POST /invoices/emit\ | P0 | 1 j | Clé d’idempotence + gestion des doublons |

**Critère de sortie** :
- Une facture non conforme EN 16931 est rejetée avec le code BR exact.
- Factur-X généré est un vrai PDF/A-3 avec XML embarqué valide.
- L’audit trail résiste à une modification manuelle de la base.

---

## Sprint 3 — Industrialisation & connectivité (1,5-2 semaines)

| Tâche | Priorité | Effort | Description |
|-------|----------|--------|-------------|
| Connecteur AS4 / PDP réel (ou mock avancé + interface) | P1 | 3 j | Retry, backoff, DLQ, corrélation message-id |
| Observabilité (métriques Prometheus + logs structurés) | P1 | 1,5 j | \/metrics\ + labels par tenant (sans fuite de données) |
| Sauvegarde + PRA formalisé | P0 | 1 j | Script backup + test de restauration documenté (RPO/RTO) |
| CI/CD renforcée | P1 | 1 j | \govulncheck\ + SBOM + scan d’image + tests de charge automatiques |
| Documentation d’exploitation finale | P1 | 1 j | Mise à jour \DOCUMENT_D_EXPLOITATION\ + runbook |

**Critère de sortie** :
- Déploiement automatisé possible.
- Sauvegarde restaurable en < RTO défini.
- Métriques visibles et alertes configurables.

---

## Vue d’ensemble du planning

\\\
Semaine 1     : Sprint 0 (sécurisation immédiate)
Semaine 2-3   : Sprint 1 (sécurité + multi-tenant)
Semaine 4-5   : Sprint 2 (conformité + formats)
Semaine 6-7/8 : Sprint 3 (industrialisation + connecteurs)
\\\

---

## Indicateurs de succès finaux

- [ ] 0 port applicatif exposé
- [ ] Secrets jamais dans Git
- [ ] Tests inter-tenant 100 % négatifs
- [ ] Validation EN 16931 + XSD passante
- [ ] Factur-X PDF/A-3 valide
- [ ] Audit trail chaîné et vérifiable
- [ ] Backup + restauration testée
- [ ] CI verte avec scan de vulnérabilités
- [ ] Documentation d’exploitation à jour
