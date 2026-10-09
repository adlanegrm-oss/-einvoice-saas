# Rapport de test isolé — pdfsign RC5

**Statut :** test de signature et de vérification réussi dans un environnement isolé.

## Objectif

Évaluer le parcours de base de signature PDF avec la version
`github.com/digitorus/pdfsign v1.0.0-rc5`, avant toute intégration
dans le backend de facturation électronique.

## Environnement et résultats

- Programme de test Go autonome, séparé du projet applicatif.
- Vérification de la correspondance entre le certificat de test et la clé privée : réussie.
- Signature du PDF d'essai au format `PAdES_B_B` : réussie.
- Réouverture du PDF signé et détection d'une signature : réussies.
- Vérification cryptographique dans le scénario de test : réussie.
- Nombre de signatures détectées : 1.

## Limites

Ce résultat valide uniquement le scénario exécuté sur un PDF et un certificat
de développement auto-signé. Il ne démontre pas à lui seul :

- la conformité complète aux exigences PAdES ;
- la détection de toutes les altérations possibles ;
- la couverture intégrale du document signé ;
- la confiance accordée à un certificat de production ;
- la conformité juridique de la signature pour les usages visés.

La confiance accordée au certificat auto-signé dans le test est une option
de diagnostic et ne constitue pas une politique de confiance de production.

## Prochaines étapes

1. Étudier l'intégration de l'API RC5 dans le projet expérimental séparé.
2. Ajouter des tests négatifs portant sur la modification du PDF après signature.
3. Tester les cas d'erreur, les documents invalides et les certificats non fiables.
4. Évaluer les exigences de conservation des clés et les exigences juridiques
   applicables avant toute mise en production.

## Séparation des environnements

Le test isolé ne modifie pas le code du backend de facturation électronique.
Aucune clé privée, aucun certificat de développement et aucun PDF de test
ne doivent être ajoutés à ce dépôt.
