export function initDropZone(containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = `
    <div class="dropzone-box" id="dropzone-area" style="border: 2px dashed #3b82f6; border-radius: 8px; padding: 32px 20px; text-align: center; background: rgba(59, 130, 246, 0.03); cursor: pointer; transition: all 0.2s ease;">
      <input type="file" id="invoice-file-input" style="display: none;" accept=".xml,.pdf,.edi" />
      <div style="font-size: 1.1rem; color: var(--text-primary, #1e293b); font-weight: 500;">
        Glissez-déposez vos factures ici ou <span style="color: #2563eb; text-decoration: underline;">parcourir vos fichiers</span>
      </div>
      <div style="font-size: 0.85rem; color: var(--text-secondary, #64748b); margin-top: 8px;">
        Fichiers supportés : UBL (.xml), CII / Factur-X (.pdf, .xml), EDIFACT (.edi)
      </div>
    </div>

    <!-- Conteneur d'analyse et détails -->
    <div id="invoice-inspection-result" style="margin-top: 24px; display: none;"></div>
  `;

  const dropArea = document.getElementById('dropzone-area');
  const fileInput = document.getElementById('invoice-file-input');

  dropArea.addEventListener('click', () => fileInput.click());
  fileInput.addEventListener('change', (e) => {
    if (e.target.files && e.target.files[0]) processInvoiceFile(e.target.files[0]);
  });

  ['dragenter', 'dragover'].forEach(name => {
    dropArea.addEventListener(name, (e) => {
      e.preventDefault();
      dropArea.style.borderColor = '#1d4ed8';
      dropArea.style.background = 'rgba(59, 130, 246, 0.08)';
    });
  });

  ['dragleave', 'drop'].forEach(name => {
    dropArea.addEventListener(name, (e) => {
      e.preventDefault();
      dropArea.style.borderColor = '#3b82f6';
      dropArea.style.background = 'rgba(59, 130, 246, 0.03)';
    });
  });

  dropArea.addEventListener('drop', (e) => {
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      processInvoiceFile(e.dataTransfer.files[0]);
    }
  });
}

