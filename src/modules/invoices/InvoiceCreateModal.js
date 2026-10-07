import { auditLogger } from '../../services/auditLogger.js';

export class InvoiceCreateModal {
  constructor({ establishments, customers, onSave, onClose }) {
    this.establishments = establishments || [];
    this.customers = customers || [];
    this.onSave = onSave;
    this.onClose = onClose;

    this.state = {
      step: 1,
      establishmentId: this.establishments[0]?.id || '',
      customerId: this.customers[0]?.id || '',
      invoiceNumber: 'INV-' + new Date().getFullYear() + '-' + String(Math.floor(1000 + Math.random() * 9000)),
      issueDate: new Date().toISOString().split('T')[0],
      dueDate: new Date(Date.now() + 30 * 86400000).toISOString().split('T')[0],
      currency: 'EUR',
      paymentTerms: 'Virement bancaire 30 jours net',
      pdpRoute: 'CHORUS_PRO',
      lines: [
        { id: 1, desc: 'Prestation d\'intégration e-Invoicing UBL', qty: 1, unitPrice: 1200.0, vatRate: 20 }
      ],
      legalMentions: 'Dispensé d\'escompte en cas de paiement anticipé. Pénalités de retard : 3x taux légal + 40€.'
    };
  }

  calculateTotals() {
    let totalHT = 0;
    const vatBuckets = {};

    this.state.lines.forEach(l => {
      const lineHT = (Number(l.qty) || 0) * (Number(l.unitPrice) || 0);
      totalHT += lineHT;
      const rate = Number(l.vatRate) || 0;
      vatBuckets[rate] = (vatBuckets[rate] || 0) + (lineHT * (rate / 100));
    });

    const totalVAT = Object.values(vatBuckets).reduce((a, b) => a + b, 0);
    const totalTTC = totalHT + totalVAT;

    return { totalHT, vatBuckets, totalVAT, totalTTC };
  }

  render() {
    const modal = document.createElement('div');
    modal.className = 'invoice-modal-overlay';
    modal.id = 'invoice-create-modal';

    const { totalHT, vatBuckets, totalVAT, totalTTC } = this.calculateTotals();

    modal.innerHTML = `
      <div class="invoice-modal-card">
        <div class="invoice-modal-header">
          <div>
            <h3>Nouvelle Facture Électronique (EN 16931 / CIUS-FR)</h3>
            <p class="text-muted">Parcours d'émission conforme - Étape ${this.state.step} / 3</p>
          </div>
          <button class="btn-close" id="modal-close-btn">&times;</button>
        </div>

        <div class="invoice-modal-body">
          ${this.renderStepContent(totalHT, vatBuckets, totalVAT, totalTTC)}
        </div>

        <div class="invoice-modal-footer">
          <div>
            ${this.state.step > 1 ? '<button class="btn btn-secondary" id="modal-prev-btn">Précédent</button>' : ''}
          </div>
          <div style="display:flex; gap:0.5rem;">
            <button class="btn btn-secondary" id="modal-cancel-btn">Annuler</button>
            ${this.state.step < 3 
              ? '<button class="btn btn-primary" id="modal-next-btn">Suivant &rarr;</button>' 
              : '<button class="btn btn-primary" id="modal-submit-btn">Émettre en Validation</button>'}
          </div>
        </div>
      </div>
    `;

    this.bindEvents(modal);
    return modal;
  }

