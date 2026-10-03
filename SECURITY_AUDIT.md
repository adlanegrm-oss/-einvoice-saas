# Audit expert — métier, connectique, protocoles, sécurité

## 1. Constat principal

La documentation décrit une plateforme SaaS d'e-invoicing avec :
- isolation multi-tenant via claims JWT ;
- validation EN 16931 ;
- piste d'audit avec SHA-256 ;
- Factur-X/CII ;
- UBL 2.1, CII, EDIFACT D96A ;
- routage PPF/PDP ou AS4 ;
- SQLite WAL ;
- workers et tâches d'arrière-plan ;
- procédure de sauvegarde/restauration.

Ces éléments sont explicitement décrits dans le dossier d'exploitation. Voir notamment l'architecture et les contrôles métier. fileciteturn1file0L10-L26

## 2. Risques infrastructure observés

### A. Exposition réseau
Le compose actuel publie directement :
- recette sur `8081:8080`
- préproduction sur `8082:8080`
- production sur `8080:8080`

Cela signifie qu'un hôte Docker exposé peut rendre les trois applications accessibles directement, sans couche TLS/reverse-proxy dans le compose fourni. Le document exige pourtant une URL HTTPS publique pour la production. fileciteturn1file2L159-L178

**Correctif :**
- ne pas publier le port applicatif ;
- réseau Docker interne ;
- TLS au reverse-proxy ;
- seuls 80/443 sont publiés au frontal ;
- accès d'administration limité au réseau privé/VPN.

### B. Secrets
Les secrets JWT et mots de passe administrateur sont injectés comme variables d'environnement. C'est compatible avec l'application telle qu'elle est décrite, mais moins robuste qu'un gestionnaire de secrets ou des fichiers secrets montés en lecture seule.

**Correctif immédiat :**
- secrets distincts par environnement ;
- jamais de secret réel dans Git ;
- rotation planifiée ;
- contrôle de longueur/entropie ;
- ne pas journaliser les variables d'environnement.

Le `.env.example` impose déjà des secrets séparés, ce qui est une bonne base. Il faut conserver cette séparation.

### C. Compte administrateur
Le compose conserve `ADMIN_EMAIL` avec une valeur par défaut. En production, un défaut d'identité administrative est inutilement dangereux.

**Correctif :**
- rendre `ADMIN_EMAIL_PROD` obligatoire ;
- rendre le mot de passe obligatoire ;
- idéalement supprimer le bootstrap password après première configuration ;
- prévoir rotation/récupération via mécanisme contrôlé.

### D. Durcissement conteneur
Le compose fournit déjà `read_only`, `/tmp` en tmpfs et `no-new-privileges`. fileciteturn1file2L106-L114

**Renforcement ajouté :**
- `cap_drop: ALL`
- limites de processus/mémoire ;
- `init: true` ;
- réseau interne ;
- pas d'exposition directe de l'application ;
- conservation du principe non-root déjà présent dans le Dockerfile.

## 3. Métier e-invoicing à contrôler dans le code

La documentation annonce un contrôle EN 16931 et les transitions `DEPOSITED -> ISSUED` ou `REJECTED`. fileciteturn1file0L28-L34

Une vraie revue métier doit vérifier, dans le code :
1. unicité de la facture et prévention des doubles émissions ;
2. idempotency key sur `POST /api/v1/invoices/emit` ;
3. atomicité validation + création + transition de statut ;
4. interdiction de revenir d'un état fiscalement final vers un état antérieur ;
5. séparation claire entre brouillon, dépôt, émission, rejet, annulation/avoir et archivage ;
6. gestion des arrondis et montants au niveau de précision attendu ;
7. cohérence devise/TVA/taux/exemptions ;
8. validation des identifiants entreprise ;
9. conservation de la version du schéma/règles utilisée pour la validation ;
10. corrélation entre facture, document original, représentation structurée et preuve d'émission.

La documentation indique également que les erreurs sémantiques sont renvoyées en HTTP 422 avec le code BR violé. fileciteturn1file1L82-L87

## 4. Multi-tenant

Le document affirme une isolation fondée sur les claims JWT et une isolation de répertoires par Tenant ID. fileciteturn1file0L16-L21 fileciteturn1file1L71-L74

**Point critique à auditer dans le code :**
- le `tenant_id` ne doit jamais provenir d'un paramètre client lorsqu'un claim signé est disponible ;
- toutes les requêtes DB doivent être tenant-scoped ;
- les téléchargements doivent vérifier tenant + autorisation avant ouverture du fichier ;
- les jobs asynchrones doivent transporter un contexte tenant explicite ;
- les clés de cache doivent intégrer le tenant ;
- les exports et rapports doivent être filtrés par tenant ;
- les logs ne doivent pas permettre de mélanger des données de deux clients.

## 5. JWT / autorisation

La documentation prévoit `CLIENT` et `ADMIN`. fileciteturn1file1L86-L87

À vérifier impérativement dans le code :
- algorithme JWT explicitement autorisé, jamais accepté depuis le token ;
- validation `iss`, `aud`, `exp`, `nbf` et éventuellement `jti` ;
- durée de vie courte pour les access tokens ;
- mécanisme de rotation/révocation adapté ;
- séparation stricte des permissions et des rôles ;
- refus par défaut ;
- protection des endpoints de jobs (`ADMIN`) ;
- impossibilité pour un CLIENT de s'auto-attribuer ADMIN ;
- pas de confiance dans des headers `X-Tenant-*` fournis par le client.

