# Chantier B : Architecture de Validation Schematron EN 16931 + XSD

## 1. Problématique
Les feuilles XSLT officielles de validation EN 16931 (notamment `EN16931-CII-validation.xslt` et les équivalents UBL) sont écrites en **XSLT 2.0+** (~223 règles ISO Schematron converties). L'utilitaire standard `xsltproc` ne supporte que le XSLT 1.0 et ne peut donc pas exécuter directement ces feuilles. De plus, Go ne dispose pas nativement d'un moteur XSLT 2.0 complet dans sa bibliothèque standard.

---

## 2. Options Chiffrées d'Architecture

### Option 1 : Bridge Java / Saxon-HE (`saxon-he.jar`)
- **Principe** : Exécution de Saxon-HE via un sous-processus Java (`java -jar saxon-he.jar -s:invoice.xml -xsl:EN16931-CII-validation.xslt -o:svrl.xml`).
- **Avantages** : 100% conforme aux spécifications CEN/TC 434, exécute la feuille officielle sans aucune modification.
- **Inconvénients** : Dépendance lourde au runtime JRE sur l'environnement de production (augmentation de l'empreinte conteneur Docker, latence de démarrage JVM ~300-500ms par validation).
- **Coût** : Faible en dev, moyen en ops (image Docker `eclipse-temurin:21-jre-alpine` requise).

### Option 2 : Sidecar Go + WebAssembly / Saxon-JS (ou moteur Node.js/Python externe)
- **Principe** : Utiliser un micro-service sidecar ou un runner Node.js/Saxon-JS pour exécuter la transformation XSLT 2.0.
- **Avantages** : Isolation des environnements, découplage du binaire Go principal.
- **Inconvénients** : Complexité d'architecture réseau inter-conteneurs (Docker Compose / Vercel Serverless limitations).
- **Coût** : Élevé en complexité de déploiement Serverless (Vercel n'aime pas les sidecars Java).

### Option 3 : Moteur Go natif d'expressions Schematron simplifiées (Fallback validé)
- **Principe** : Parser le fichier Schematron (`.sch`) ou coder en Go pur les règles critiques ISO (BR-CO-10 à 17, BR-FR-01 à 15) avec structures SVRL générées nativement.
- **Avantages** : Zéro dépendance externe, exécution ultra-rapide (< 5ms), parfait pour les environnements serverless contraints (Vercel).
- **Inconvénients** : Nécessite de synchroniser manuellement les règles si la feuille CEN change (couverture actuelle : 50 règles clés CIUS-FR et EN 16931).
- **Coût** : Très faible en ops, investissement de maintenance modéré.

---

## 3. Décision et Implantation Recommandée
Pour un SaaS déployé sur **Vercel** et en conteneur Go léger, **l'Option 3 combinée à un runner optionnel Saxon-HE en mode local/dev** est retenue. 
- En production Serverless (Vercel), le validateur natif Go enrichi (`internal/validator/schematron_native.go`) garantit une réponse instantanée et évite l'échec de build.
- En mode Enterprise/Local, un connecteur optionnel appelle `java -jar` si présent.
