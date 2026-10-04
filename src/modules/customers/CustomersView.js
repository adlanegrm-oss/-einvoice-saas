export function renderCustomersView() {
  return `
    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:20px;">
      <div>
        <h2 style="font-size:1.4rem; font-weight:700;">Annuaire Clients & Débiteurs</h2>
        <p style="font-size:0.85rem; color:var(--text-secondary); margin-top:4px;">Référentiel des tiers avec identifiants fiscaux et routage PDP.</p>
      </div>
      <button class="btn btn-primary">+ Nouveau Client</button>
    </div>

    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th>Identifiant</th>
            <th>Raison Sociale</th>
            <th>N° TVA Intracommunautaire</th>
            <th>Routage PDP / Plateforme</th>
            <th>Conditions de Paiement</th>
            <th>Statut</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>CLI-0091</code></td>
            <td><b>TechCorp Global SAS</b></td>
            <td>FR49849201948</td>
            <td>PDP Partenaire (PPF / Chorus Pro)</td>
            <td>45 jours fin de mois</td>
            <td><span class="badge badge-valid">Vérifié</span></td>
          </tr>
          <tr>
            <td><code>CLI-0092</code></td>
            <td><b>Logistique Moderne SARL</b></td>
            <td>FR12401928401</td>
            <td>Réseau Peppol (ID: 0009:12401928401)</td>
            <td>30 jours net</td>
            <td><span class="badge badge-valid">Vérifié</span></td>
          </tr>
        </tbody>
      </table>
    </div>
  `;
}