## 6. API

Endpoints documentés :
- `POST /api/v1/invoices/emit`
- `GET /api/v1/invoices/{id}/audit-trail`
- `POST /api/v1/invoices/deposit`
- `GET /api/v1/invoices/download`
- endpoints de jobs/admin. fileciteturn1file1L60-L81

À ajouter/vérifier :
- limite de taille des requêtes ;
- timeouts serveur ;
- rate limiting par tenant et par identité ;
- idempotence ;
- validation stricte `Content-Type` ;
- contrôle des extensions et du contenu réel des fichiers ;
- protection XXE pour XML ;
- désactivation DTD/résolution externe ;
- prévention zip-bomb/XML-bomb ;
- noms de fichiers non contrôlés ;
- protection path traversal ;
- téléchargement en `Content-Disposition` sûr ;
- réponses d'erreur sans stack trace/secrets.

## 7. Connectique / protocoles

Le dossier annonce le routage PPF/PDP et AS4. fileciteturn1file0L28-L34

Pour une vraie qualification connecteur :
- TLS 1.2+ ;
- validation stricte des certificats ;
- pas de `InsecureSkipVerify` ;
- validation hostname/SAN ;
- timeouts connect/read/write ;
- retry avec backoff + jitter ;
- idempotence côté émission ;
- dead-letter/rejeu contrôlé ;
- corrélation `invoice_id` / message ID / transmission ID ;
- journalisation des réponses sans données sensibles ;
- signature/chiffrement et profils AS4 réellement conformes au partenaire ;
- contrôle anti-rejeu ;
- validation du contenu reçu ;
- gestion des erreurs temporaires vs définitives.

## 8. XML / formats facture

Formats annoncés : UBL 2.1, UN/CEFACT CII, EDIFACT D96A, Factur-X. fileciteturn1file1L71-L74

Pour Factur-X, la documentation annonce un XML embarqué dans PDF/A-3 avec `AFRelationship/Alternative`. fileciteturn1file0L20-L21

À tester dans le code :
- XML well-formed ;
- namespaces stricts ;
- règles métier après parsing ;
- absence de résolution externe ;
- correspondance PDF ↔ XML ;
- métadonnées PDF/A-3 ;
- relation de fichier embarqué ;
- hash du document effectivement transmis ;
- refus des documents dont les deux représentations divergent.

## 9. Piste d'audit

Le système annonce une empreinte SHA-256 à chaque statut et une restitution horodatée UTC avec acteur/preuve. fileciteturn1file1L67-L70

Le point à sécuriser est la différence entre :
- intégrité cryptographique d'un enregistrement ;
- chaîne d'audit réellement infalsifiable ;
- stockage durable ;
- preuve externe de l'horodatage.

Une simple SHA-256 calculée sur une ligne ne suffit pas à elle seule à démontrer l'immutabilité si un administrateur peut réécrire DB + hash.

**Recommandation de conception :**
`hash_n = SHA256(canonical_event_n || hash_{n-1})`, avec stockage append-only et sauvegarde hors système. Ajouter une politique de conservation et un contrôle périodique d'intégrité.

## 10. SQLite / concurrence / PRA

La documentation prévoit SQLite WAL et un backup à chaud via `.backup`, puis sauvegarde des archives. fileciteturn1file0L23-L26 fileciteturn1file1L88-L91

Le backup à chaud est pertinent, mais le PRA doit aussi définir :
- RPO ;
- RTO ;
- fréquence ;
- rétention ;
- chiffrement des backups ;
- copie hors machine ;
- test de restauration périodique ;
- contrôle d'intégrité ;
- restauration conjointe DB + archives ;
- procédure de rotation des clés/secrets.

## 11. Worker / jobs

La documentation indique 4 workers et une queue de 50 jobs. fileciteturn1file1L76-L81

À vérifier :
- persistance des jobs (une queue mémoire ne survit pas à un crash) ;
- idempotence des jobs ;
- verrouillage/dédoublonnage ;
- reprise après redémarrage ;
- DLQ ;
- retry borné ;
- visibilité/timeout ;
- isolation tenant ;
- privilèges minimum des jobs.

## 12. Priorités

### P0 — avant exposition Internet
- TLS/reverse proxy ;
- ne pas exposer les ports applicatifs ;
- validation JWT complète ;
- isolation tenant vérifiée sur chaque endpoint ;
- upload/XML sécurisé ;
- secrets réels non présents dans Git ;
- idempotence de l'émission ;
- contrôle des droits ADMIN.

### P1 — avant certification/partenaires
- tests de connectivité AS4/PDP ;
- tests de rejeu/retry ;
- tests de divergence Factur-X ;
- tests de transitions métier ;
- tests de concurrence SQLite ;
- audit trail append-only ;
- restauration PRA testée.

### P2 — industrialisation
- observabilité structurée ;
- métriques par tenant sans fuite de données ;
- alerting ;
- rotation secrets ;
- SBOM et scan images ;
- SAST/DAST/dependency scanning ;
- tests de charge ;
- tests de chaos sur workers/connecteurs.

## Conclusion

Le projet présente une base d'exploitation raisonnablement durcie, mais les artefacts fournis ne permettent pas de conclure que la plateforme est sûre ou conforme au niveau applicatif. La correction livrée traite surtout l'exposition réseau, le durcissement conteneur, le bootstrap et l'exploitation.

**Le prochain passage doit porter sur le dépôt Go complet et les connecteurs.**
