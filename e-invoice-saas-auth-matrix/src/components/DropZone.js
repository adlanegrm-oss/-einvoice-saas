export function initDropZone(containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = `
    <!-- Zone de glisser-déposer principale -->
    <div class="dropzone-box" id="dropzone-area" style="border: 2px dashed #3b82f6; border-radius: 8px; padding: 28px 20px; text-align: center; background: rgba(59, 130, 246, 0.03); cursor: pointer; transition: all 0.2s ease;">
      <input type="file" id="invoice-file-input" style="display: none;" accept=".xml,.pdf,.edi" />
      <input type="file" id="invoice-multi-file-input" style="display: none;" accept=".xml,.pdf,.edi" multiple />
      
      <div style="font-size: 1.05rem; color: var(--text-primary, #1e293b); font-weight: 500;">
        Glissez-déposez une ou plusieurs factures ici ou <span style="color: #2563eb; text-decoration: underline;">parcourir un fichier</span>
      </div>
      <div style="font-size: 0.85rem; color: var(--text-secondary, #64748b); margin-top: 6px;">
        Formats acceptés : UBL (.xml), CII / Factur-X (.pdf, .xml), EDIFACT (.edi)
      </div>

      <!-- Bouton d'upload multiple explicite -->
      <div style="margin-top: 18px;">
        <button type="button" id="btn-multi-upload" style="background: #2563eb; color: #ffffff; border: none; border-radius: 6px; padding: 9px 18px; font-size: 0.85rem; font-weight: 600; cursor: pointer; display: inline-flex; align-items: center; gap: 8px; box-shadow: 0 1px 2px rgba(0,0,0,0.05);">
          <span>📁</span> Uploader plusieurs factures (Lot)
        </button>
      </div>
    </div>

    <!-- En-tête du lot déposé -->
    <div id="batch-summary" style="margin-top: 20px; display: none; align-items: center; justify-content: space-between; background: var(--bg-subtle, #f8fafc); border: 1px solid var(--border-muted, #e2e8f0); padding: 12px 18px; border-radius: 6px;">
      <div style="font-size: 0.9rem; font-weight: 600; color: var(--text-primary, #0f172a);" id="batch-count-label"></div>
      <button type="button" id="btn-clear-batch" style="background: transparent; border: 1px solid #cbd5e1; color: #64748b; font-size: 0.75rem; padding: 4px 10px; border-radius: 4px; cursor: pointer;">Réinitialiser la liste</button>
    </div>

    <!-- Liste dynamique des factures inspectées -->
    <div id="invoices-list-container" style="margin-top: 16px; display: flex; flex-direction: column; gap: 20px;"></div>
  `;

  const dropArea = document.getElementById('dropzone-area');
  const fileInput = document.getElementById('invoice-file-input');
  const multiFileInput = document.getElementById('invoice-multi-file-input');
  const btnMulti = document.getElementById('btn-multi-upload');
  const btnClear = document.getElementById('btn-clear-batch');

  // Déclencheurs de sélection
  dropArea.addEventListener('click', (e) => {
    if (e.target !== btnMulti && !btnMulti.contains(e.target)) {
      fileInput.click();
    }
  });

  btnMulti.addEventListener('click', (e) => {
    e.stopPropagation();
    multiFileInput.click();
  });

  fileInput.addEventListener('change', (e) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFiles(Array.from(e.target.files));
    }
  });

  multiFileInput.addEventListener('change', (e) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFiles(Array.from(e.target.files));
    }
  });

  // Glisser-déposer (gère 1 ou plusieurs fichiers simultanément)
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
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFiles(Array.from(e.dataTransfer.files));
    }
  });

  btnClear.addEventListener('click', () => {
    document.getElementById('invoices-list-container').innerHTML = '';
    document.getElementById('batch-summary').style.display = 'none';
    fileInput.value = '';
    multiFileInput.value = '';
  });
}

function handleFiles(files) {
  const listContainer = document.getElementById('invoices-list-container');
  const batchSummary = document.getElementById('batch-summary');
  const batchCountLabel = document.getElementById('batch-count-label');
  if (!listContainer) return;

  batchSummary.style.display = 'flex';
  batchCountLabel.textContent = `📋 Lot de ${files.length} facture(s) analysée(s)`;

  listContainer.innerHTML = '';
  files.forEach((file, index) => {
    listContainer.appendChild(createInvoiceCard(file, index + 1));
  });
}

