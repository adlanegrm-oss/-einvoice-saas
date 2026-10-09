### Avancement — Ingestion transactionnelle des factures

**Commit de référence :** `9bef74d` — `fix: make invoice ingestion transactional`

#### Validations effectuées

* Tests Go réussis.
* `go vet ./...` réussi.
* Test d’intégration PostgreSQL validant le rollback transactionnel.
* Pipeline GitHub Actions réussi.

#### Prochaines étapes

* Préparer l’environnement de préproduction sans déploiement immédiat.
* Concevoir un module de signature PDF PAdES intégré au backend Go, sans prestataire externe.
* Commencer par un prototype isolé avec un certificat de test.
* Prévoir le chiffrement et l’isolation des clés privées par client.
* Évaluer les exigences juridiques applicables au Maroc et à l’Union européenne avant toute commercialisation de la signature.

**Périmètre actuel :** aucun changement supplémentaire du code ni déploiement. Les fichiers locaux de test, XML et sauvegardes doivent être préservés.

La prochaine étape sera définie après validation de la feuille de route.