  renderStepContent(totalHT, vatBuckets, totalVAT, totalTTC) {
    if (this.state.step === 1) {
      return `
        <div class="form-grid">
          <div class="form-group">
            <label>1. Établissement Émetteur (SIRET / Succursale)</label>
            <select class="form-control" id="f-establishment">
              ${this.establishments.map(e => `<option value="${e.id}" ${this.state.establishmentId === e.id ? 'selected' : ''}>${e.name} (${e.siret})</option>`).join('')}
            </select>
          </div>

          <div class="form-group">
            <label>2. Débiteur / Client Destinataire</label>
            <select class="form-control" id="f-customer">
              ${this.customers.map(c => `<option value="${c.id}" ${this.state.customerId === c.id ? 'selected' : ''}>${c.name} (SIREN:${c.siren})</option>`).join('')}
            </select>
          </div>

          <div class="form-group">
            <label>3. Numéro de Facture</label>
            <input type="text" class="form-control" id="f-number" value="${this.state.invoiceNumber}">
          </div>

          <div class="form-group">
            <label>4. Date d'Émission</label>
            <input type="date" class="form-control" id="f-issue-date" value="${this.state.issueDate}">
          </div>

          <div class="form-group">
            <label>5. Échéance de Règlement</label>
            <input type="date" class="form-control" id="f-due-date" value="${this.state.dueDate}">
          </div>

          <div class="form-group">
            <label>6. Acheminement Plateforme (PDP / PPF)</label>
            <select class="form-control" id="f-pdp">
              <option value="CHORUS_PRO" ${this.state.pdpRoute === 'CHORUS_PRO' ? 'selected' : ''}>Chorus Pro (Portail Public)</option>
              <option value="PEPPOL_FR" ${this.state.pdpRoute === 'PEPPOL_FR' ? 'selected' : ''}>Réseau PEPPOL (PDP Partenaire)</option>
            </select>
          </div>
        </div>
      `;
    }

    if (this.state.step === 2) {
      return `
        <div>
          <label style="font-weight:600; margin-bottom:0.5rem; display:block;">7. Lignes de Facturation & Prestations</label>
          <table class="table-modal-lines" style="width:100%; border-collapse:collapse; margin-bottom:1rem;">
            <thead>
              <tr style="border-bottom:1px solid #e2e8f0; text-align:left; font-size:0.85rem; color:#64748b;">
                <th>Description</th>
                <th style="width:80px;">Qté</th>
                <th style="width:110px;">Prix Unit. HT</th>
                <th style="width:90px;">Taux TVA</th>
                <th style="width:100px; text-align:right;">Total HT</th>
                <th style="width:40px;"></th>
              </tr>
            </thead>
            <tbody id="lines-tbody">
              ${this.state.lines.map((l, idx) => `
                <tr style="border-bottom:1px solid #f1f5f9;">
                  <td><input type="text" class="form-control line-desc" data-idx="${idx}" value="${l.desc}"></td>
                  <td><input type="number" class="form-control line-qty" data-idx="${idx}" value="${l.qty}" min="1"></td>
                  <td><input type="number" class="form-control line-price" data-idx="${idx}" value="${l.unitPrice}" step="0.01"></td>
                  <td>
                    <select class="form-control line-vat" data-idx="${idx}">
                      <option value="20" ${l.vatRate === 20 ? 'selected' : ''}>20 %</option>
                      <option value="10" ${l.vatRate === 10 ? 'selected' : ''}>10 %</option>
                      <option value="5.5" ${l.vatRate === 5.5 ? 'selected' : ''}>5.5 %</option>
                      <option value="0" ${l.vatRate === 0 ? 'selected' : ''}>0 %</option>
                    </select>
                  </td>
                  <td style="text-align:right; font-weight:600;">${((l.qty || 0) * (l.unitPrice || 0)).toFixed(2)} €</td>
                  <td><button class="btn-del-line" data-idx="${idx}" style="background:none; border:none; color:#ef4444; cursor:pointer;">&times;</button></td>
                </tr>
              `).join('')}
            </tbody>
          </table>
          <button class="btn btn-secondary" id="btn-add-line" style="font-size:0.85rem; padding:0.35rem 0.75rem;">+ Ajouter une ligne</button>
        </div>

        <div class="totals-summary-box" style="margin-top:1.5rem; background:#f8fafc; padding:1rem; border-radius:6px; display:flex; justify-content:flex-end;">
          <div style="min-width:240px; font-size:0.9rem;">
            <div style="display:flex; justify-content:space-between; margin-bottom:0.25rem;">
              <span>Total Brut HT :</span>
              <strong>${totalHT.toFixed(2)} €</strong>
            </div>
            ${Object.entries(vatBuckets).map(([rate, amt]) => `
              <div style="display:flex; justify-content:space-between; color:#64748b; font-size:0.85rem;">
                <span>TVA à ${rate}% :</span>
                <span>${amt.toFixed(2)} €</span>
              </div>
            `).join('')}
            <div style="display:flex; justify-content:space-between; margin-top:0.5rem; padding-top:0.5rem; border-top:1px solid #cbd5e1; font-size:1.05rem;">
              <span>Total TTC :</span>
              <strong style="color:#2563eb;">${totalTTC.toFixed(2)} €</strong>
            </div>
          </div>
        </div>
      `;
    }

    if (this.state.step === 3) {
      return `
        <div class="form-grid">
          <div class="form-group" style="grid-column: span 2;">
            <label>8. Conditions de Paiement & Modalités</label>
            <input type="text" class="form-control" id="f-payment-terms" value="${this.state.paymentTerms}">
          </div>

          <div class="form-group" style="grid-column: span 2;">
            <label>9. Mentions Légales Obligatoires (CIUS-FR)</label>
            <textarea class="form-control" id="f-legal" rows="3">${this.state.legalMentions}</textarea>
          </div>
        </div>

        <div style="background:#eff6ff; border:1px solid #bfdbfe; padding:1rem; border-radius:6px; margin-top:1rem;">
          <h4 style="margin:0 0 0.5rem 0; color:#1e40af; font-size:0.95rem;">10. Contrôle de Conformité EN 16931 Pré-Émission</h4>
          <ul style="margin:0; padding-left:1.2rem; font-size:0.85rem; color:#1e3a8a;">
            <li>Format pivot cible : UBL 2.1 / Factur-X profil EXTENDED</li>
            <li>Identifiants SIRET vendeur et acheteur validés par l'annuaire</li>
            <li>Règle BR-CO-15 (Égalité stricte HT + TVA = TTC) vérifiée</li>
          </ul>
        </div>
      `;
    }
  }

