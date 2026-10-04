export function renderGenericView(roleKey, viewId) {
  return `
    <div class="card">
      <h2>Module ${viewId}</h2>
      <p style="margin-top:8px; color:var(--text-secondary);">
        Interface opérationnelle pour le profil <b>${roleKey}</b>.
      </p>
      <div style="margin-top:20px; padding:16px; background:var(--bg-subtle); border-radius:6px; font-family:monospace; font-size:0.85rem;">
        CONTEXT_ROLE: ${roleKey}<br>
        ENDPOINT_BINDING: /api/v1/${roleKey.toLowerCase()}/${viewId}<br>
        ENFORCE_SCHEMA: EN_16931_CIUS_FR
      </div>
    </div>
  `;
}
