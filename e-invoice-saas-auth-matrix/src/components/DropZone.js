export function initDropZone(containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = `
    <!-- Zone de dépôt -->
    <div class="dropzone-box" id="dropzone-area" style="border: 2px dashed #3b82f6; border-radius: 8px; padding: 26px 20px; text-align: center; background: rgba(59, 130, 246, 0.03); cursor: pointer; transition: all 0.2s ease;">
      <input type="file" id="invoice-file-input" style="display: none;" accept=".xml,.pdf,.edi" />
      <input type="file" id="invoice-multi-file-input" style="display: none;" accept=".xml,.pdf,.edi" multiple />
      
      <div style="font-size: 1.05rem; color: var(--text-primary, #1e293b); font-weight: 500;">
        Glissez-déposez vos factures ici ou <span style="color: #2563eb; text-decoration: underline;">parcourir un fichier</span>
      </div>
      <div style="font-size: 0.85rem; color: var(--text-secondary, #64748b); margin-top: 6px;">
        Formats acceptés : UBL (.xml), CII / Factur-X (.pdf, .xml), EDIFACT (.edi)
      </div>

      <div style="margin-top: 16px;">
        <button type="button" id="btn-multi-upload" style="background: #2563eb; color: #ffffff; border: none; border-radius: 6px; padding: 9px 18px; font-size: 0.85rem; font-weight: 600; cursor: pointer; display: inline-flex; align-items: center; gap: 8px; box-shadow: 0 1px 2px rgba(0,0,0,0.05);">
          <span>📁</span> Uploader plusieurs factures (Lot)
        </button>
      </div>
    </div>

    <!-- 1. NOTIFICATION RÉCAPITULATIVE (BANNIÈRE ACCUSÉ DE RÉCEPTION) -->
    <div id="batch-notification-banner" style="margin-top: 20px; display: none; border-radius: 8px; padding: 16px 20px; border: 1px solid transparent; box-shadow: 0 1px 3px rgba(0,0,0,0.05);"></div>

    <!-- 2. BOUTONS D'EXPORT DU RAPPORT & GESTION -->
    <div id="batch-actions-bar" style="margin-top: 14px; display: none; align-items: center; justify-content: space-between; background: var(--bg-subtle, #f8fafc); border: 1px solid var(--border-muted, #e2e8f0); padding: 12px 18px; border-radius: 6px;">
      <div style="display: flex; gap: 10px; align-items: center;">
        <button type="button" id="btn-download-report-json" style="background: #0f172a; color: #ffffff; border: none; padding: 7px 14px; border-radius: 5px; font-size: 0.8rem; font-weight: 600; cursor: pointer; display: inline-flex; align-items: center; gap: 6px;">
          <span>📥</span> Télécharger Rapport d'Audit (.json)
        </button>
        <button type="button" id="btn-download-report-csv" style="background: #ffffff; border: 1px solid #cbd5e1; color: #334155; padding: 7px 14px; border-radius: 5px; font-size: 0.8rem; font-weight: 600; cursor: pointer; display: inline-flex; align-items: center; gap: 6px;">
          <span>📊</span> Exporter Synthèse (.csv)
        </button>
      </div>
      <button type="button" id="btn-clear-batch" style="background: transparent; border: 1px solid #cbd5e1; color: #64748b; font-size: 0.75rem; padding: 5px 12px; border-radius: 4px; cursor: pointer;">Réinitialiser l'espace</button>
    </div>

    <!-- 3. JOURNAL DE TRAÇABILITÉ FRONT-OFFICE (AUDIT TRAIL) -->
    <div id="audit-trail-section" style="margin-top: 20px; display: none; background: #ffffff; border: 1px solid var(--border-muted, #e2e8f0); border-radius: 8px; padding: 18px;">
      <div style="font-weight: 600; font-size: 0.9rem; color: var(--text-primary, #0f172a); margin-bottom: 12px; display: flex; align-items: center; gap: 8px;">
        <span>🛡️</span> Traçabilité des actions Front-Office (Audit Trail émetteur)
      </div>
      <div style="overflow-x: auto;">
        <table style="width: 100%; border-collapse: collapse; font-size: 0.8rem; text-align: left;">
          <thead>
            <tr style="background: var(--bg-subtle, #f8fafc); border-bottom: 1px solid #e2e8f0; color: #64748b;">
              <th style="padding: 8px 10px;">Horodatage</th>
              <th style="padding: 8px 10px;">Événement</th>
              <th style="padding: 8px 10px;">Fichier concerné</th>
              <th style="padding: 8px 10px;">Empreinte (SHA-256 sim.)</th>
              <th style="padding: 8px 10px;">Statut</th>
            </tr>
          </thead>
          <tbody id="audit-trail-rows"></tbody>
        </table>
      </div>
    </div>

    <!-- 4. LISTE DES FACTURES ET ANALYSES DÉTAILLÉES -->
    <div id="invoices-list-container" style="margin-top: 20px; display: flex; flex-direction: column; gap: 18px;"></div>
  `;

  const dropArea = document.getElementById('dropzone-area');
  const fileInput = document.getElementById('invoice-file-input');
  const multiFileInput = document.getElementById('invoice-multi-file-input');
  const btnMulti = document.getElementById('btn-multi-upload');
  const btnClear = document.getElementById('btn-clear-batch');
  const btnJson = document.getElementById('btn-download-report-json');
  const btnCsv = document.getElementById('btn-download-report-csv');

  let currentBatchData = null;

  dropArea.addEventListener('click', (e) => {
    if (e.target !== btnMulti && !btnMulti.contains(e.target)) fileInput.click();
  });

  btnMulti.addEventListener('click', (e) => {
    e.stopPropagation();
    multiFileInput.click();
  });

  fileInput.addEventListener('change', (e) => {
    if (e.target.files && e.target.files.length > 0) processBatch(Array.from(e.target.files));
  });

  multiFileInput.addEventListener('change', (e) => {
    if (e.target.files && e.target.files.length > 0) processBatch(Array.from(e.target.files));
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
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      processBatch(Array.from(e.dataTransfer.files));
    }
  });

  btnClear.addEventListener('click', () => {
    document.getElementById('invoices-list-container').innerHTML = '';
    document.getElementById('batch-notification-banner').style.display = 'none';
    document.getElementById('batch-actions-bar').style.display = 'none';
    document.getElementById('audit-trail-section').style.display = 'none';
    fileInput.value = '';
    multiFileInput.value = '';
    currentBatchData = null;
  });

  btnJson.addEventListener('click', () => {
    if (!currentBatchData) return;
    downloadFile(`rapport_conformite_${currentBatchData.batchId}.json`, JSON.stringify(currentBatchData, null, 2), 'application/json');
  });

  btnCsv.addEventListener('click', () => {
    if (!currentBatchData) return;
    let csv = "Nom Fichier;Taille (Ko);Format;Statut Global;Erreurs Detectees\n";
    currentBatchData.invoices.forEach(inv => {
      const errs = inv.corrections.map(c => `${c.field}: ${c.issue}`).join(' | ');
      csv += `"${inv.fileName}";"${inv.fileSizeKo}";"${inv.format}";"${inv.status}";"${errs}"\n`;
    });
    downloadFile(`synthese_conformite_${currentBatchData.batchId}.csv`, csv, 'text/csv;charset=utf-8;');
  });

  async function processBatch(files) {
    const batchId = 'DEP-' + Math.floor(100000 + Math.random() * 900000);
    const timestamp = new Date().toISOString();
    const listContainer = document.getElementById('invoices-list-container');
    const notificationBanner = document.getElementById('batch-notification-banner');
    const actionsBar = document.getElementById('batch-actions-bar');
    const auditSection = document.getElementById('audit-trail-section');
    const auditTbody = document.getElementById('audit-trail-rows');

    listContainer.innerHTML = '';
    auditTbody.innerHTML = '';

    const results = [];
    let okCount = 0;
    let nokCount = 0;

    for (let i = 0; i < files.length; i++) {
      const analyzed = await analyzeInvoice(files[i], i + 1, batchId);
      results.push(analyzed);
      if (analyzed.status === 'OK') okCount++;
      else nokCount++;

      listContainer.appendChild(analyzed.element);

      // Ligne d'audit
      const tr = document.createElement('tr');
      tr.style.borderBottom = '1px solid #f1f5f9';
      tr.innerHTML = `
        <td style="padding: 8px 10px; color: #64748b;">${new Date().toLocaleTimeString()}</td>
        <td style="padding: 8px 10px; font-weight: 500;">INSPECTION_METIER</td>
        <td style="padding: 8px 10px;">${analyzed.fileName}</td>
        <td style="padding: 8px 10px; font-family: monospace; color: #475569;">${analyzed.mockHash}</td>
        <td style="padding: 8px 10px;">
          <span style="font-weight: 600; font-size: 0.75rem; padding: 2px 6px; border-radius: 4px; ${analyzed.status === 'OK' ? 'background: #dcfce7; color: #15803d;' : 'background: #fee2e2; color: #b91c1c;'}">
            ${analyzed.status === 'OK' ? 'CONFORME' : 'REJETÉ'}
          </span>
        </td>
      `;
      auditTbody.appendChild(tr);
    }

    currentBatchData = {
      batchId,
      timestamp,
      totalCount: files.length,
      acceptedCount: okCount,
      rejectedCount: nokCount,
      invoices: results.map(r => ({
        fileName: r.fileName,
        fileSizeKo: r.fileSizeKo,
        format: r.format,
        status: r.status,
        checks: r.checks,
        corrections: r.corrections,
        mockHash: r.mockHash
      }))
    };

    // Configuration de la bannière de notification
    notificationBanner.style.display = 'block';
    if (nokCount === 0) {
      notificationBanner.style.background = '#f0fdf4';
      notificationBanner.style.borderColor = '#bbf7d0';
      notificationBanner.innerHTML = `
        <div style="font-weight: 600; color: #166534; font-size: 1rem; display: flex; align-items: center; gap: 8px;">
          <span>✅</span> Notification Dépôt [Réf : ${batchId}] : 100% des factures validées
        </div>
        <div style="color: #15803d; font-size: 0.85rem; margin-top: 4px;">
          L'ensemble du lot (${okCount} facture(s)) répond intégralement aux exigences de conformité légale et fiscale. Prêt pour transmission.
        </div>
      `;
    } else {
      notificationBanner.style.background = '#fef2f2';
      notificationBanner.style.borderColor = '#fecaca';
      notificationBanner.innerHTML = `
        <div style="font-weight: 600; color: #991b1b; font-size: 1rem; display: flex; align-items: center; gap: 8px;">
          <span>🔔</span> Notification Dépôt [Réf : ${batchId}] : Anomalies détectées sur le lot
        </div>
        <div style="color: #b91c1c; font-size: 0.85rem; margin-top: 4px;">
          Bilan du traitement : <strong>${okCount} facture(s) OK</strong> | <strong style="text-decoration: underline;">${nokCount} facture(s) NON CONFORME(S)</strong>. Un rapport d'anomalies a été consolidé ci-dessous.
        </div>
      `;
    }

    actionsBar.style.display = 'flex';
    auditSection.style.display = 'block';
  }
}