  bindEvents(modal) {
    const refresh = () => {
      const newModal = this.render();
      modal.replaceWith(newModal);
    };

    modal.querySelector('#modal-close-btn')?.addEventListener('click', () => this.onClose());
    modal.querySelector('#modal-cancel-btn')?.addEventListener('click', () => this.onClose());

    modal.querySelector('#modal-next-btn')?.addEventListener('click', () => {
      this.collectStepData(modal);
      this.state.step++;
      refresh();
    });

    modal.querySelector('#modal-prev-btn')?.addEventListener('click', () => {
      this.collectStepData(modal);
      this.state.step--;
      refresh();
    });

    modal.querySelector('#btn-add-line')?.addEventListener('click', () => {
      this.collectStepData(modal);
      this.state.lines.push({
        id: Date.now(),
        desc: 'Nouvelle ligne de prestation',
        qty: 1,
        unitPrice: 100.0,
        vatRate: 20
      });
      refresh();
    });

    modal.querySelectorAll('.btn-del-line').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const idx = Number(e.target.dataset.idx);
        if (this.state.lines.length > 1) {
          this.state.lines.splice(idx, 1);
          refresh();
        }
      });
    });

    modal.querySelector('#modal-submit-btn')?.addEventListener('click', () => {
      this.collectStepData(modal);
      const totals = this.calculateTotals();
      const newInvoice = {
        id: this.state.invoiceNumber,
        establishmentId: this.state.establishmentId,
        customerId: this.state.customerId,
        amount: totals.totalTTC,
        currency: 'EUR',
        status: 'PENDING_VALIDATION',
        pdpRoute: this.state.pdpRoute,
        issueDate: this.state.issueDate,
        dueDate: this.state.dueDate,
        lines: this.state.lines,
        totals
      };

      auditLogger.log({
        actor: 'Commercial / Facturation',
        action: 'INVOICE_CREATED',
        target: newInvoice.id,
        status: 'SUCCESS',
        metadata: { amount: totals.totalTTC, customerId: newInvoice.customerId }
      });

      this.onSave(newInvoice);
      this.onClose();
    });
  }

  collectStepData(modal) {
    if (this.state.step === 1) {
      const est = modal.querySelector('#f-establishment');
      const cust = modal.querySelector('#f-customer');
      const num = modal.querySelector('#f-number');
      const iss = modal.querySelector('#f-issue-date');
      const due = modal.querySelector('#f-due-date');
      const pdp = modal.querySelector('#f-pdp');

      if (est) this.state.establishmentId = est.value;
      if (cust) this.state.customerId = cust.value;
      if (num) this.state.invoiceNumber = num.value;
      if (iss) this.state.issueDate = iss.value;
      if (due) this.state.dueDate = due.value;
      if (pdp) this.state.pdpRoute = pdp.value;
    }

    if (this.state.step === 2) {
      modal.querySelectorAll('#lines-tbody tr').forEach((row, idx) => {
        const desc = row.querySelector('.line-desc')?.value;
        const qty = Number(row.querySelector('.line-qty')?.value);
        const unitPrice = Number(row.querySelector('.line-price')?.value);
        const vatRate = Number(row.querySelector('.line-vat')?.value);

        if (this.state.lines[idx]) {
          this.state.lines[idx] = { ...this.state.lines[idx], desc, qty, unitPrice, vatRate };
        }
      });
    }

    if (this.state.step === 3) {
      const pt = modal.querySelector('#f-payment-terms');
      const leg = modal.querySelector('#f-legal');
      if (pt) this.state.paymentTerms = pt.value;
      if (leg) this.state.legalMentions = leg.value;
    }
  }
}
