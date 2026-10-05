export function renderDeployView() {
  return `
    <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px;">
      <div>
        <h2 style="font-size: 1.5rem; font-weight: 700; color: var(--text-primary);">Statut des Environnements & Releases</h2>
        <p style="color: var(--text-secondary); font-size: 0.875rem; margin-top: 4px;">Supervision du cycle de vie CI/CD, qualification des artefacts et politique de rollback.</p>
      </div>
      <button class="btn btn-primary" onclick="alert('Démarrage du pipeline de release...')">🚀 Nouveau Déploiement</button>
    </div>

    <!-- Matrice des 4 environnements -->
    <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px;">
      <div class="card" style="border-top: 4px solid var(--accent-primary);">
        <div style="display:flex; justify-content:space-between; align-items:center;">
          <h4 style="color: var(--text-secondary); text-transform: uppercase; font-size: 0.75rem; letter-spacing: 0.05em;">DEV</h4>
          <span class="badge badge-valid">STABLE</span>
        </div>
        <div style="font-size: 1.25rem; font-weight: 700; margin: 12px 0 4px 0;">v2.5.0-rc.3</div>
        <p style="font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 12px;">Commit: <code>#a8f4c10</code></p>
        <span style="font-size: 0.75rem; color: var(--text-secondary);">Déployé il y a 38 min</span>
      </div>

      <div class="card" style="border-top: 4px solid var(--accent-primary);">
        <div style="display:flex; justify-content:space-between; align-items:center;">
          <h4 style="color: var(--text-secondary); text-transform: uppercase; font-size: 0.75rem; letter-spacing: 0.05em;">RECETTE</h4>
          <span class="badge badge-valid">VALIDE</span>
        </div>
        <div style="font-size: 1.25rem; font-weight: 700; margin: 12px 0 4px 0;">v2.4.9</div>
        <p style="font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 12px;">Commit: <code>#3b89e21</code></p>
        <span style="font-size: 0.75rem; color: var(--text-secondary);">Tests automatisés : 100%</span>
      </div>

      <div class="card" style="border-top: 4px solid var(--accent-primary);">
        <div style="display:flex; justify-content:space-between; align-items:center;">
          <h4 style="color: var(--text-secondary); text-transform: uppercase; font-size: 0.75rem; letter-spacing: 0.05em;">PREPROD</h4>
          <span class="badge badge-transit">SYNC</span>
        </div>
        <div style="font-size: 1.25rem; font-weight: 700; margin: 12px 0 4px 0;">v2.4.8</div>
        <p style="font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 12px;">Commit: <code>#910dc84</code></p>
        <span style="font-size: 0.75rem; color: var(--text-secondary);">Conforme CIUS-FR v2.0</span>
      </div>

      <div class="card" style="border-top: 4px solid #10b981;">
        <div style="display:flex; justify-content:space-between; align-items:center;">
          <h4 style="color: var(--text-secondary); text-transform: uppercase; font-size: 0.75rem; letter-spacing: 0.05em;">PROD</h4>
          <span class="badge badge-valid">ACTIF</span>
        </div>
        <div style="font-size: 1.25rem; font-weight: 700; margin: 12px 0 4px 0;">v2.4.8</div>
        <p style="font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 12px;">Commit: <code>#910dc84</code></p>
        <div style="display:flex; gap:8px;">
          <button class="btn btn-danger" style="padding: 4px 8px; font-size: 0.75rem;" onclick="window.triggerRollback('PROD', 'v2.4.7')">Rollback v2.4.7</button>
        </div>
      </div>
    </div>

    <!-- Pipeline CI/CD et historique -->
    <div class="card">
      <h3 style="margin-bottom: 16px; font-size: 1rem;">Pipelines CI/CD & Scans de conformité</h3>
      <table class="table">
        <thead>
          <tr>
            <th>Branche / Release</th>
            <th>Auteur</th>
            <th>Scan Vulnérabilités (Trivy)</th>
            <th>Schematron Validation</th>
            <th>Statut Build</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>release/v2.5.0-rc.3</code></td>
            <td>deploy-bot</td>
            <td><span class="badge badge-valid">0 High / 0 Critical</span></td>
            <td><span class="badge badge-valid">100% EN 16931</span></td>
            <td><span class="badge badge-valid">SUCCÈS (4m 12s)</span></td>
            <td><button class="btn btn-subtle" style="padding: 4px 8px;" onclick="alert('Déploiement en RECETTE initié...')">Promouvoir en Recette</button></td>
          </tr>
          <tr>
            <td><code>feat/cius-fr-patch-3</code></td>
            <td>rnd-engine</td>
            <td><span class="badge badge-draft">1 Medium</span></td>
            <td><span class="badge badge-valid">100% CIUS-FR</span></td>
            <td><span class="badge badge-valid">SUCCÈS (3m 48s)</span></td>
            <td><button class="btn btn-subtle" style="padding: 4px 8px;" disabled>Artefact verrouillé</button></td>
          </tr>
          <tr>
            <td><code>hotfix/parser-vat-calc</code></td>
            <td>m.lefebvre</td>
            <td><span class="badge badge-rejected">1 High</span></td>
            <td><span class="badge badge-rejected">Échec Schematron BR-CO-15</span></td>
            <td><span class="badge badge-rejected">ÉCHEC</span></td>
            <td><button class="btn btn-subtle" style="padding: 4px 8px;" onclick="alert('Affichage des logs d\'exécution...')">Voir logs</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Checklist de Mise en Production (CAB) -->
    <div class="card">
      <h3 style="margin-bottom: 12px; font-size: 1rem;">Checklist de Déploiement en Production (RFC 3161 & PDP)</h3>
      <div style="display:flex; flex-direction:column; gap:10px; font-size: 0.875rem;">
        <label style="display:flex; align-items:center; gap:8px;">
          <input type="checkbox" checked disabled /> Migration schéma Base de Données (Flyway / Liquibase) validée sans locks exclusifs
        </label>
        <label style="display:flex; align-items:center; gap:8px;">
          <input type="checkbox" checked disabled /> Audit d'intégrité des règles Schematron conforme ISO/IEC 19757-3
        </label>
        <label style="display:flex; align-items:center; gap:8px;">
          <input type="checkbox" checked disabled /> Certificat X.509 de signature et cachet serveur actif (validité > 60j)
        </label>
        <label style="display:flex; align-items:center; gap:8px;">
          <input type="checkbox" /> Validation formelle de bascule (CAB - Change Advisory Board)
        </label>
      </div>
    </div>
  `;
}
