export function renderOrganizationView() {
  return `
    <div style="margin-bottom:20px;">
      <h2 style="font-size:1.4rem; font-weight:700;">Fiche Organisation & Établissements</h2>
      <p style="font-size:0.85rem; color:var(--text-secondary); margin-top:4px;">Entité légale centrale, paramétrage fiscal et succursales rattachées.</p>
    </div>

    <div style="display:grid; grid-template-columns: 2fr 1fr; gap:20px;">
      <div class="card">
        <h3 style="font-size:1rem; margin-bottom:16px;">Informations Légales & Fiscales</h3>
        <div style="display:grid; grid-template-columns: 1fr 1fr; gap:12px; font-size:0.875rem;">
          <div><label style="color:var(--text-secondary); font-size:0.75rem;">Raison Sociale</label><div style="font-weight:600;">ACME Technologies France SAS</div></div>
          <div><label style="color:var(--text-secondary); font-size:0.75rem;">Numéro SIREN / SIRET Siège</label><div>802 910 249 00021</div></div>
          <div><label style="color:var(--text-secondary); font-size:0.75rem;">N° TVA Intracommunautaire</label><div>FR 49 802910249</div></div>
          <div><label style="color:var(--text-secondary); font-size:0.75rem;">Code NAF / APE</label><div>6201Z (Programmation informatique)</div></div>
          <div><label style="color:var(--text-secondary); font-size:0.75rem;">Devise Opérationnelle</label><div>EUR (€) — ISO 4217</div></div>
          <div><label style="color:var(--text-secondary); font-size:0.75rem;">Format Réglementaire par Défaut</label><div>Factur-X / EN 16931</div></div>
        </div>
      </div>

      <div class="card">
        <h3 style="font-size:1rem; margin-bottom:16px;">Paramètres de Facturation</h3>
        <ul style="font-size:0.85rem; line-height:1.8; color:var(--text-secondary); list-style:none;">
          <li>✓ Préfixe légal : <code>INV-2026-</code></li>
          <li>✓ Conditions de règlement : 30 jours net</li>
          <li>✓ Mentions pénalités : Taux BCE + 10 points</li>
          <li>✓ Rapprochement bancaire : SEPA B2B actif</li>
        </ul>
      </div>
    </div>

    <div class="card">
      <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:12px;">
        <h3 style="font-size:1rem;">Établissements & Succursales (Scopes Géographiques)</h3>
        <button class="btn btn-subtle" style="font-size:0.75rem;">+ Ajouter un Établissement</button>
      </div>
      <table class="table">
        <thead>
          <tr><th>ID Établissement</th><th>Nom</th><th>SIRET Spécifique</th><th>Ville</th><th>Utilisateurs Rattachés</th></tr>
        </thead>
        <tbody>
          <tr><td><code>EST-01</code></td><td><b>Agence Paris Nord (Siège)</b></td><td>802 910 249 00021</td><td>Saint-Denis (93)</td><td>8 collaborateurs</td></tr>
          <tr><td><code>EST-02</code></td><td><b>Agence Lyon Centre</b></td><td>802 910 249 00039</td><td>Lyon (69)</td><td>3 collaborateurs</td></tr>
          <tr><td><code>EST-03</code></td><td><b>Plateforme Ouest</b></td><td>802 910 249 00047</td><td>Nantes (44)</td><td>2 collaborateurs</td></tr>
        </tbody>
      </table>
    </div>
  `;
}
