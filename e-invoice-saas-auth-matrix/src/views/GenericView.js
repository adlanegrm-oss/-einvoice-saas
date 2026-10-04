import { initDropZone } from '../components/DropZone.js';

export function renderGenericView(moduleKey, moduleTitle, roleKey, roleCfg) {
  const baseRoute = (roleCfg && roleCfg.baseRoute) ? roleCfg.baseRoute : (roleKey ? roleKey.toLowerCase() : 'client');

  // Déclencher automatiquement l'initialisation dès que l'élément est dans le DOM
  setTimeout(() => {
    initDropZone('dropzone-root');
  }, 50);

  return `
    <div class="module-card">
      <div class="module-header">
        <h1 class="module-title">Module ${moduleKey || 'CLIENT'}</h1>
        <p class="module-subtitle">Interface opérationnelle pour le profil <strong>${roleKey || 'UTILISATEUR'}</strong>.</p>
      </div>

      <div class="code-banner">
        <div><code>CONTEXT_ROLE: ${roleKey || 'CLIENT'}</code></div>
        <div><code>ENDPOINT_BINDING: /api/v1/${baseRoute}/${moduleKey || 'upload'}</code></div>
        <div><code>ENFORCE_SCHEMA: EN_16931_CIUS_FR</code></div>
      </div>

      <div id="dropzone-root" style="margin-top: 24px;"></div>
    </div>
  `;
}

export function bindGenericViewEvents(moduleKey, roleKey, moduleTitle) {
  initDropZone('dropzone-root');
}
