import { initDropZone } from '../components/DropZone.js';

export function renderGenericView(moduleKey, moduleTitle, roleKey, roleCfg) {
  const isClientUpload = moduleKey === 'client-upload';

  return `
    <div class="module-card">
      <div class="module-header">
        <h1 class="module-title">Module ${moduleKey}</h1>
        <p class="module-subtitle">Interface opérationnelle pour le profil <strong>${roleKey}</strong>.</p>
      </div>

      <div class="code-banner">
        <div><code>CONTEXT_ROLE: ${roleKey}</code></div>
        <div><code>ENDPOINT_BINDING: /api/v1/${roleCfg.baseRoute || 'client'}/${moduleKey}</code></div>
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