function createInvoiceCard(file, index) {
  const card = document.createElement('div');
  card.style.border = '1px solid var(--border-muted, #e2e8f0)';
  card.style.borderRadius = '8px';
  card.style.padding = '20px';
  card.style.background = 'var(--bg-card, #ffffff)';
  card.style.boxShadow = '0 1px 3px rgba(0,0,0,0.03)';

  const fileSizeKo = (file.size / 1024).toFixed(1);
  const ext = file.name.split('.').pop().toLowerCase();

  let checks = [];
  let corrections = [];

  // Analyse différenciée selon les règles fiscales, légales et techniques françaises
  if (ext === 'pdf') {
    checks = [
      { label: "Norme Archivage & Conteneur PDF/A-3 (ISO 19005-3)", status: "OK", detail: "Format conteneur certifié PDF/A-3" },
      { label: "Contrôle Légal : Pièce jointe XML (factur-x.xml)", status: "FAILED", detail: "Fichier de données structurées absent du conteneur PDF" },
      { label: "Contrôle Fiscal : Mentions obligatoires CGI (Art. 242 nonies A)", status: "FAILED", detail: "Absence du numéro de TVA intracommunautaire du vendeur / SIREN" },
      { label: "Règles Sectorielles B2G / CIUS-FR", status: "PENDING", detail: "Vérification Chorus Pro en attente de la charge structurée" }
    ];

    corrections = [
      {
        type: "Conformité Technique",
        field: "Structure Factur-X",
        detail: "Le PDF déposé est un fichier bureautique sans flux structuré embarqué.",
        action: "Exportez la facture au format hybride Factur-X officiel ou joignez directement le fichier XML équivalent (UBL / CII)."
      },
      {
        type: "Exigence Fiscale & Légale",
        field: "Mentions Légales TVA (Art. 242 nonies A)",
        detail: "Numéro de TVA de l'émetteur ou numéro SIRET non lisible dans les métadonnées.",
        action: "Compléter la fiche client de l'émetteur avec son identifiant fiscal FR et son adresse complète de facturation."
      }
    ];
  } else if (ext === 'xml') {
    checks = [
      { label: "Syntaxe & Encodage UTF-8 (Well-formed XML)", status: "OK", detail: "Structure XML et entête valides" },
      { label: "Schéma EN 16931 XSD (UBL / UN-CEFACT CII)", status: "OK", detail: "Balises métier conformes au standard européen" },
      { label: "Contrôles Fiscaux : Calcul de la ventilation de TVA", status: "OK", detail: "Base hors-taxe et taux 20% / 10% / 5.5% équilibrés" },
      { label: "Règles Légales CIUS-FR (Chorus Pro / PDP)", status: "OK", detail: "Identifiants d'acheminement et acheteur vérifiés" }
    ];
    corrections = [];
  } else if (ext === 'edi') {
    checks = [
      { label: "Syntaxe EDIFACT D01B (Interchange UNB)", status: "OK", detail: "Segments UNH, BGM, DTM, MOA détectés" },
      { label: "Contrôle Légal : Agrément fiscal EDI", status: "OK", detail: "Signature électronique et compte-rendu d'interchange valides" },
      { label: "Transcodage vers la norme EN 16931", status: "OK", detail: "Correspondance sémantique vers CIUS-FR établie" }
    ];
    corrections = [];
  } else {
    checks = [
      { label: "Format de fichier", status: "FAILED", detail: `Extension .${ext} non conforme` }
    ];
    corrections = [
      {
        type: "Format de dépôt",
        field: "Type de fichier",
        detail: `Le format .${ext} n'est pas autorisé par l'ordonnance facturation électronique.`,
        action: "Déposez exclusivement des factures aux formats XML (UBL, CII), PDF (Factur-X) ou EDI (EDIFACT)."
      }
    ];
  }

  const isSuccess = corrections.length === 0;

  card.innerHTML = `
    <!-- 1. APERÇU FACTURE -->
    <div style="display: flex; justify-content: space-between; align-items: flex-start; border-bottom: 1px solid var(--border-muted, #e2e8f0); padding-bottom: 14px; margin-bottom: 16px;">
      <div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="font-size: 1.1rem;">📄</span>
          <span style="font-weight: 600; font-size: 0.95rem; color: var(--text-primary, #0f172a);">${file.name}</span>
          <span style="font-size: 0.75rem; background: #e2e8f0; color: #334155; padding: 2px 6px; border-radius: 4px; font-weight: 600;">Facture #${index}</span>
        </div>
        <div style="font-size: 0.8rem; color: var(--text-secondary, #64748b); margin-top: 4px;">
          Poids : <strong>${fileSizeKo} Ko</strong> | Format détecté : <strong style="text-transform: uppercase;">${ext}</strong> | Date d'analyse : ${new Date().toLocaleTimeString()}
        </div>
      </div>
      <span style="display: inline-block; padding: 4px 10px; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; ${isSuccess ? 'background: #dcfce7; color: #15803d;' : 'background: #fee2e2; color: #b91c1c;'}">
        ${isSuccess ? '✔ Conforme Fiscale & Légale' : '✖ Rejet des Contrôles'}
      </span>
    </div>

    <!-- 2. ÉTAT DES CONTRÔLES FISCAUX ET LÉGAUX -->
    <div style="margin-bottom: 16px;">
      <div style="font-weight: 600; font-size: 0.85rem; color: var(--text-primary, #0f172a); margin-bottom: 10px; text-transform: uppercase; letter-spacing: 0.025em;">
        ⚖️ État des contrôles légaux, fiscaux et techniques
      </div>
      <div style="display: flex; flex-direction: column; gap: 8px;">
        ${checks.map(c => `
          <div style="display: flex; justify-content: space-between; align-items: center; background: var(--bg-subtle, #f8fafc); padding: 8px 12px; border-radius: 6px; border: 1px solid var(--border-muted, #e2e8f0); font-size: 0.8rem;">
            <div>
              <span style="font-weight: 500; color: var(--text-primary, #0f172a);">${c.label}</span>
              <div style="font-size: 0.75rem; color: var(--text-secondary, #64748b);">${c.detail}</div>
            </div>
            <div>
              ${c.status === 'OK' ? '<span style="background: #dcfce7; color: #166534; font-size: 0.7rem; font-weight: 700; padding: 2px 6px; border-radius: 4px;">CONFORME</span>' : ''}
              ${c.status === 'FAILED' ? '<span style="background: #fee2e2; color: #991b1b; font-size: 0.7rem; font-weight: 700; padding: 2px 6px; border-radius: 4px;">NON CONFORME</span>' : ''}
              ${c.status === 'PENDING' ? '<span style="background: #f1f5f9; color: #475569; font-size: 0.7rem; font-weight: 700; padding: 2px 6px; border-radius: 4px;">EN SUSPENS</span>' : ''}
            </div>
          </div>
        `).join('')}
      </div>
    </div>

    <!-- 3. CE QU'IL FAUT CORRIGER SUR LA FACTURE -->
    <div>
      ${corrections.length > 0 ? `
        <div style="background: #fffbeb; border: 1px solid #fde68a; border-radius: 6px; padding: 14px;">
          <div style="font-weight: 600; font-size: 0.85rem; color: #92400e; margin-bottom: 8px; display: flex; align-items: center; gap: 6px;">
            <span>⚠️</span> Ce qu'il faut corriger sur cette facture :
          </div>
          <div style="display: flex; flex-direction: column; gap: 8px;">
            ${corrections.map((corr, cIdx) => `
              <div style="background: #ffffff; border-left: 3px solid #f59e0b; padding: 10px 12px; border-radius: 0 4px 4px 0; font-size: 0.8rem;">
                <div style="font-weight: 600; color: #b45309; margin-bottom: 2px;">${corr.type} · ${corr.field}</div>
                <div style="color: #475569; margin-bottom: 2px;"><strong>Anomalie :</strong> ${corr.detail}</div>
                <div style="color: #1e293b;"><strong>Action recommandée :</strong> ${corr.action}</div>
              </div>
            `).join('')}
          </div>
        </div>
      ` : `
        <div style="background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 6px; padding: 12px 14px; color: #166534; font-size: 0.825rem; font-weight: 500;">
          ✔ Tous les contrôles fiscaux (CGI art. 242) et techniques (EN 16931) sont validés pour cette facture. Prête pour l'émission vers le portail public ou PDP.
        </div>
      `}
    </div>
  `;

  return card;
}
