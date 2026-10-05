import { USERS_STORE } from '../auth/rbac.js';

export function renderAdminUsersView() {
  const rows = USERS_STORE.map(u => `
    <tr>
      <td><b>${u.name}</b></td>
      <td>${u.email}</td>
      <td><span class="badge badge-valid">${u.role}</span></td>
      <td><code>${u.tenant}</code></td>
      <td><span class="badge badge-valid">${u.status}</span></td>
      <td>
        <button class="btn btn-subtle" style="padding:4px 8px;" onclick="window.triggerRevoke('${u.id}')">Révoquer</button>
      </td>
    </tr>
  `).join('');

  return `
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
          ${rows}
        </tbody>
      </table>
    </div>
  `;
}
