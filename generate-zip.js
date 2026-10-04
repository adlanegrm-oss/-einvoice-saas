const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const PROJECT_DIR = 'e-invoice-saas-auth-matrix';

// Structure des répertoires
const directories = [
  `${PROJECT_DIR}/src/config`,
  `${PROJECT_DIR}/src/auth`,
  `${PROJECT_DIR}/src/components`,
  `${PROJECT_DIR}/src/views`,
  `${PROJECT_DIR}/src/styles`
];

directories.forEach(dir => fs.mkdirSync(dir, { recursive: true }));

// 1. package.json
fs.writeFileSync(`${PROJECT_DIR}/package.json`, JSON.stringify({
  name: "e-invoice-saas-auth-matrix",
  version: "1.0.0",
  description: "Matrice d'accès et profils pour la suite e-invoice-saas",
  type: "module",
  scripts: {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  devDependencies: {
    "vite": "^5.0.0"
  }
}, null, 2));

// 2. index.html
fs.writeFileSync(`${PROJECT_DIR}/index.html`, `<!DOCTYPE html>
<html lang="fr" data-role="CLIENT" data-theme="light">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>e-Invoice SaaS — Gestion des Accès & Profils</title>
  <link rel="stylesheet" href="/src/styles/main.css" />
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>`);

// 3. src/styles/main.css
fs.writeFileSync(`${PROJECT_DIR}/src/styles/main.css`, `:root {
  --bg-canvas: #f8fafc;
  --bg-surface: #ffffff;
  --bg-subtle: #f1f5f9;
  --border-muted: #e2e8f0;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-disabled: #94a3b8;

  --status-valid-bg: #dcfce7;
  --status-valid-fg: #15803d;
  --status-draft-bg: #fef3c7;
  --status-draft-fg: #b45309;
  --status-rejected-bg: #fee2e2;
  --status-rejected-fg: #b91c1c;
  --status-transit-bg: #e0f2fe;
  --status-transit-fg: #0369a1;

  --accent-primary: #2563eb;
  --accent-hover: #1d4ed8;
}

[data-role="CLIENT"]  { --accent-primary: #2563eb; --accent-hover: #1d4ed8; }
[data-role="PARTNER"] { --accent-primary: #7c3aed; --accent-hover: #6d28d9; }
[data-role="ADMIN"]   { --accent-primary: #ea580c; --accent-hover: #c2410c; }
[data-role="OPS"]     { --accent-primary: #059669; --accent-hover: #047857; }
[data-role="MCO"]     { --accent-primary: #475569; --accent-hover: #334155; }
[data-role="RND"]     { --accent-primary: #4f46e5; --accent-hover: #4338ca; }
[data-role="DEPLOY"]  { --accent-primary: #0d9488; --accent-hover: #0f766e; }

[data-theme="dark"],
[data-role="OPS"], [data-role="MCO"], [data-role="RND"], [data-role="DEPLOY"] {
  --bg-canvas: #090d16;
  --bg-surface: #111827;
  --bg-subtle: #1f2937;
  --border-muted: #374151;
  --text-primary: #f9fafb;
  --text-secondary: #9ca3af;
  --text-disabled: #4b5563;
}

* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: var(--bg-canvas); color: var(--text-primary); }

.layout-shell { display: flex; height: 100vh; overflow: hidden; }
.sidebar { width: 270px; background: var(--bg-surface); border-right: 1px solid var(--border-muted); display: flex; flex-direction: column; justify-content: space-between; }
.brand { padding: 20px; font-weight: 700; font-size: 1.15rem; border-bottom: 1px solid var(--border-muted); color: var(--accent-primary); }
.nav-menu { list-style: none; padding: 15px 10px; display: flex; flex-direction: column; gap: 6px; }
.nav-menu a { display: block; padding: 10px 14px; border-radius: 6px; text-decoration: none; color: var(--text-secondary); font-size: 0.875rem; font-weight: 500; cursor: pointer; }
.nav-menu a.active, .nav-menu a:hover { background: var(--bg-subtle); color: var(--accent-primary); }

.main-content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.topbar { height: 60px; background: var(--bg-surface); border-bottom: 1px solid var(--border-muted); display: flex; align-items: center; justify-content: space-between; padding: 0 24px; }
.content-body { flex: 1; overflow-y: auto; padding: 24px; }

.badge { display: inline-flex; align-items: center; padding: 3px 8px; border-radius: 4px; font-size: 0.75rem; font-weight: 600; }
.badge-valid { background: var(--status-valid-bg); color: var(--status-valid-fg); }
.badge-draft { background: var(--status-draft-bg); color: var(--status-draft-fg); }
.badge-rejected { background: var(--status-rejected-bg); color: var(--status-rejected-fg); }
.badge-transit { background: var(--status-transit-bg); color: var(--status-transit-fg); }

.btn { display: inline-flex; align-items: center; justify-content: center; padding: 8px 14px; border-radius: 6px; font-size: 0.85rem; font-weight: 500; border: none; cursor: pointer; }
.btn-primary { background: var(--accent-primary); color: #fff; }
.btn-danger { background: #dc2626; color: #fff; }
.btn-subtle { background: var(--bg-subtle); color: var(--text-primary); }

.card { background: var(--bg-surface); border: 1px solid var(--border-muted); border-radius: 8px; padding: 20px; margin-bottom: 20px; }
.table { width: 100%; border-collapse: collapse; text-align: left; font-size: 0.875rem; }
.table th, .table td { padding: 12px; border-bottom: 1px solid var(--border-muted); }
.table th { color: var(--text-secondary); font-weight: 600; }
`);

// 4. src/config/roles.js
fs.writeFileSync(`${PROJECT_DIR}/src/config/roles.js`, `export const ROLES_CONFIG = {
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
`);

// 5. src/auth/rbac.js
fs.writeFileSync(`${PROJECT_DIR}/src/auth/rbac.js`, `export const USERS_STORE = [
  { id: "usr_001", name: "Jean Dupont", email: "jean.dupont@client.fr", role: "CLIENT", tenant: "dupont-fr", status: "ACTIF" },
  { id: "usr_002", name: "Claire Valette", email: "c.valette@nexus-partner.io", role: "PARTNER", tenant: "nexus-audit", status: "ACTIF" },
  { id: "usr_003", name: "Marc Lefebvre", email: "m.lefebvre@einvoice-saas.com", role: "ADMIN", tenant: "master", status: "ACTIF" },
  { id: "usr_004", name: "Alexandre Roux", email: "ops@einvoice-saas.com", role: "OPS", tenant: "master", status: "ACTIF" },
  { id: "usr_005", name: "Équipe MCO", email: "mco-support@einvoice-saas.com", role: "MCO", tenant: "master", status: "ACTIF" },
  { id: "usr_006", name: "Ingénierie R&D", email: "rnd-engine@einvoice-saas.com", role: "RND", tenant: "master", status: "ACTIF" },
  { id: "usr_007", name: "Release Lead", email: "deploy-bot@einvoice-saas.com", role: "DEPLOY", tenant: "master", status: "ACTIF" }
];

export function hasAccess(userRole, requiredRole) {
  if (userRole === "ADMIN") return true;
  return userRole === requiredRole;
}
`);

// 6. src/components/AppShell.js
fs.writeFileSync(`${PROJECT_DIR}/src/components/AppShell.js`, `import { ROLES_CONFIG } from '../config/roles.js';

export function renderAppShell(activeRole, activeView, contentHtml) {
  const currentRoleCfg = ROLES_CONFIG[activeRole];
  const navItems = currentRoleCfg.routes.map(r => 
    \`<li><a class="\${r.id === activeView ? 'active' : ''}" data-nav="\${r.id}">\${r.label}</a></li>\`
  ).join('');

  const roleOptions = Object.keys(ROLES_CONFIG).map(roleKey => 
    \`<option value="\${roleKey}" \${roleKey === activeRole ? 'selected' : ''}>\${ROLES_CONFIG[roleKey].label} (\${ROLES_CONFIG[roleKey].path})</option>\`
  ).join('');

  return \`
    <div class="layout-shell">
      <aside class="sidebar">
        <div>
          <div class="brand">e-invoice-saas</div>
          <div style="padding:12px;">
            <label style="font-size:0.75rem; color:var(--text-secondary); font-weight:600; text-transform:uppercase;">Changer d'espace :</label>
            <select id="role-selector" style="width:100%; margin-top:6px; padding:8px; border-radius:6px; border:1px solid var(--border-muted); background:var(--bg-subtle); color:var(--text-primary); font-size:0.825rem;">
              \${roleOptions}
            </select>
          </div>
          <ul class="nav-menu">
            \${navItems}
          </ul>
        </div>
        <div style="padding:16px; border-top:1px solid var(--border-muted); font-size:0.8rem;">
          <div style="color:var(--text-secondary);">Espace Actif</div>
          <div style="font-weight:600; color:var(--accent-primary);">\${currentRoleCfg.label}</div>
          <div style="font-size:0.75rem; color:var(--text-disabled); margin-top:4px;">Accès : \${currentRoleCfg.path}</div>
        </div>
      </aside>
      <main class="main-content">
        <header class="topbar">
          <div>
            <span style="font-size:0.875rem; color:var(--text-secondary);">\${currentRoleCfg.description}</span>
          </div>
          <div style="display:flex; align-items:center; gap:12px;">
            <span class="badge badge-valid">Système Opérationnel</span>
            <button id="theme-toggle" class="btn btn-subtle" style="font-size:0.75rem;">Bascule Clair/Sombre</button>
          </div>
        </header>
        <div class="content-body">
          \${contentHtml}
        </div>
      </main>
    </div>
  \`;
}
`);

// 7. src/views/AdminUsersView.js
fs.writeFileSync(`${PROJECT_DIR}/src/views/AdminUsersView.js`, `import { USERS_STORE } from '../auth/rbac.js';

export function renderAdminUsersView() {
  const rows = USERS_STORE.map(u => \`
    <tr>
      <td><b>\${u.name}</b></td>
      <td>\${u.email}</td>
      <td><span class="badge badge-valid">\${u.role}</span></td>
      <td><code>\${u.tenant}</code></td>
      <td><span class="badge badge-valid">\${u.status}</span></td>
      <td>
        <button class="btn btn-subtle" style="padding:4px 8px;" onclick="window.triggerRevoke('\${u.id}')">Révoquer</button>
      </td>
    </tr>
  \`).join('');

  return \`
    <div class="card">
      <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:16px;">
        <div>
          <h2>Gestion des Profils & Habilitations</h2>
          <p style="font-size:0.85rem; color:var(--text-secondary); margin-top:4px;">Configuration des accès RBAC multi-tenants</p>
        </div>
        <button class="btn btn-primary">+ Créer un Utilisateur</button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>Utilisateur</th>
            <th>Email</th>
            <th>Rôle Assigné</th>
            <th>Tenant Rattaché</th>
            <th>Statut</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          \${rows}
        </tbody>
      </table>
    </div>
  \`;
}
`);

// 8. src/views/GenericView.js
fs.writeFileSync(`${PROJECT_DIR}/src/views/GenericView.js`, `export function renderGenericView(roleKey, viewId) {
  return \`
    <div class="card">
      <h2>Module \${viewId}</h2>
      <p style="margin-top:8px; color:var(--text-secondary);">
        Interface opérationnelle pour le profil <b>\${roleKey}</b>.
      </p>
      <div style="margin-top:20px; padding:16px; background:var(--bg-subtle); border-radius:6px; font-family:monospace; font-size:0.85rem;">
        CONTEXT_ROLE: \${roleKey}<br>
        ENDPOINT_BINDING: /api/v1/\${roleKey.toLowerCase()}/\${viewId}<br>
        ENFORCE_SCHEMA: EN_16931_CIUS_FR
      </div>
    </div>
  \`;
}
`);

// 9. src/main.js
fs.writeFileSync(`${PROJECT_DIR}/src/main.js`, `import { ROLES_CONFIG } from './config/roles.js';
import { renderAppShell } from './components/AppShell.js';
import { renderAdminUsersView } from './views/AdminUsersView.js';
import { renderGenericView } from './views/GenericView.js';

let state = {
  role: 'CLIENT',
  view: 'client-upload',
  theme: 'light'
};

function render() {
  document.documentElement.setAttribute('data-role', state.role);
  document.documentElement.setAttribute('data-theme', state.theme);

  let viewHtml = '';
  if (state.role === 'ADMIN' && state.view === 'admin-users') {
    viewHtml = renderAdminUsersView();
  } else {
    viewHtml = renderGenericView(state.role, state.view);
  }

  document.getElementById('app').innerHTML = renderAppShell(state.role, state.view, viewHtml);

  // Événements
  document.getElementById('role-selector').addEventListener('change', (e) => {
    state.role = e.target.value;
    state.view = ROLES_CONFIG[state.role].routes[0].id;
    state.theme = ROLES_CONFIG[state.role].defaultTheme;
    render();
  });

  document.querySelectorAll('[data-nav]').forEach(el => {
    el.addEventListener('click', (e) => {
      state.view = e.target.getAttribute('data-nav');
      render();
    });
  });

  document.getElementById('theme-toggle').addEventListener('click', () => {
    state.theme = state.theme === 'light' ? 'dark' : 'light';
    render();
  });
}

window.triggerRevoke = function(userId) {
  if (confirm(\`Voulez-vous révoquer les accès de l'identifiant \${userId} ?\`)) {
    alert(\`Accès de l'utilisateur \${userId} révoqué.\`);
  }
};

render();
`);

// Création du ZIP
try {
  execSync(`zip -r e-invoice-saas-auth-matrix.zip ${PROJECT_DIR}`);
  console.log(`\n Archive générée avec succès : e-invoice-saas-auth-matrix.zip`);
} catch (e) {
  console.log(`\n Les fichiers ont été écrits dans le dossier "${PROJECT_DIR}".`);
  console.log(`Vous pouvez créer l'archive manuellement avec la commande : zip -r e-invoice-saas-auth-matrix.zip ${PROJECT_DIR}`);
}