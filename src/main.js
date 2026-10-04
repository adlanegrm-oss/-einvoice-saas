import { ROLES_CONFIG } from './config/roles.js';
import { renderAppShell } from './components/AppShell.js';
import { renderAdminUsersView } from './views/AdminUsersView.js';
import { renderDeployView } from './views/DeployView.js';
import { renderClientUploadView } from './views/ClientUploadView.js';
import { renderInvoicesView, INVOICES_DATA } from './modules/invoices/InvoicesView.js';
import { renderOrganizationView } from './modules/organization/OrganizationView.js';
import { renderCustomersView } from './modules/customers/CustomersView.js';
import { renderAuditView } from './modules/audit/AuditView.js';
import { renderGenericView } from './views/GenericView.js';
import { logAuditEvent } from './services/auditLogger.js';

let currentUser = {
  id: "usr_003",
  name: "Marc Lefebvre",
  email: "m.lefebvre@company.fr",
  role: "ADMIN",
  establishmentId: "EST-01"
};

let state = {
  role: 'CLIENT',
  view: 'invoices-list',
  theme: 'light'
};

function render() {
  document.documentElement.setAttribute('data-role', state.role);
  document.documentElement.setAttribute('data-theme', state.theme);

  let viewHtml = '';
  if (state.view === 'invoices-list') {
    viewHtml = renderInvoicesView(currentUser);
  } else if (state.view === 'client-upload') {
    viewHtml = renderClientUploadView();
  } else if (state.view === 'org-profile') {
    viewHtml = renderOrganizationView();
  } else if (state.view === 'customers-list') {
    viewHtml = renderCustomersView();
  } else if (state.view === 'audit-logs') {
    viewHtml = renderAuditView();
  } else if (state.role === 'ADMIN' && state.view === 'admin-users') {
    viewHtml = renderAdminUsersView();
  } else if (state.role === 'DEPLOY' && state.view === 'deploy-envs') {
    viewHtml = renderDeployView();
  } else {
    viewHtml = renderGenericView(state.role, state.view);
  }

  document.getElementById('app').innerHTML = renderAppShell(state.role, state.view, viewHtml);

  // Bascule d'espace
  document.getElementById('role-selector').addEventListener('change', (e) => {
    state.role = e.target.value;
    state.view = ROLES_CONFIG[state.role].routes[0].id;
    state.theme = ROLES_CONFIG[state.role].defaultTheme;
    render();
  });

  // Navigation dans les sous-menus
  document.querySelectorAll('[data-nav]').forEach(el => {
    el.addEventListener('click', (e) => {
      state.view = e.target.getAttribute('data-nav');
      render();
    });
  });

  // Thème Clair / Sombre
  document.getElementById('theme-toggle').addEventListener('click', () => {
    state.theme = state.theme === 'light' ? 'dark' : 'light';
    render();
  });
}

// Machine d'états des factures avec journalisation immédiate
window.transitionInvoice = function(invoiceId, nextStatus) {
  const inv = INVOICES_DATA.find(i => i.id === invoiceId);
  if (!inv) return;

  const previousStatus = inv.status;
  inv.status = nextStatus;

  logAuditEvent({
    userId: currentUser.id,
    userEmail: currentUser.email,
    action: `TRANSITION_TO_${nextStatus}`,
    target: inv.id,
    establishment: inv.establishmentName || inv.establishmentId,
    before: { status: previousStatus },
    after: { status: nextStatus }
  });

  render();
};

window.toggleDrawer = function(open) {
  const drawer = document.getElementById('detail-drawer');
  if (drawer) drawer.classList.toggle('open', open);
};

window.triggerRollback = function(env, targetVersion) {
  const confirmation = prompt(`[ALERTE SÉCURITÉ] Confirmation de rollback sur ${env}.\nTapez ROLLBACK-${env} :`);
  if (confirmation === `ROLLBACK-${env}`) {
    alert(`Rollback vers ${targetVersion} déclenché.`);
  }
};

render();
