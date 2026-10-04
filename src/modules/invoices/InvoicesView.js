import { can, PERMISSIONS } from '../../services/policyEngine.js';
import { logAuditEvent } from '../../services/auditLogger.js';

export let INVOICES_DATA = [
  {
    id: "INV-2026-00125",
    client: "TechCorp Global SAS",
    establishmentId: "EST-01",
    establishmentName: "Agence Paris Nord",
    createdBy: "usr_001",
    date: "2026-10-04",
    amountHT: 12500.00,
    vatAmount: 2500.00,
    amountTTC: 15000.00,
    status: "PENDING_VALIDATION",
    complianceFormat: "Factur-X (EN 16931 Extended)"
  },
  {
    id: "INV-2026-00124",
    client: "Logistique Moderne SARL",
    establishmentId: "EST-02",
    establishmentName: "Agence Lyon Centre",
    createdBy: "usr_002",
    date: "2026-10-02",
    amountHT: 4200.00,
    vatAmount: 840.00,
    amountTTC: 5040.00,
    status: "VALIDATED",
    complianceFormat: "UBL 2.1 (CIUS-FR)"
  },
  {
    id: "INV-2026-00123",
    client: "Solutions IT Europe",
    establishmentId: "EST-01",
    createdBy: "usr_001",
    date: "2026-09-28",
    amountHT: 8900.00,
    vatAmount: 1780.00,
    amountTTC: 10680.00,
    status: "ISSUED",
    complianceFormat: "CII (D16B)"
  }
];

export function renderInvoicesView(currentUser) {
  const allowedInvoices = INVOICES_DATA.filter(inv => can(currentUser, PERMISSIONS.INVOICE_READ, inv));

  const rows = allowedInvoices.map(inv => {
    const canValidate = can(currentUser, PERMISSIONS.INVOICE_VALIDATE, inv) && inv.status === 'PENDING_VALIDATION';
    const canIssue = can(currentUser, PERMISSIONS.INVOICE_ISSUE, inv) && inv.status === 'VALIDATED';
    const canCancel = can(currentUser, PERMISSIONS.INVOICE_CANCEL, inv) && inv.status !== 'CANCELLED';

    let statusBadge = "badge-draft";
    if (inv.status === "VALIDATED") statusBadge = "badge-valid";
    if (inv.status === "ISSUED") statusBadge = "badge-transit";
    if (inv.status === "CANCELLED") statusBadge = "badge-rejected";

    return `
      <tr>
        <td><b>${inv.id}</b></td>
        <td>${inv.client}</td>
        <td><span style="font-size:0.75rem; color:var(--text-secondary);">${inv.establishmentName || inv.establishmentId}</span></td>
        <td>${inv.amountTTC.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })}</td>
        <td><span class="badge ${statusBadge}">${inv.status}</span></td>
        <td><code style="font-size:0.75rem;">${inv.complianceFormat}</code></td>
        <td style="display:flex; gap:6px;">
          ${canValidate ? `<button class="btn btn-primary" style="padding:4px 8px; font-size:0.75rem;" onclick="window.transitionInvoice('${inv.id}', 'VALIDATED')">Valider</button>` : ''}
          ${canIssue ? `<button class="btn btn-primary" style="padding:4px 8px; font-size:0.75rem; background:#10b981;" onclick="window.transitionInvoice('${inv.id}', 'ISSUED')">Émettre PDP</button>` : ''}
          ${canCancel ? `<button class="btn btn-subtle" style="padding:4px 8px; font-size:0.75rem; color:#dc2626;" onclick="window.transitionInvoice('${inv.id}', 'CANCELLED')">Annuler</button>` : ''}
        </td>
      </tr>
    `;
  }).join('');

  return `
    <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:20px;">
      <div>
        <h2 style="font-size:1.4rem; font-weight:700;">Cycle de Vie de la Facturation</h2>
        <p style="font-size:0.85rem; color:var(--text-secondary); margin-top:4px;">
          Filtre de scope actif : <b>${currentUser.role}</b> (${ROLES_DEFINITION[currentUser.role]?.scope || 'Direct'})
        </p>
      </div>
      ${can(currentUser, PERMISSIONS.INVOICE_CREATE) ? `<button class="btn btn-primary">+ Nouvelle Facture</button>` : `<span class="badge badge-draft">Création non autorisée</span>`}
    </div>

    <!-- Machine d'États Visuelle -->
    <div class="card" style="display:flex; justify-content:space-around; align-items:center; padding:12px; font-size:0.8rem;">
      <span style="color:var(--text-secondary);">1. BROUILLON</span> ➔
      <span style="color:#b45309; font-weight:600;">2. À VALIDER</span> ➔
      <span style="color:#15803d; font-weight:600;">3. VALIDÉE</span> ➔
      <span style="color:#0369a1; font-weight:600;">4. ÉMISE / PDP</span> ➔
      <span style="color:var(--text-secondary);">5. PAYÉE</span>
    </div>

    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th>N° Facture</th>
            <th>Client</th>
            <th>Établissement</th>
            <th>Total TTC</th>
            <th>Statut Workflow</th>
            <th>Norme Électronique</th>
            <th>Actions Habilitées</th>
          </tr>
        </thead>
        <tbody>
          ${rows.length > 0 ? rows : `<tr><td colspan="7" style="text-align:center; padding:24px; color:var(--text-secondary);">Aucune facture accessible pour ce profil et ce périmètre d'établissement.</td></tr>`}
        </tbody>
      </table>
    </div>
  `;
}
