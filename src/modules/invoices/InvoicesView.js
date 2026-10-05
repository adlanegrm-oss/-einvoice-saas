import { can, PERMISSIONS } from '../../services/policyEngine.js';
import { auditLogger } from '../../services/auditLogger.js';
import { InvoiceCreateModal } from './InvoiceCreateModal.js';

export class InvoicesView {
  constructor(appContext) {
    this.ctx = appContext;
    this.invoices = [
      { id: 'INV-2026-001', establishmentId: 'est-paris', customerId: 'cust-acme', amount: 14400.00, status: 'VALIDATED', pdpRoute: 'CHORUS_PRO', date: '2026-10-01' },
      { id: 'INV-2026-002', establishmentId: 'est-lyon', customerId: 'cust-globex', amount: 3250.50, status: 'PENDING_VALIDATION', pdpRoute: 'PEPPOL_FR', date: '2026-10-03' },
      { id: 'INV-2026-003', establishmentId: 'est-paris', customerId: 'cust-acme', amount: 890.00, status: 'ISSUED', pdpRoute: 'CHORUS_PRO', date: '2026-09-28' },
    ];
  }

  render() {
    const container = document.createElement('div');
    container.className = 'module-view';

    const canCreate = can(this.ctx.user, PERMISSIONS.INVOICE_CREATE);

    container.innerHTML = `
      <div class="module-header" style="display:flex; justify-content:space-between; align-items:center; margin-bottom:1.5rem;">
        <div>
          <h2>Gestion des Factures Électroniques</h2>
          <p class="text-muted">Conformité EN 16931 & CIUS-FR — Cycle de vie & Machine d'états</p>
        </div>
        ${canCreate ? '<button class="btn btn-primary" id="btn-open-create-invoice">+ Nouvelle Facture</button>' : ''}
      </div>

      <div class="table-container">
        <table class="data-table" style="width:100%; border-collapse:collapse;">
          <thead>
            <tr>
              <th>Identifiant</th>
              <th>Établissement</th>
              <th>Montant TTC</th>
              <th>Statut</th>
              <th>Acheminement</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody id="invoices-tbody">
            ${this.renderRows()}
          </tbody>
        </table>
      </div>
    `;

    this.bindEvents(container);
    return container;
  }

  renderRows() {
    return this.invoices.map(inv => `
      <tr style="border-bottom:1px solid #e2e8f0;">
        <td style="font-weight:600;">${inv.id}</td>
        <td>${inv.establishmentId}</td>
        <td>${inv.amount.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })}</td>
        <td><span class="badge badge-${inv.status.toLowerCase()}">${inv.status}</span></td>
        <td><code>${inv.pdpRoute}</code></td>
        <td>
          <button class="btn btn-secondary btn-sm" onclick="alert('Audit logs de la facture ${inv.id}')">Audit</button>
        </td>
      </tr>
    `).join('');
  }

  bindEvents(container) {
    container.querySelector('#btn-open-create-invoice')?.addEventListener('click', () => {
      const dummyEstablishments = [
        { id: 'est-paris', name: 'Siège Social Paris', siret: '80943210900012' },
        { id: 'est-lyon', name: 'Succursale Rhône-Alpes', siret: '80943210900020' }
      ];
      const dummyCustomers = [
        { id: 'cust-acme', name: 'ACME Corporation SA', siren: '552032541' },
        { id: 'cust-globex', name: 'Globex Logistics SARL', siren: '912845672' }
      ];

      const modalInstance = new InvoiceCreateModal({
        establishments: dummyEstablishments,
        customers: dummyCustomers,
        onSave: (newInvoice) => {
          this.invoices.unshift(newInvoice);
          const tbody = container.querySelector('#invoices-tbody');
          if (tbody) tbody.innerHTML = this.renderRows();
        },
        onClose: () => {
          document.getElementById('invoice-create-modal')?.remove();
        }
      });

      document.body.appendChild(modalInstance.render());
    });
  }
}
