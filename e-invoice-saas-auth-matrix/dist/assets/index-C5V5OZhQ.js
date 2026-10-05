(function(){const o=document.createElement("link").relList;if(o&&o.supports&&o.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))t(n);new MutationObserver(n=>{for(const a of n)if(a.type==="childList")for(const l of a.addedNodes)l.tagName==="LINK"&&l.rel==="modulepreload"&&t(l)}).observe(document,{childList:!0,subtree:!0});function r(n){const a={};return n.integrity&&(a.integrity=n.integrity),n.referrerPolicy&&(a.referrerPolicy=n.referrerPolicy),n.crossOrigin==="use-credentials"?a.credentials="include":n.crossOrigin==="anonymous"?a.credentials="omit":a.credentials="same-origin",a}function t(n){if(n.ep)return;n.ep=!0;const a=r(n);fetch(n.href,a)}})();const v={CLIENT:{label:"Front Client",path:"/app",color:"#2563eb",defaultTheme:"light",description:"Utilisation quotidienne métier (dépôt, suivi, téléchargement)",routes:[{id:"client-upload",label:"Déposer des factures"},{id:"client-invoices",label:"Liste des factures"},{id:"client-validation",label:"Validation seule"}]},PARTNER:{label:"Front Partenaire",path:"/partner",color:"#7c3aed",defaultTheme:"light",description:"Intégrateurs, cabinets, ERP, PDP partenaires (multi-tenants)",routes:[{id:"partner-dash",label:"Dashboard Consolidé"},{id:"partner-clients",label:"Gestion des Clients"},{id:"partner-api",label:"Console API & Sandbox"}]},ADMIN:{label:"Console Admin",path:"/admin",color:"#ea580c",defaultTheme:"dark",description:"Administration globale multi-tenant, contrôle et gouvernance",routes:[{id:"admin-users",label:"Utilisateurs & Rôles"},{id:"admin-tenants",label:"Gestion des Tenants"},{id:"admin-outbox",label:"Supervision Outbox & Audit"}]},OPS:{label:"Console Exploitation",path:"/ops",color:"#059669",defaultTheme:"dark",description:"Supervision production, incidents, santé workers, DLQ",routes:[{id:"ops-metrics",label:"Santé Système & Métriques"},{id:"ops-dlq",label:"Outbox & File DLQ"}]},MCO:{label:"Console MCO",path:"/mco",color:"#475569",defaultTheme:"dark",description:"Maintenance, support N2/N3, runbooks et plans PRA",routes:[{id:"mco-runbooks",label:"Runbooks d'exploitation"},{id:"mco-pra",label:"Procédures PRA & Backup"}]},RND:{label:"Console R&D",path:"/rnd",color:"#4f46e5",defaultTheme:"dark",description:"Sandbox de conformité, règles BR-xx et qualification CIUS-FR",routes:[{id:"rnd-schematron",label:"Banc de Test Schematron"},{id:"rnd-golden",label:"Comparateur Golden Files"}]},DEPLOY:{label:"Console Déploiement",path:"/deploy",color:"#0d9488",defaultTheme:"dark",description:"Releases, statuts environnements, CI/CD et rollbacks",routes:[{id:"deploy-envs",label:"Statut des Environnements"},{id:"deploy-pipeline",label:"Pipelines CI/CD & Scans"}]}};function N(i,o,r){const t=v[i],n=t.routes.map(l=>`<li><a class="${l.id===o?"active":""}" data-nav="${l.id}">${l.label}</a></li>`).join("");return`
    <div class="layout-shell">
      <aside class="sidebar">
        <div>
          <div class="brand">e-invoice-saas</div>
          <div style="padding:12px;">
            <label style="font-size:0.75rem; color:var(--text-secondary); font-weight:600; text-transform:uppercase;">Changer d'espace :</label>
            <select id="role-selector" style="width:100%; margin-top:6px; padding:8px; border-radius:6px; border:1px solid var(--border-muted); background:var(--bg-subtle); color:var(--text-primary); font-size:0.825rem;">
              ${Object.keys(v).map(l=>`<option value="${l}" ${l===i?"selected":""}>${v[l].label} (${v[l].path})</option>`).join("")}
            </select>
          </div>
          <ul class="nav-menu">
            ${n}
          </ul>
        </div>
        <div style="padding:16px; border-top:1px solid var(--border-muted); font-size:0.8rem;">
          <div style="color:var(--text-secondary);">Espace Actif</div>
          <div style="font-weight:600; color:var(--accent-primary);">${t.label}</div>
          <div style="font-size:0.75rem; color:var(--text-disabled); margin-top:4px;">Accès : ${t.path}</div>
        </div>
      </aside>
      <main class="main-content">
        <header class="topbar">
          <div>
            <span style="font-size:0.875rem; color:var(--text-secondary);">${t.description}</span>
          </div>
          <div style="display:flex; align-items:center; gap:12px;">
            <span class="badge badge-valid">Système Opérationnel</span>
            <button id="theme-toggle" class="btn btn-subtle" style="font-size:0.75rem;">Bascule Clair/Sombre</button>
          </div>
        </header>
        <div class="content-body">
          ${r}
        </div>
      </main>
    </div>
  `}const R=[{id:"usr_001",name:"Jean Dupont",email:"jean.dupont@client.fr",role:"CLIENT",tenant:"dupont-fr",status:"ACTIF"},{id:"usr_002",name:"Claire Valette",email:"c.valette@nexus-partner.io",role:"PARTNER",tenant:"nexus-audit",status:"ACTIF"},{id:"usr_003",name:"Marc Lefebvre",email:"m.lefebvre@einvoice-saas.com",role:"ADMIN",tenant:"master",status:"ACTIF"},{id:"usr_004",name:"Alexandre Roux",email:"ops@einvoice-saas.com",role:"OPS",tenant:"master",status:"ACTIF"},{id:"usr_005",name:"Équipe MCO",email:"mco-support@einvoice-saas.com",role:"MCO",tenant:"master",status:"ACTIF"},{id:"usr_006",name:"Ingénierie R&D",email:"rnd-engine@einvoice-saas.com",role:"RND",tenant:"master",status:"ACTIF"},{id:"usr_007",name:"Release Lead",email:"deploy-bot@einvoice-saas.com",role:"DEPLOY",tenant:"master",status:"ACTIF"}];function w(){return`
    <div class="card">
      <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:16px;">
        <div>
          <h2>Gestion des Profils & Habilitations</h2>
          <p style="font-size:0.85rem; color:var(--text-secondary); margin-top:4px;">Configuration des accès RBAC multi-tenants</p>
        </div>
        <button class="btn btn-primary">+ Créer un Utilisateur</button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>Utilisateur</th>
            <th>Email</th>
            <th>Rôle Assigné</th>
            <th>Tenant Rattaché</th>
            <th>Statut</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          ${R.map(o=>`
    <tr>
      <td><b>${o.name}</b></td>
      <td>${o.email}</td>
      <td><span class="badge badge-valid">${o.role}</span></td>
      <td><code>${o.tenant}</code></td>
      <td><span class="badge badge-valid">${o.status}</span></td>
      <td>
        <button class="btn btn-subtle" style="padding:4px 8px;" onclick="window.triggerRevoke('${o.id}')">Révoquer</button>
      </td>
    </tr>
  `).join("")}
        </tbody>
      </table>
    </div>
  `}function F(i,o,r,t){const n=i==="client-upload",a=t&&t.path?t.path.replace("/",""):"client";return`
    <div class="module-card">
      <div class="module-header">
        <h1 class="module-title">${o||i}</h1>
        <p class="module-subtitle">Interface opérationnelle pour le profil <strong>${r}</strong>.</p>
      </div>

      <div class="code-banner">
        <div><code>CONTEXT_ROLE: ${r}</code></div>
        <div><code>ENDPOINT_BINDING: /api/v1/${a}/${i}</code></div>
        <div><code>ENFORCE_SCHEMA: EN_16931_CIUS_FR</code></div>
      </div>

      ${n?'<div id="dropzone-root" style="margin-top: 24px;"></div>':""}
    </div>
  `}function $(i){const o=document.getElementById(i);if(!o)return;o.innerHTML=`
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
  `;const r=document.getElementById("dropzone-area"),t=document.getElementById("invoice-file-input"),n=document.getElementById("invoice-multi-file-input"),a=document.getElementById("btn-multi-upload"),l=document.getElementById("btn-clear-batch"),f=document.getElementById("btn-download-report-json"),u=document.getElementById("btn-download-report-csv");let c=null;r.addEventListener("click",e=>{e.target!==a&&!a.contains(e.target)&&t.click()}),a.addEventListener("click",e=>{e.stopPropagation(),n.click()}),t.addEventListener("change",e=>{e.target.files&&e.target.files.length>0&&m(Array.from(e.target.files))}),n.addEventListener("change",e=>{e.target.files&&e.target.files.length>0&&m(Array.from(e.target.files))}),["dragenter","dragover"].forEach(e=>{r.addEventListener(e,d=>{d.preventDefault(),r.style.borderColor="#1d4ed8",r.style.background="rgba(59, 130, 246, 0.08)"})}),["dragleave","drop"].forEach(e=>{r.addEventListener(e,d=>{d.preventDefault(),r.style.borderColor="#3b82f6",r.style.background="rgba(59, 130, 246, 0.03)"})}),r.addEventListener("drop",e=>{e.dataTransfer.files&&e.dataTransfer.files.length>0&&m(Array.from(e.dataTransfer.files))}),l.addEventListener("click",()=>{document.getElementById("invoices-list-container").innerHTML="",document.getElementById("batch-notification-banner").style.display="none",document.getElementById("batch-actions-bar").style.display="none",document.getElementById("audit-trail-section").style.display="none",t.value="",n.value="",c=null}),f.addEventListener("click",()=>{c&&L(`rapport_conformite_${c.batchId}.json`,JSON.stringify(c,null,2),"application/json")}),u.addEventListener("click",()=>{if(!c)return;let e=`Nom Fichier;Taille (Ko);Format;Statut Global;Erreurs Detectees
`;c.invoices.forEach(d=>{const b=d.corrections.map(y=>`${y.field}: ${y.issue}`).join(" | ");e+=`"${d.fileName}";"${d.fileSizeKo}";"${d.format}";"${d.status}";"${b}"
`}),L(`synthese_conformite_${c.batchId}.csv`,e,"text/csv;charset=utf-8;")});async function m(e){const d="DEP-"+Math.floor(1e5+Math.random()*9e5),b=new Date().toISOString(),y=document.getElementById("invoices-list-container"),g=document.getElementById("batch-notification-banner"),S=document.getElementById("batch-actions-bar"),O=document.getElementById("audit-trail-section"),T=document.getElementById("audit-trail-rows");y.innerHTML="",T.innerHTML="";const A=[];let h=0,E=0;for(let p=0;p<e.length;p++){const x=await k(e[p],p+1,d);A.push(x),x.status==="OK"?h++:E++,y.appendChild(x.element);const C=document.createElement("tr");C.style.borderBottom="1px solid #f1f5f9",C.innerHTML=`
        <td style="padding: 8px 10px; color: #64748b;">${new Date().toLocaleTimeString()}</td>
        <td style="padding: 8px 10px; font-weight: 500;">INSPECTION_METIER</td>
        <td style="padding: 8px 10px;">${x.fileName}</td>
        <td style="padding: 8px 10px; font-family: monospace; color: #475569;">${x.mockHash}</td>
        <td style="padding: 8px 10px;">
          <span style="font-weight: 600; font-size: 0.75rem; padding: 2px 6px; border-radius: 4px; ${x.status==="OK"?"background: #dcfce7; color: #15803d;":"background: #fee2e2; color: #b91c1c;"}">
            ${x.status==="OK"?"CONFORME":"REJETÉ"}
          </span>
        </td>
      `,T.appendChild(C)}c={batchId:d,timestamp:b,totalCount:e.length,acceptedCount:h,rejectedCount:E,invoices:A.map(p=>({fileName:p.fileName,fileSizeKo:p.fileSizeKo,format:p.format,status:p.status,checks:p.checks,corrections:p.corrections,mockHash:p.mockHash}))},g.style.display="block",E===0?(g.style.background="#f0fdf4",g.style.borderColor="#bbf7d0",g.innerHTML=`
        <div style="font-weight: 600; color: #166534; font-size: 1rem; display: flex; align-items: center; gap: 8px;">
          <span>✅</span> Notification Dépôt [Réf : ${d}] : 100% des factures validées
        </div>
        <div style="color: #15803d; font-size: 0.85rem; margin-top: 4px;">
          L'ensemble du lot (${h} facture(s)) répond intégralement aux exigences de conformité légale et fiscale. Prêt pour transmission.
        </div>
      `):(g.style.background="#fef2f2",g.style.borderColor="#fecaca",g.innerHTML=`
        <div style="font-weight: 600; color: #991b1b; font-size: 1rem; display: flex; align-items: center; gap: 8px;">
          <span>🔔</span> Notification Dépôt [Réf : ${d}] : Anomalies détectées sur le lot
        </div>
        <div style="color: #b91c1c; font-size: 0.85rem; margin-top: 4px;">
          Bilan du traitement : <strong>${h} facture(s) OK</strong> | <strong style="text-decoration: underline;">${E} facture(s) NON CONFORME(S)</strong>. Un rapport d'anomalies a été consolidé ci-dessous.
        </div>
      `),S.style.display="flex",O.style.display="block"}}function D(i){return new Promise(o=>{if(i.type==="application/pdf"||i.name.endsWith(".pdf")){o("");return}const r=new FileReader;r.onload=t=>o(t.target.result||""),r.onerror=()=>o(""),r.readAsText(i)})}function L(i,o,r){const t=new Blob([o],{type:r}),n=URL.createObjectURL(t),a=document.createElement("a");a.href=n,a.download=i,document.body.appendChild(a),a.click(),document.body.removeChild(a),URL.revokeObjectURL(n)}async function k(i,o,r){const t=await D(i),n=(i.size/1024).toFixed(1),a=i.name.split(".").pop().toLowerCase(),l="e3b0c442..."+Math.random().toString(16).substring(2,8);let f=[],u=[];if(a==="pdf")f=[{label:"Norme Archivage & Conteneur PDF/A-3",status:"OK",detail:"Profil ISO 19005-3 valide"},{label:"Contrôle Légal : Pièce jointe XML (factur-x.xml)",status:"FAILED",detail:"Fichier structuré absent du conteneur PDF"},{label:"Contrôle Fiscal : Mentions obligatoires CGI (Art. 242 nonies A)",status:"FAILED",detail:"Numéro de TVA intracommunautaire émetteur absent"},{label:"Règles Sectorielles B2G / Chorus Pro",status:"PENDING",detail:"En suspens"}],u=[{field:"Structure hybride Factur-X",issue:"Le PDF déposé est un fichier bureautique sans flux XML joint.",action:"Exportez la facture en mode Factur-X ou déposez directement le flux UBL/CII."},{field:"Identification fiscale de l'émetteur",issue:"Numéro de TVA ou SIREN manquant dans les métadonnées.",action:"Ajouter la mention TVA intracommunautaire du vendeur."}];else if(a==="xml"){const e=t.includes("cius-fr")||t.includes("EN16931"),d=t.includes("CompanyID")||t.includes("TaxScheme"),b=t.includes("TaxTotal");e&&d&&b?(f=[{label:"Syntaxe & Encodage UTF-8",status:"OK",detail:"XML bien formé"},{label:"Schéma EN 16931 XSD (UBL / CII)",status:"OK",detail:"Profil CIUS-FR conforme"},{label:"Contrôles Fiscaux : Ventilation de TVA",status:"OK",detail:"Taux et montants équilibrés"},{label:"Règles Légales CIUS-FR",status:"OK",detail:"Identifiants d'acheminement validés"}],u=[]):(f=[{label:"Syntaxe & Encodage UTF-8",status:"OK",detail:"XML bien formé"},{label:"Schéma EN 16931 XSD (UBL / CII)",status:e?"OK":"FAILED",detail:e?"Conforme":"CustomizationID CIUS-FR absent"},{label:"Contrôles Fiscaux : N° TVA (Art. 242 nonies A)",status:d?"OK":"FAILED",detail:d?"Conforme":"PartyTaxScheme émetteur manquant"},{label:"Règles Légales CIUS-FR : Bloc TVA",status:b?"OK":"FAILED",detail:b?"Conforme":"TaxTotal manquant"}],u=[{field:"Identifiant fiscal émetteur",issue:"Numéro de TVA intracommunautaire émetteur absent du bloc fournisseur.",action:"Déclarer la balise <cac:PartyTaxScheme> avec le numéro de TVA français."},{field:"Identifiant de personnalisation CIUS-FR",issue:"La personnalisation française EN 16931 n'est pas stipulée.",action:"Ajouter CustomizationID 'urn:cen.eu:en16931:2017#compliant#urn:facx.org:1p0:cius-fr'."}])}else a==="edi"?(f=[{label:"Syntaxe EDIFACT D01B",status:"OK",detail:"Segments UNH, BGM, DTM, MOA valides"},{label:"Contrôle Légal : Agrément fiscal EDI",status:"OK",detail:"Interchange régulier"},{label:"Transcodage vers norme EN 16931",status:"OK",detail:"Mapping CIUS-FR établi"}],u=[]):(f=[{label:"Extension",status:"FAILED",detail:`Type .${a} non pris en charge`}],u=[{field:"Format",issue:"Format refusé",action:"Utiliser .xml, .pdf Factur-X ou .edi."}]);const c=u.length===0,m=document.createElement("div");return m.style.border="1px solid var(--border-muted, #e2e8f0)",m.style.borderRadius="8px",m.style.padding="18px",m.style.background="var(--bg-card, #ffffff)",m.style.boxShadow="0 1px 3px rgba(0,0,0,0.03)",m.innerHTML=`
    <div style="display: flex; justify-content: space-between; align-items: flex-start; border-bottom: 1px solid var(--border-muted, #e2e8f0); padding-bottom: 12px; margin-bottom: 14px;">
      <div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="font-size: 1.1rem;">📄</span>
          <span style="font-weight: 600; font-size: 0.95rem; color: var(--text-primary, #0f172a);">${i.name}</span>
          <span style="font-size: 0.75rem; background: #e2e8f0; color: #334155; padding: 2px 6px; border-radius: 4px; font-weight: 600;">#${o}</span>
        </div>
        <div style="font-size: 0.8rem; color: var(--text-secondary, #64748b); margin-top: 4px;">
          Poids : <strong>${n} Ko</strong> | Format : <strong style="text-transform: uppercase;">${a}</strong> | Lot : <code>${r}</code>
        </div>
      </div>
      <span style="display: inline-block; padding: 4px 10px; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; ${c?"background: #dcfce7; color: #15803d;":"background: #fee2e2; color: #b91c1c;"}">
        ${c?"✔ Conforme Fiscale & Légale":"✖ Rejet des Contrôles"}
      </span>
    </div>

    <div style="margin-bottom: 14px;">
      <div style="font-weight: 600; font-size: 0.825rem; color: var(--text-primary, #0f172a); margin-bottom: 8px; text-transform: uppercase;">
        ⚖️ Grille des contrôles légaux et fiscaux
      </div>
      <div style="display: flex; flex-direction: column; gap: 6px;">
        ${f.map(e=>`
          <div style="display: flex; justify-content: space-between; align-items: center; background: var(--bg-subtle, #f8fafc); padding: 7px 12px; border-radius: 6px; border: 1px solid var(--border-muted, #e2e8f0); font-size: 0.8rem;">
            <div>
              <span style="font-weight: 500; color: var(--text-primary, #0f172a);">${e.label}</span>
              <div style="font-size: 0.75rem; color: var(--text-secondary, #64748b);">${e.detail}</div>
            </div>
            <div>
              ${e.status==="OK"?'<span style="background: #dcfce7; color: #166534; font-size: 0.7rem; font-weight: 700; padding: 2px 6px; border-radius: 4px;">CONFORME</span>':""}
              ${e.status==="FAILED"?'<span style="background: #fee2e2; color: #991b1b; font-size: 0.7rem; font-weight: 700; padding: 2px 6px; border-radius: 4px;">NON CONFORME</span>':""}
              ${e.status==="PENDING"?'<span style="background: #f1f5f9; color: #475569; font-size: 0.7rem; font-weight: 700; padding: 2px 6px; border-radius: 4px;">EN SUSPENS</span>':""}
            </div>
          </div>
        `).join("")}
      </div>
    </div>

    <div>
      ${u.length>0?`
        <div style="background: #fffbeb; border: 1px solid #fde68a; border-radius: 6px; padding: 12px;">
          <div style="font-weight: 600; font-size: 0.825rem; color: #92400e; margin-bottom: 8px; display: flex; align-items: center; gap: 6px;">
            <span>⚠️</span> Ce qu'il faut corriger sur cette facture :
          </div>
          <div style="display: flex; flex-direction: column; gap: 8px;">
            ${u.map(e=>`
              <div style="background: #ffffff; border-left: 3px solid #f59e0b; padding: 8px 12px; border-radius: 0 4px 4px 0; font-size: 0.8rem;">
                <div style="font-weight: 600; color: #b45309; margin-bottom: 2px;">${e.field}</div>
                <div style="color: #475569; margin-bottom: 2px;"><strong>Anomalie :</strong> ${e.issue}</div>
                <div style="color: #1e293b;"><strong>Action recommandée :</strong> ${e.action}</div>
              </div>
            `).join("")}
          </div>
        </div>
      `:`
        <div style="background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 6px; padding: 10px 14px; color: #166534; font-size: 0.825rem; font-weight: 500;">
          ✔ Tous les contrôles fiscaux (CGI art. 242) et techniques (EN 16931) sont validés pour cette facture.
        </div>
      `}
    </div>
  `,{fileName:i.name,fileSizeKo:n,format:a,status:c?"OK":"REJECTED",checks:f,corrections:u,mockHash:l,element:m}}let s={role:"CLIENT",view:"client-upload",theme:"light"};function I(){document.documentElement.setAttribute("data-role",s.role),document.documentElement.setAttribute("data-theme",s.theme);const i=v[s.role],o=i.routes.find(t=>t.id===s.view)||i.routes[0];let r="";s.role==="ADMIN"&&s.view==="admin-users"?r=w():r=F(s.view,o.label,s.role,i),document.getElementById("app").innerHTML=N(s.role,s.view,r),s.role==="CLIENT"&&s.view==="client-upload"&&$("dropzone-root"),document.getElementById("role-selector").addEventListener("change",t=>{s.role=t.target.value,s.view=v[s.role].routes[0].id,s.theme=v[s.role].defaultTheme,I()}),document.querySelectorAll("[data-nav]").forEach(t=>{t.addEventListener("click",n=>{s.view=n.target.getAttribute("data-nav"),I()})}),document.getElementById("theme-toggle").addEventListener("click",()=>{s.theme=s.theme==="light"?"dark":"light",I()})}window.triggerRevoke=function(i){confirm(`Voulez-vous révoquer les accès de l'identifiant ${i} ?`)&&alert(`Accès de l'utilisateur ${i} révoqué.`)};I();
