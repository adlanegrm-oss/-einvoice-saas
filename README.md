# Documentation Technique - E-Invoice SaaS (v1.4.0)

## 1. Vue d'ensemble de l'Architecture
Le SaaS repose sur une architecture **Serverless** (déployée sur Vercel) avec un point d'entrée unique (`api/index.go`) développé en **Go**. Les données de traçabilité et les journaux d'événements sont reliés à une base de données relationnelle **PostgreSQL** (Piste d'Audit Fiable - PAF).

---

## 2. Répertoire des Endpoints & Chemins

| Méthode | Chemin | Description | Module Associé |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Informations de base du SaaS et liste des routes | Core / Router |
| `GET` | `/health` | État de santé global du système et des services | Monitoring |
| `GET` | `/health/db` | Vérification de la connectivité PostgreSQL | Database / PAF |
| `POST/GET` | `/ocr/scan-and-audit` | Numérisation OCR/LAD + Contrôles fiscaux & arithmétiques | OCR / Fiscal Engine |

---

## 3. Requêtes de Test en Production (`vercel curl`)

Pour vérifier le bon fonctionnement de chaque composant en production, utilisez les commandes suivantes dans votre terminal PowerShell :

### A. Test de santé et connectivité de la base de données
```powershell
npx vercel curl [https://einvoice-saas-nine.vercel.app/health](https://einvoice-saas-nine.vercel.app/health)
npx vercel curl [https://einvoice-saas-nine.vercel.app/health/db](https://einvoice-saas-nine.vercel.app/health/db)