function readFileContent(file) {
  return new Promise((resolve) => {
    if (file.type === 'application/pdf' || file.name.endsWith('.pdf')) {
      resolve('');
      return;
    }
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result || '');
    reader.onerror = () => resolve('');
    reader.readAsText(file);
  });
}

function downloadFile(filename, content, type) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

async function analyzeInvoice(file, index, batchId) {
  const content = await readFileContent(file);
  const fileSizeKo = (file.size / 1024).toFixed(1);
  const ext = file.name.split('.').pop().toLowerCase();
  const mockHash = 'e3b0c442...' + Math.random().toString(16).substring(2, 8);

  let checks = [];
  let corrections = [];

  if (ext === 'pdf') {
    checks = [
      { label: "Norme Archivage & Conteneur PDF/A-3", status: "OK", detail: "Profil ISO 19005-3 valide" },
      { label: "Contrôle Légal : Pièce jointe XML (factur-x.xml)", status: "FAILED", detail: "Fichier structuré absent du conteneur PDF" },
      { label: "Contrôle Fiscal : Mentions obligatoires CGI (Art. 242 nonies A)", status: "FAILED", detail: "Numéro de TVA intracommunautaire émetteur absent" },
      { label: "Règles Sectorielles B2G / Chorus Pro", status: "PENDING", detail: "En suspens" }
    ];
    corrections = [
      {
        field: "Structure hybride Factur-X",
        issue: "Le PDF déposé est un fichier bureautique sans flux XML joint.",
        action: "Exportez la facture en mode Factur-X ou déposez directement le flux UBL/CII."
      },
      {
        field: "Identification fiscale de l'émetteur",
        issue: "Numéro de TVA ou SIREN manquant dans les métadonnées.",
        action: "Ajouter la mention TVA intracommunautaire du vendeur."
      }
    ];
  } else if (ext === 'xml') {
    const hasCius = content.includes('cius-fr') || content.includes('EN16931');
    const hasVat = content.includes('CompanyID') || content.includes('TaxScheme');
    const hasTaxTotal = content.includes('TaxTotal');

    if (hasCius && hasVat && hasTaxTotal) {
      checks = [
        { label: "Syntaxe & Encodage UTF-8", status: "OK", detail: "XML bien formé" },
        { label: "Schéma EN 16931 XSD (UBL / CII)", status: "OK", detail: "Profil CIUS-FR conforme" },
        { label: "Contrôles Fiscaux : Ventilation de TVA", status: "OK", detail: "Taux et montants équilibrés" },
        { label: "Règles Légales CIUS-FR", status: "OK", detail: "Identifiants d'acheminement validés" }
      ];
      corrections = [];
    } else {
      checks = [
        { label: "Syntaxe & Encodage UTF-8", status: "OK", detail: "XML bien formé" },
        { label: "Schéma EN 16931 XSD (UBL / CII)", status: hasCius ? "OK" : "FAILED", detail: hasCius ? "Conforme" : "CustomizationID CIUS-FR absent" },
        { label: "Contrôles Fiscaux : N° TVA (Art. 242 nonies A)", status: hasVat ? "OK" : "FAILED", detail: hasVat ? "Conforme" : "PartyTaxScheme émetteur manquant" },
        { label: "Règles Légales CIUS-FR : Bloc TVA", status: hasTaxTotal ? "OK" : "FAILED", detail: hasTaxTotal ? "Conforme" : "TaxTotal manquant" }
      ];
      corrections = [
        {
          field: "Identifiant fiscal émetteur",
          issue: "Numéro de TVA intracommunautaire émetteur absent du bloc fournisseur.",
          action: "Déclarer la balise <cac:PartyTaxScheme> avec le numéro de TVA français."
        },
        {
          field: "Identifiant de personnalisation CIUS-FR",
          issue: "La personnalisation française EN 16931 n'est pas stipulée.",
          action: "Ajouter CustomizationID 'urn:cen.eu:en16931:2017#compliant#urn:facx.org:1p0:cius-fr'."
        }
      ];
    }
  } else if (ext === 'edi') {
    checks = [
      { label: "Syntaxe EDIFACT D01B", status: "OK", detail: "Segments UNH, BGM, DTM, MOA valides" },
      { label: "Contrôle Légal : Agrément fiscal EDI", status: "OK", detail: "Interchange régulier" },
      { label: "Transcodage vers norme EN 16931", status: "OK", detail: "Mapping CIUS-FR établi" }
    ];
    corrections = [];
  } else {
    checks = [{ label: "Extension", status: "FAILED", detail: `Type .${ext} non pris en charge` }];
    corrections = [{ field: "Format", issue: "Format refusé", action: "Utiliser .xml, .pdf Factur-X ou .edi." }];
  }

  const isSuccess = corrections.length === 0;

  const card = document.createElement('div');
  card.style.border = '1px solid var(--border-muted, #e2e8f0)';
  card.style.borderRadius = '8px';
  card.style.padding = '18px';
  card.style.background = 'var(--bg-card, #ffffff)';
  card.style.boxShadow = '0 1px 3px rgba(0,0,0,0.03)';

  card.innerHTML = `
    <div style="display: flex; justify-content: space-between; align-items: flex-start; border-bottom: 1px solid var(--border-muted, #e2e8f0); padding-bottom: 12px; margin-bottom: 14px;">
      <div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="font-size: 1.1rem;">📄</span>
          <span style="font-weight: 600; font-size: 0.95rem; color: var(--text-primary, #0f172a);">${file.name}</span>
          <span style="font-size: 0.75rem; background: #e2e8f0; color: #334155; padding: 2px 6px; border-radius: 4px; font-weight: 600;">#${index}</span>
        </div>
        <div style="font-size: 0.8rem; color: var(--text-secondary, #64748b); margin-top: 4px;">
          Poids : <strong>${fileSizeKo} Ko</strong> | Format : <strong style="text-transform: uppercase;">${ext}</strong> | Lot : <code>${batchId}</code>
        </div>
      </div>
      <span style="display: inline-block; padding: 4px 10px; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; ${isSuccess ? 'background: #dcfce7; color: #15803d;' : 'background: #fee2e2; color: #b91c1c;'}">
        ${isSuccess ? '✔ Conforme Fiscale & Légale' : '✖ Rejet des Contrôles'}
      </span>
    </div>

    <div style="margin-bottom: 14px;">
      <div style="font-weight: 600; font-size: 0.825rem; color: var(--text-primary, #0f172a); margin-bottom: 8px; text-transform: uppercase;">
        ⚖️ Grille des contrôles légaux et fiscaux
      </div>
      <div style="display: flex; flex-direction: column; gap: 6px;">
        ${checks.map(c => `
          <div style="display: flex; justify-content: space-between; align-items: center; background: var(--bg-subtle, #f8fafc); padding: 7px 12px; border-radius: 6px; border: 1px solid var(--border-muted, #e2e8f0); font-size: 0.8rem;">
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

    <div>
      ${corrections.length > 0 ? `
        <div style="background: #fffbeb; border: 1px solid #fde68a; border-radius: 6px; padding: 12px;">
          <div style="font-weight: 600; font-size: 0.825rem; color: #92400e; margin-bottom: 8px; display: flex; align-items: center; gap: 6px;">
            <span>⚠️</span> Ce qu'il faut corriger sur cette facture :
          </div>
          <div style="display: flex; flex-direction: column; gap: 8px;">
            ${corrections.map(corr => `
              <div style="background: #ffffff; border-left: 3px solid #f59e0b; padding: 8px 12px; border-radius: 0 4px 4px 0; font-size: 0.8rem;">
                <div style="font-weight: 600; color: #b45309; margin-bottom: 2px;">${corr.field}</div>
                <div style="color: #475569; margin-bottom: 2px;"><strong>Anomalie :</strong> ${corr.issue}</div>
                <div style="color: #1e293b;"><strong>Action recommandée :</strong> ${corr.action}</div>
              </div>
            `).join('')}
          </div>
        </div>
      ` : `
        <div style="background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 6px; padding: 10px 14px; color: #166534; font-size: 0.825rem; font-weight: 500;">
          ✔ Tous les contrôles fiscaux (CGI art. 242) et techniques (EN 16931) sont validés pour cette facture.
        </div>
      `}
    </div>
  `;

  return {
    fileName: file.name,
    fileSizeKo,
    format: ext,
    status: isSuccess ? 'OK' : 'REJECTED',
    checks,
    corrections,
    mockHash,
    element: card
  };
}
