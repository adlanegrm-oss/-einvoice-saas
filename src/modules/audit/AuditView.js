import { getAuditLogs } from '../../services/auditLogger.js';

export function renderAuditView() {
  const logs = getAuditLogs();
  const rows = logs.map(l => `
    <tr>
      <td><code>${l.id}</code></td>
      <td>${l.timestamp}</td>
      <td><b>${l.userEmail}</b></td>
      <td><span class="badge badge-transit">${l.action}</span></td>
      <td><b>${l.target}</b></td>
      <td><span style="font-size:0.75rem;">${l.establishment || 'N/A'}</span></td>
      <td><code style="font-size:0.7rem;">${l.hash}</code></td>
    </tr>
  `).join('');

  return `
    <div style="margin-bottom:20px;">
      <h2 style="font-size:1.4rem; font-weight:700;">Journal d'Audit Immuable (Append-Only)</h2>
      <p style="font-size:0.85rem; color:var(--text-secondary); margin-top:4px;">Traçabilité infalsifiable des transitions d'états, modifications de sécurité et accès.</p>
    </div>

    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th>Event ID</th>
            <th>Horodatage (UTC)</th>
            <th>Auteur (User)</th>
            <th>Action</th>
            <th>Cible</th>
            <th>Établissement</th>
            <th>Empreinte Cryptographique</th>
          </tr>
        </thead>
        <tbody>
          ${rows}
        </tbody>
      </table>
    </div>
  `;
}
