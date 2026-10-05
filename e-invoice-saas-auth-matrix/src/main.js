import { ROLES_CONFIG } from './config/roles.js';
import { renderAppShell } from './components/AppShell.js';
import { renderAdminUsersView } from './views/AdminUsersView.js';
import { renderGenericView } from './views/GenericView.js';
import { initDropZone } from './components/DropZone.js';

let state = {
  role: 'CLIENT',
  view: 'client-upload',
  theme: 'light'
};

function render() {
  document.documentElement.setAttribute('data-role', state.role);
  document.documentElement.setAttribute('data-theme', state.theme);

  const currentRoleCfg = ROLES_CONFIG[state.role];
  const currentRoute = currentRoleCfg.routes.find(r => r.id === state.view) || currentRoleCfg.routes[0];

  let viewHtml = '';
  if (state.role === 'ADMIN' && state.view === 'admin-users') {
    viewHtml = renderAdminUsersView();
  } else {
    // Ordre aligné: moduleKey, moduleTitle, roleKey, roleCfg
    viewHtml = renderGenericView(state.view, currentRoute.label, state.role, currentRoleCfg);
  }

  document.getElementById('app').innerHTML = renderAppShell(state.role, state.view, viewHtml);

  // Initialisation post-rendu pour le dépôt client
  if (state.role === 'CLIENT' && state.view === 'client-upload') {
    initDropZone('dropzone-root');
  }

  // Événements de navigation
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
  if (confirm(`Voulez-vous révoquer les accès de l'identifiant ${userId} ?`)) {
    alert(`Accès de l'utilisateur ${userId} révoqué.`);
  }
};

render();
