export const ROLES_CONFIG = {
  CLIENT: {
    label: "Front Client",
    path: "/app",
    color: "#2563eb",
    defaultTheme: "light",
    description: "Utilisation quotidienne métier (dépôt, suivi, téléchargement)",
    routes: [
      { id: "client-upload", label: "Déposer des factures" },
      { id: "client-invoices", label: "Liste des factures" },
      { id: "client-validation", label: "Validation seule" }
    ]
  },
  PARTNER: {
    label: "Front Partenaire",
    path: "/partner",
    color: "#7c3aed",
    defaultTheme: "light",
    description: "Intégrateurs, cabinets, ERP, PDP partenaires (multi-tenants)",
    routes: [
      { id: "partner-dash", label: "Dashboard Consolidé" },
      { id: "partner-clients", label: "Gestion des Clients" },
      { id: "partner-api", label: "Console API & Sandbox" }
    ]
  },
  ADMIN: {
    label: "Console Admin",
    path: "/admin",
    color: "#ea580c",
    defaultTheme: "dark",
    description: "Administration globale multi-tenant, contrôle et gouvernance",
    routes: [
      { id: "admin-users", label: "Utilisateurs & Rôles" },
      { id: "admin-tenants", label: "Gestion des Tenants" },
      { id: "admin-outbox", label: "Supervision Outbox & Audit" }
    ]
  },
  OPS: {
    label: "Console Exploitation",
    path: "/ops",
    color: "#059669",
    defaultTheme: "dark",
    description: "Supervision production, incidents, santé workers, DLQ",
    routes: [
      { id: "ops-metrics", label: "Santé Système & Métriques" },
      { id: "ops-dlq", label: "Outbox & File DLQ" }
    ]
  },
  MCO: {
    label: "Console MCO",
    path: "/mco",
    color: "#475569",
    defaultTheme: "dark",
    description: "Maintenance, support N2/N3, runbooks et plans PRA",
    routes: [
      { id: "mco-runbooks", label: "Runbooks d'exploitation" },
      { id: "mco-pra", label: "Procédures PRA & Backup" }
    ]
  },
  RND: {
    label: "Console R&D",
    path: "/rnd",
    color: "#4f46e5",
    defaultTheme: "dark",
    description: "Sandbox de conformité, règles BR-xx et qualification CIUS-FR",
    routes: [
      { id: "rnd-schematron", label: "Banc de Test Schematron" },
      { id: "rnd-golden", label: "Comparateur Golden Files" }
    ]
  },
  DEPLOY: {
    label: "Console Déploiement",
    path: "/deploy",
    color: "#0d9488",
    defaultTheme: "dark",
    description: "Releases, statuts environnements, CI/CD et rollbacks",
    routes: [
      { id: "deploy-envs", label: "Statut des Environnements" },
      { id: "deploy-pipeline", label: "Pipelines CI/CD & Scans" }
    ]
  }
};
