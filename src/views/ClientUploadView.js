export function renderClientUploadView() {
  return `
    <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px;">
      <div>
        <h2 style="font-size: 1.5rem; font-weight: 700; color: var(--text-primary);">Dépôt & Contrôle de Conformité</h2>
        <p style="color: var(--text-secondary); font-size: 0.875rem; margin-top: 4px;">Validation instantanée selon les règles EN 16931 et profils CIUS-FR (UBL, CII, Factur-X).</p>
      </div>
      <div style="display: flex; gap: 8px;">
        <span class="badge badge-valid">Quota Mensuel : 82%</span>
      </div>
    </div>

    <!-- Mode de traitement -->
    <div class="card" style="display: flex; gap: 24px; align-items: center; padding: 14px 20px;">
      <span style="font-size: 0.875rem; font-weight: 600;">Mode d'opération :</span>
      <label style="display: flex; align-items: center; gap: 8px; font-size: 0.875rem; cursor: pointer;">
        <input type="radio" name="upload-mode" checked /> Enregistrement & Archivage PDP
      </label>
      <label style="display: flex; align-items: center; gap: 8px; font-size: 0.875rem; cursor: pointer;">
        <input type="radio" name="upload-mode" /> Validation seule (Test Schematron sans stockage)
      </label>
      <label style="display: flex; align-items: center; gap: 8px; font-size: 0.875rem; cursor: pointer;">
        <input type="radio" name="upload-mode" /> Brouillon temporaire (72h)
      </label>
    </div>

    <!-- Zone Drag & Drop -->
    <div class="dropzone" id="drop-area" style="margin-bottom: 24px;">
      <div style="font-size: 2rem; margin-bottom: 8px;">📄</div>
      <p style="font-weight: 600; font-size: 1rem; color: var(--text-primary);">Glissez-déposez vos factures ici (XML, Factur-X PDF, EDI)</p>
      <p style="font-size: 0.8rem; color: var(--text-secondary); margin-top: 4px;">ou <span style="color: var(--accent-primary); text-decoration: underline; font-weight: 500;">parcourez vos fichiers</span> (Lots jusqu'à 50 Mo)</p>
    </div>

    <!-- Rapport de validation en direct -->
    <div class="card">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h3 style="font-size: 1rem; font-weight: 600;">Analyse du lot en cours (3 documents détectés)</h3>
        <button class="btn btn-primary" onclick="alert('Traitement et émission du lot initiés...')">Émettre les factures conformes (2)</button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>Nom de Fichier</th>
            <th>Format Détecté</th>
            <th>Statut Conformité</th>
            <th>Diagnostics Règles BR-xx / CIUS-FR</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>FA-2026-0091.pdf</code></td>
            <td>Factur-X (BASIC-WL)</td>
            <td><span class="badge badge-valid">CONFORME</span></td>
            <td>Validation syntaxique et sémantique EN 16931 réussie</td>
            <td><button class="btn btn-subtle" style="padding: 4px 8px;" onclick="window.toggleDrawer(true)">Inspecter</button></td>
          </tr>
          <tr>
            <td><code>INV-FR-49102.xml</code></td>
            <td>UBL 2.1 (Invoice)</td>
            <td><span class="badge badge-rejected">REJETÉ</span></td>
            <td><b style="color: #dc2626;">Erreur BR-CO-15 :</b> Invoice total VAT amount mismatch (ligne 142)</td>
            <td><button class="btn btn-subtle" style="padding: 4px 8px;" onclick="alert('Détails erreur : BR-CO-15 attendait 240.00 EUR, obtenu 239.98 EUR')">Corriger</button></td>
          </tr>
          <tr>
            <td><code>LOT-2901-CII.xml</code></td>
            <td>CII (D16B)</td>
            <td><span class="badge badge-valid">CONFORME</span></td>
            <td>Extension CIUS-FR profil Facture standard validée</td>
            <td><button class="btn btn-subtle" style="padding: 4px 8px;" onclick="window.toggleDrawer(true)">Inspecter</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Drawer d'audit cryptographique -->
    <div id="detail-drawer" class="drawer">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
        <h3 style="font-size: 1.1rem;">Facture FA-2026-0091</h3>
        <button class="btn btn-subtle" style="padding: 4px 8px;" onclick="window.toggleDrawer(false)">✕</button>
      </div>
      <div style="font-size: 0.85rem; line-height: 1.6;">
        <p><b>UUID :</b> <code>4f3b7d12-9c31-4822-a9b0-23098f9aef01</code></p>
        <p style="word-break: break-all; margin-top: 6px;"><b>Empreinte SHA-256 :</b><br><code>e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855</code></p>
        <hr style="margin: 16px 0; border: 0; border-top: 1px solid var(--border-muted);" />
        <h4 style="margin-bottom: 8px;">Audit Trail Immuable (Horodatage RFC 3161)</h4>
        <ul style="padding-left: 18px; color: var(--text-secondary);">
          <li>18:14:02 — Déposé via Web UI par <code>user@acme.fr</code></li>
          <li>18:14:03 — Validation Schematron EN 16931 OK</li>
          <li>18:14:04 — Cachet serveur apposé (Certificat eIDAS)</li>
          <li>18:14:05 — Ancrage Merkle Tree bloc #491,204</li>
        </ul>
        <hr style="margin: 16px 0; border: 0; border-top: 1px solid var(--border-muted);" />
        <h4 style="margin-bottom: 12px;">Téléchargements</h4>
        <div style="display: flex; flex-direction: column; gap: 8px;">
          <button class="btn btn-subtle" onclick="alert('Téléchargement PDF Factur-X...')">Télécharger Factur-X hybride (PDF/A-3)</button>
          <button class="btn btn-subtle" onclick="alert('Téléchargement XML CII extrait...')">Télécharger XML CII intégré</button>
          <button class="btn btn-subtle" onclick="alert('Téléchargement jeton d\'horodatage...')">Jeton de preuve RFC 3161 (.tsr)</button>
        </div>
      </div>
    </div>
  `;
}
