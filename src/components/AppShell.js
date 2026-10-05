import { ROLES_CONFIG } from '../config/roles.js';

export function renderAppShell(activeRole, activeView, contentHtml) {
  const currentRoleCfg = ROLES_CONFIG[activeRole];
  const navItems = currentRoleCfg.routes.map(r => 
    `<li><a class="${r.id === activeView ? 'active' : ''}" data-nav="${r.id}">${r.label}</a></li>`
  ).join('');

  const roleOptions = Object.keys(ROLES_CONFIG).map(roleKey => 
    `<option value="${roleKey}" ${roleKey === activeRole ? 'selected' : ''}>${ROLES_CONFIG[roleKey].label} (${ROLES_CONFIG[roleKey].path})</option>`
  ).join('');

  return `
    <div class="layout-shell">
      <aside class="sidebar">
        <div>
          <div class="brand">e-invoice-saas</div>
          <div style="padding:12px;">
            <label style="font-size:0.75rem; color:var(--text-secondary); font-weight:600; text-transform:uppercase;">Changer d'espace :</label>
            <select id="role-selector" style="width:100%; margin-top:6px; padding:8px; border-radius:6px; border:1px solid var(--border-muted); background:var(--bg-subtle); color:var(--text-primary); font-size:0.825rem;">
              ${roleOptions}
            </select>
          </div>
          <ul class="nav-menu">
            ${navItems}
          </ul>
        </div>
        <div style="padding:16px; border-top:1px solid var(--border-muted); font-size:0.8rem;">
          <div style="color:var(--text-secondary);">Espace Actif</div>
          <div style="font-weight:600; color:var(--accent-primary);">${currentRoleCfg.label}</div>
          <div style="font-size:0.75rem; color:var(--text-disabled); margin-top:4px;">Accès : ${currentRoleCfg.path}</div>
        </div>
      </aside>
      <main class="main-content">
        <header class="topbar">
          <div>
            <span style="font-size:0.875rem; color:var(--text-secondary);">${currentRoleCfg.description}</span>
          </div>
          <div style="display:flex; align-items:center; gap:12px;">
            <span class="badge badge-valid">Système Opérationnel</span>
            <button id="theme-toggle" class="btn btn-subtle" style="font-size:0.75rem;">Bascule Clair/Sombre</button>
          </div>
        </header>
        <div class="content-body">
          ${contentHtml}
        </div>
      </main>
    </div>
  `;
}
