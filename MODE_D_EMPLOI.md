Voici le manuel d'utilisation détaillé couvrant l'exploitation du Front-Office et du Back-Office de votre plateforme SaaS de facturation électronique.
Vous pouvez sauvegarder ce contenu directement à la racine de votre dépôt sous le fichier MODE_D_EMPLOI.md.
MANUEL D'UTILISATION & MODE D'EMPLOI
Platform SaaS B2B e-Invoicing (Front-Office & Back-Office)
1. MODE D'EMPLOI — FRONT-OFFICE (Espace Client & Partenaire)
Le Front-Office est destiné aux clients (donneurs d'ordre) et partenaires (fournisseurs) pour la gestion quotidienne des factures, le contrôle de conformité et l'export des fichiers.
1.1 Accès au Portail Web
 * Ouvrez votre navigateur et accédez à l'URL du service (par défaut : http://localhost:8080).
 * L'interface charge automatiquement le tableau de bord interactif (propulsé par Tailwind CSS).
1.2 Saisie et Création Manuelle d'une Facture
 * Sur le tableau de bord principal, cliquez sur "Nouvelle Facture".
 * Renseignez les champs obligatoires du formulaire :
   * Numéro de facture (ex: INV-2026-001).
   * Vendeur / Émetteur (Nom et numéro de TVA intracommunautaire).
   * Acheteur / Client (Nom et coordonnées).
   * Devise (EUR, USD, etc.).
   * Lignes d'articles (Désignation, Quantité > 0, Prix unitaire, Taux de TVA).
 * Cliquez sur "Calculer & Enregistrer" : les totaux HT et TTC sont calculés instantanément, et la facture est ajoutée au registre SQLite.
1.3 Validation de Fichiers Multi-Formats (Drag & Drop / Import)
Pour vérifier la conformité d'un document avant envoi ou archivage :
 * Naviguez vers l'onglet "Validateur de Documents".
 * Glissez-déposez ou sélectionnez un fichier parmi les formats supportés :
   * Factur-X / CII (.xml ou .pdf hybride)
   * UBL 2.1 / 2.3 (.xml)
   * EDIFACT (.edi / .txt - INVOIC)
   * IDDoc (.xml)
   * PDF Signé PAdES (.pdf)
 * L'application analyse la structure binaire et affiche :
   * Le format détecté (ex: FACTUR-X, UBL, EDIFACT).
   * L'état de conformité (Valide ou Invalide).
   * Les erreurs bloquantes ou avertissements éventuels.
1.4 Consultation du Registre & Export Factur-X
 * Dans l'onglet "Registre des Factures", consultez la liste de vos documents enregistrés avec leur statut.
 * Pour chaque facture, cliquez sur "Télécharger XML Factur-X" pour générer et récupérer l'empreinte XML structurée conforme à la norme CrossIndustryInvoice.
2. MODE D'EMPLOI — BACK-OFFICE (Administration & Supervision)
Le Back-Office est réservé aux administrateurs système et auditeurs fiscaux pour la gestion de la sécurité, le suivi des performances et la consultation des rapports globaux.
2.1 Authentification & Gestion des Clés API
Tous les accès aux endpoints API stratégiques du Back-Office nécessitent une clé API valide.
 * En-tête standard : X-API-Key: secret-api-key-123
 * Alternative Bearer Token : Authorization: Bearer secret-api-key-123
Exemple de requête sécurisée via cURL :
curl -X GET http://localhost:8080/api/v1/invoices \
  -H "X-API-Key: secret-api-key-123"

2.2 Supervision du Moteur Asynchrone (Worker Pool)
Pour les traitements de masse (batchs EDI, parsing lourd) :
 * Les fichiers reçus sont placés dans la file d'attente asynchrone Go.
 * Les administrateurs peuvent suivre l'exécution des tâches dans les journaux système :
   docker-compose logs -f einvoice-saas

 * En cas d'erreur de parsing, le document est marqué en statut REJECTED dans SQLite avec le détail de la balise ou du segment défaillant.
2.3 Génération des Rapports Financiers
L'API Back-Office permet de calculer en temps réel des rapports agrégés optimisés SQL :
 * Endpoint : GET /api/v1/reports
 * Commande d'extraction :
   curl -X GET http://localhost:8080/api/v1/reports \
  -H "Authorization: Bearer secret-api-key-123"

 * Réponse fournie :
   {
  "total_invoices": 42,
  "total_ht": 12500.50,
  "total_ttc": 15000.60
}

2.4 Registre d'Audit & Conformité Légale
 * Chaque action (création, validation, échec) est consignée dans un journal d'audit infalsifiable en base de données.
 * L'auditeur peut accéder à la documentation interactive des API à l'adresse : http://localhost:8080/swagger.yaml ou via l'interface Swagger UI (/docs.html).
