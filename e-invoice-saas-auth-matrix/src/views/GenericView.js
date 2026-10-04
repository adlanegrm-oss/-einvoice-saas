import { initDropZone } from '../components/DropZone.js';

export function renderGenericView(moduleKey, moduleTitle, roleKey, roleCfg) {
  const isClientUpload = moduleKey === 'client-upload';
  const baseRoute = (roleCfg && roleCfg.baseRoute) ? roleCfg.baseRoute : (roleKey ? roleKey.toLowerCase() : 'client');

  return `
    <div class="module-card">
      <div class="module-header">
        <h1 class="module-title">Module ${moduleKey || ''}</h1>
        <p class="module-subtitle">Interface opérationnelle pour le profil <strong>${roleKey || 'UTILISATEUR'}</strong>.</p>
      </div>

      <div class="code-banner">
        <div><code>CONTEXT_ROLE: ${roleKey || 'CLIENT'}</code></div>
        <div><code>ENDPOINT_BINDING: /api/v1/${baseRoute}/${moduleKey || ''}</code></div>
        <div><code>ENFORCE_SCHEMA: EN_16931_CIUS_FR</code></div>
      </div>

      ${isClientUpload ? `<div id="dropzone-root"></div>` : ''}
    </div>
  `;
}

export function bindGenericViewEvents(moduleKey) {
  if (moduleKey === 'client-upload') {
    initDropZone('dropzone-root');
  }
}