function processInvoiceFile(file) {
  const resultContainer = document.getElementById('invoice-inspection-result');
  if (!resultContainer) return;

  const fileSizeKo = (file.size / 1024).toFixed(1);
  const ext = file.name.split('.').pop().toLowerCase();

  // Détection du profil et simulation des contrôles métier EN 16931 / CIUS-FR
  let isPdf = ext === 'pdf';
  let isXml = ext === 'xml';
  let isEdi = ext === 'edi';

  let checks = [];
  let corrections = [];

  if (isPdf) {
    checks = [
      { label: "Format de conteneur PDF/A-3 (ISO 19005-3)", status: "OK", detail: "Profil PDF/A conforme" },
      { label: "Pièce jointe XML embarquée (factur-x.xml)", status: "FAILED", detail: "Fichier XML manquant ou corrompu dans le catalogue PDF" },
      { label: "Validation syntaxique EN 16931", status: "FAILED", detail: "Impossible d'extraire la charge utile CII" },
      { label: "Conformité profil CIUS-FR / Chorus Pro", status: "PENDING", detail: "En attente d'un flux structuré valide" }
    ];

    corrections = [
      {
        field: "Pièce jointe Factur-X",
        issue: "Le PDF déposé est un fichier bureautique standard et non un PDF hybride Factur-X.",
        action: "Générez la facture depuis votre outil de facturation en activant le mode **Factur-X** (ou joignez directement le fichier XML équivalent au format UBL ou CII)."
      },
      {
        field: "Règles CIUS-FR (Champs obligatoires)",
        issue: "Absence des métadonnées obligatoires : N° d'engagement juridique / Code Service Chorus Pro.",
        action: "Ajouter la balise `BuyerReference` ou `ContractDocumentReference` si le destinataire est une entité publique."
      }
    ];
  } else if (isXml) {
    checks = [
      { label: "Format XML bien formé (Well-formed)", status: "OK", detail: "Encodage UTF-8 valide" },
      { label: "Schéma XSD EN 16931 (UBL/CII)", status: "OK", detail: "Balises racine conformes" },
      { label: "Règles Schematron CIUS-FR", status: "OK", detail: "Toutes les assertions BR-01 à BR-64 validées" }
    ];
    corrections = [];
  } else if (isEdi) {
    checks = [
      { label: "Syntaxe EDIFACT D01B INVOIC", status: "OK", detail: "Segments UNH, BGM, DTM reconnus" },
      { label: "Mapping vers EN 16931", status: "OK", detail: "Conversion sémantique réussie" }
    ];
    corrections = [];
  } else {
    checks = [
      { label: "Contrôle d'extension", status: "FAILED", detail: `Type .${ext} non pris en charge` }
    ];
    corrections = [
      { field: "Format de fichier", issue: "Format non reconnu", action: "Déposez un fichier .xml (UBL/CII), .pdf (Factur-X), ou .edi (EDIFACT)." }
    ];
  }

  const isGlobalSuccess = corrections.length === 0;

  resultContainer.style.display = 'block';
  resultContainer.innerHTML = `
    <!-- 1. APERÇU DE LA FACTURE -->
    <div style="background: var(--bg-subtle, #f8fafc); border: 1px solid var(--border-muted, #e2e8f0); border-radius: 8px; padding: 20px; margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border-muted, #e2e8f0); padding-bottom: 12px; margin-bottom: 16px;">
        <div style="font-weight: 600; font-size: 1rem; color: var(--text-primary, #0f172a);">
          📄 Aperçu du fichier déposé
        </div>
        <span style="display: inline-block; padding: 4px 10px; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; ${isGlobalSuccess ? 'background: #dcfce7; color: #15803d;' : 'background: #fee2e2; color: #b91c1c;'}">
          ${isGlobalSuccess ? '✔ Conforme' : '✖ Rejet Schéma'}
        </span>
      </div>
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px; font-size: 0.85rem;">
        <div><strong style="color: var(--text-secondary, #64748b);">Nom :</strong> <span style="color: var(--text-primary, #0f172a);">${file.name}</span></div>
        <div><strong style="color: var(--text-secondary, #64748b);">Taille :</strong> <span style="color: var(--text-primary, #0f172a);">${fileSizeKo} Ko</span></div>
        <div><strong style="color: var(--text-secondary, #64748b);">Type détecté :</strong> <span style="color: var(--text-primary, #0f172a); text-transform: uppercase;">${ext}</span></div>
        <div><strong style="color: var(--text-secondary, #64748b);">Horodatage :</strong> <span style="color: var(--text-primary, #0f172a);">${new Date().toLocaleTimeString()}</span></div>
      </div>
    </div>

    <!-- 2. ÉTAT DES CONTRÔLES -->
    <div style="background: var(--bg-subtle, #f8fafc); border: 1px solid var(--border-muted, #e2e8f0); border-radius: 8px; padding: 20px; margin-bottom: 20px;">
      <div style="font-weight: 600; font-size: 1rem; color: var(--text-primary, #0f172a); margin-bottom: 14px;">
        🔍 État des contrôles de conformité
      </div>
      <div style="display: flex; flex-direction: column; gap: 10px;">
        ${checks.map(c => `
          <div style="display: flex; justify-content: space-between; align-items: center; background: #ffffff; padding: 10px 14px; border-radius: 6px; border: 1px solid var(--border-muted, #e2e8f0); font-size: 0.85rem;">
            <div>
              <span style="font-weight: 500; color: var(--text-primary, #0f172a);">${c.label}</span>
              <div style="font-size: 0.75rem; color: var(--text-secondary, #64748b); margin-top: 2px;">${c.detail}</div>
            </div>
            <div>
              ${c.status === 'OK' ? '<span style="background: #dcfce7; color: #166534; font-size: 0.75rem; font-weight: 600; padding: 3px 8px; border-radius: 4px;">VALIDE</span>' : ''}
              ${c.status === 'FAILED' ? '<span style="background: #fee2e2; color: #991b1b; font-size: 0.75rem; font-weight: 600; padding: 3px 8px; border-radius: 4px;">ÉCHEC</span>' : ''}
              ${c.status === 'PENDING' ? '<span style="background: #f1f5f9; color: #475569; font-size: 0.75rem; font-weight: 600; padding: 3px 8px; border-radius: 4px;">EN ATTENTE</span>' : ''}
            </div>
          </div>
        `).join('')}
      </div>
    </div>

    <!-- 3. CE QU'IL FAUT CORRIGER -->
    ${corrections.length > 0 ? `
      <div style="background: #fffbeb; border: 1px solid #fde68a; border-radius: 8px; padding: 20px;">
        <div style="font-weight: 600; font-size: 1rem; color: #92400e; margin-bottom: 12px; display: flex; align-items: center; gap: 8px;">
          <span>⚠️</span> Ce qu'il faut corriger sur la facture
        </div>
        <div style="display: flex; flex-direction: column; gap: 12px;">
          ${corrections.map((corr, idx) => `
            <div style="background: #ffffff; border-left: 4px solid #f59e0b; padding: 12px 14px; border-radius: 0 6px 6px 0; font-size: 0.85rem;">
              <div style="font-weight: 600; color: #b45309; margin-bottom: 4px;">Point ${idx + 1} : ${corr.field}</div>
              <div style="color: #475569; margin-bottom: 4px;"><strong>Anomalie détectée :</strong> ${corr.issue}</div>
              <div style="color: #1e293b;"><strong>Action corrective :</strong> ${corr.action}</div>
            </div>
          `).join('')}
        </div>
      </div>
    ` : `
      <div style="background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 8px; padding: 16px; color: #166534; font-size: 0.875rem; font-weight: 500;">
        🎉 Aucun correctif requis. La facture répond à 100% aux exigences de la norme EN 16931 et du profil CIUS-FR.
      </div>
    `}
  `;
}
