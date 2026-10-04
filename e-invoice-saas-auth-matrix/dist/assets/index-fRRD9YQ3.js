(function(){const e=document.createElement("link").relList;if(e&&e.supports&&e.supports("modulepreload"))return;for(const t of document.querySelectorAll('link[rel="modulepreload"]'))r(t);new MutationObserver(t=>{for(const n of t)if(n.type==="childList")for(const o of n.addedNodes)o.tagName==="LINK"&&o.rel==="modulepreload"&&r(o)}).observe(document,{childList:!0,subtree:!0});function l(t){const n={};return t.integrity&&(n.integrity=t.integrity),t.referrerPolicy&&(n.referrerPolicy=t.referrerPolicy),t.crossOrigin==="use-credentials"?n.credentials="include":t.crossOrigin==="anonymous"?n.credentials="omit":n.credentials="same-origin",n}function r(t){if(t.ep)return;t.ep=!0;const n=l(t);fetch(t.href,n)}})();const s={CLIENT:{label:"Front Client",path:"/app",color:"#2563eb",defaultTheme:"light",description:"Utilisation quotidienne métier (dépôt, suivi, téléchargement)",routes:[{id:"client-upload",label:"Déposer des factures"},{id:"client-invoices",label:"Liste des factures"},{id:"client-validation",label:"Validation seule"}]},PARTNER:{label:"Front Partenaire",path:"/partner",color:"#7c3aed",defaultTheme:"light",description:"Intégrateurs, cabinets, ERP, PDP partenaires (multi-tenants)",routes:[{id:"partner-dash",label:"Dashboard Consolidé"},{id:"partner-clients",label:"Gestion des Clients"},{id:"partner-api",label:"Console API & Sandbox"}]},ADMIN:{label:"Console Admin",path:"/admin",color:"#ea580c",defaultTheme:"dark",description:"Administration globale multi-tenant, contrôle et gouvernance",routes:[{id:"admin-users",label:"Utilisateurs & Rôles"},{id:"admin-tenants",label:"Gestion des Tenants"},{id:"admin-outbox",label:"Supervision Outbox & Audit"}]},OPS:{label:"Console Exploitation",path:"/ops",color:"#059669",defaultTheme:"dark",description:"Supervision production, incidents, santé workers, DLQ",routes:[{id:"ops-metrics",label:"Santé Système & Métriques"},{id:"ops-dlq",label:"Outbox & File DLQ"}]},MCO:{label:"Console MCO",path:"/mco",color:"#475569",defaultTheme:"dark",description:"Maintenance, support N2/N3, runbooks et plans PRA",routes:[{id:"mco-runbooks",label:"Runbooks d'exploitation"},{id:"mco-pra",label:"Procédures PRA & Backup"}]},RND:{label:"Console R&D",path:"/rnd",color:"#4f46e5",defaultTheme:"dark",description:"Sandbox de conformité, règles BR-xx et qualification CIUS-FR",routes:[{id:"rnd-schematron",label:"Banc de Test Schematron"},{id:"rnd-golden",label:"Comparateur Golden Files"}]},DEPLOY:{label:"Console Déploiement",path:"/deploy",color:"#0d9488",defaultTheme:"dark",description:"Releases, statuts environnements, CI/CD et rollbacks",routes:[{id:"deploy-envs",label:"Statut des Environnements"},{id:"deploy-pipeline",label:"Pipelines CI/CD & Scans"}]}};function c(i,e,l){const r=s[i],t=r.routes.map(o=>`<li><a class="${o.id===e?"active":""}" data-nav="${o.id}">${o.label}</a></li>`).join("");return`
    <div class="layout-shell">
      <aside class="sidebar">
        <div>
          <div class="brand">e-invoice-saas</div>
          <div style="padding:12px;">
            <label style="font-size:0.75rem; color:var(--text-secondary); font-weight:600; text-transform:uppercase;">Changer d'espace :</label>
            <select id="role-selector" style="width:100%; margin-top:6px; padding:8px; border-radius:6px; border:1px solid var(--border-muted); background:var(--bg-subtle); color:var(--text-primary); font-size:0.825rem;">
              ${Object.keys(s).map(o=>`<option value="${o}" ${o===i?"selected":""}>${s[o].label} (${s[o].path})</option>`).join("")}
            </select>
          </div>
          <ul class="nav-menu">
            ${t}
          </ul>
        </div>
        <div style="padding:16px; border-top:1px solid var(--border-muted); font-size:0.8rem;">
          <div style="color:var(--text-secondary);">Espace Actif</div>
          <div style="font-weight:600; color:var(--accent-primary);">${r.label}</div>
          <div style="font-size:0.75rem; color:var(--text-disabled); margin-top:4px;">Accès : ${r.path}</div>
        </div>
      </aside>
      <main class="main-content">
        <header class="topbar">
          <div>
            <span style="font-size:0.875rem; color:var(--text-secondary);">${r.description}</span>
          </div>
          <div style="display:flex; align-items:center; gap:12px;">
            <span class="badge badge-valid">Système Opérationnel</span>
            <button id="theme-toggle" class="btn btn-subtle" style="font-size:0.75rem;">Bascule Clair/Sombre</button>
          </div>
        </header>
        <div class="content-body">
          ${l}
        </div>
      </main>
    </div>
  `}const u=[{id:"usr_001",name:"Jean Dupont",email:"jean.dupont@client.fr",role:"CLIENT",tenant:"dupont-fr",status:"ACTIF"},{id:"usr_002",name:"Claire Valette",email:"c.valette@nexus-partner.io",role:"PARTNER",tenant:"nexus-audit",status:"ACTIF"},{id:"usr_003",name:"Marc Lefebvre",email:"m.lefebvre@einvoice-saas.com",role:"ADMIN",tenant:"master",status:"ACTIF"},{id:"usr_004",name:"Alexandre Roux",email:"ops@einvoice-saas.com",role:"OPS",tenant:"master",status:"ACTIF"},{id:"usr_005",name:"Équipe MCO",email:"mco-support@einvoice-saas.com",role:"MCO",tenant:"master",status:"ACTIF"},{id:"usr_006",name:"Ingénierie R&D",email:"rnd-engine@einvoice-saas.com",role:"RND",tenant:"master",status:"ACTIF"},{id:"usr_007",name:"Release Lead",email:"deploy-bot@einvoice-saas.com",role:"DEPLOY",tenant:"master",status:"ACTIF"}];function p(){return`
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
          ${u.map(e=>`
    <tr>
      <td><b>${e.name}</b></td>
      <td>${e.email}</td>
      <td><span class="badge badge-valid">${e.role}</span></td>
      <td><code>${e.tenant}</code></td>
      <td><span class="badge badge-valid">${e.status}</span></td>
      <td>
        <button class="btn btn-subtle" style="padding:4px 8px;" onclick="window.triggerRevoke('${e.id}')">Révoquer</button>
      </td>
    </tr>
  `).join("")}
        </tbody>
      </table>
    </div>
  `}function m(i,e){return`
    <div class="card">
      <h2>Module ${e}</h2>
      <p style="margin-top:8px; color:var(--text-secondary);">
        Interface opérationnelle pour le profil <b>${i}</b>.
      </p>
      <div style="margin-top:20px; padding:16px; background:var(--bg-subtle); border-radius:6px; font-family:monospace; font-size:0.85rem;">
        CONTEXT_ROLE: ${i}<br>
        ENDPOINT_BINDING: /api/v1/${i.toLowerCase()}/${e}<br>
        ENFORCE_SCHEMA: EN_16931_CIUS_FR
      </div>
    </div>
  `}let a={role:"CLIENT",view:"client-upload",theme:"light"};function d(){document.documentElement.setAttribute("data-role",a.role),document.documentElement.setAttribute("data-theme",a.theme);let i="";a.role==="ADMIN"&&a.view==="admin-users"?i=p():i=m(a.role,a.view),document.getElementById("app").innerHTML=c(a.role,a.view,i),document.getElementById("role-selector").addEventListener("change",e=>{a.role=e.target.value,a.view=s[a.role].routes[0].id,a.theme=s[a.role].defaultTheme,d()}),document.querySelectorAll("[data-nav]").forEach(e=>{e.addEventListener("click",l=>{a.view=l.target.getAttribute("data-nav"),d()})}),document.getElementById("theme-toggle").addEventListener("click",()=>{a.theme=a.theme==="light"?"dark":"light",d()})}window.triggerRevoke=function(i){confirm(`Voulez-vous révoquer les accès de l'identifiant ${i} ?`)&&alert(`Accès de l'utilisateur ${i} révoqué.`)};d();
