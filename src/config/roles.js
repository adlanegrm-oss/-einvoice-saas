export const ROLES_CONFIG = {
  CLIENT: {
    label: "Front Client (Entreprise)",
    path: "/app",
    color: "#2563eb",
    defaultTheme: "light",
    description: "Opérations quotidiennes, facturation et gestion documentaire",
    routes: [
      { id: "invoices-list", label: "Cycle des Factures" },
      { id: "client-upload", label: "Dépôt & Conformité" },
      { id: "customers-list", label: "Référentiel Clients" },
      { id: "org-profile", label: "Organisation & Succursales" },
      { id: "audit-logs", label: "Journal d'Audit" }
    ]
  },
  PARTNER: {
    label: "Front Partenaire",
    path: "/partner",
    color: "#7c3aed",
    defaultTheme: "light",
    description: "Intégrateurs, cabinets, ERP, PDP partenaires",
    routes: [
      { id: "partner-dash", label: "Dashboard Multi-Tenants" },
      { id: "customers-list", label: "Portefeuille Clients" }
    ]
  },
  ADMIN: {
    label: "Console Admin",
    path: "/admin",
    color: "#ea580c",
    defaultTheme: "dark",
    description: "Administration globale, habilitations RBAC et gouvernance",
    routes: [
      { id: "admin-users", label: "Utilisateurs & Permissions" },
      { id: "audit-logs", label: "Audit Global Append-Only" }
    ]
  },
  OPS: {
    label: "Console Exploitation",
    path: "/ops",
    color: "#059669",
    defaultTheme: "dark",
    description: "Supervision production, métriques temps réel et workers",
    routes: [
      { id: "ops-metrics", label: "Santé Système & Métriques" }
    ]
  },
  MCO: {
    label: "Console MCO",
    path: "/mco",
    color: "#475569",
    defaultTheme: "dark",
    description: "Maintenance, support N2/N3, runbooks et PRA",
    routes: [
      { id: "mco-runbooks", label: "Runbooks d'exploitation" }
    ]
  },
  RND: {
    label: "Console R&D",
    path: "/rnd",
    color: "#4f46e5",
    defaultTheme: "dark",
    description: "Sandbox de qualification et matrice Schematron BR-xx",
    routes: [
      { id: "rnd-schematron", label: "Banc de Test Schematron" }
    ]
  },
  DEPLOY: {
    label: "Console Déploiement",
    path: "/deploy",
    color: "#0d9488",
    defaultTheme: "dark",
    description: "Releases, statuts environnements et CI/CD",
    routes: [
      { id: "deploy-envs", label: "Statut des Environnements" }
    ]
  }
};
