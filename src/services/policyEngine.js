// Matrice granulaire de permissions avec gestion des Scopes
export const PERMISSIONS = {
  INVOICE_READ: "invoice.read",
  INVOICE_CREATE: "invoice.create",
  INVOICE_UPDATE: "invoice.update",
  INVOICE_VALIDATE: "invoice.validate",
  INVOICE_ISSUE: "invoice.issue",
  INVOICE_CANCEL: "invoice.cancel",
  INVOICE_EXPORT: "invoice.export",
  CUSTOMER_MANAGE: "customer.manage",
  USER_MANAGE: "user.manage",
  ORGANIZATION_UPDATE: "organization.update",
  AUDIT_READ: "audit.read"
};

export const SCOPES = {
  ALL_ORGANIZATION: "ALL_ORGANIZATION",
  OWN_ESTABLISHMENT: "OWN_ESTABLISHMENT",
  OWN_DOCUMENTS: "OWN_DOCUMENTS"
};

export const ROLES_DEFINITION = {
  ADMIN: {
    label: "Administrateur Global",
    scope: SCOPES.ALL_ORGANIZATION,
    permissions: Object.values(PERMISSIONS)
  },
  FINANCE_ADMIN: {
    label: "Administrateur Financier",
    scope: SCOPES.ALL_ORGANIZATION,
    permissions: [
      PERMISSIONS.INVOICE_READ, PERMISSIONS.INVOICE_CREATE, PERMISSIONS.INVOICE_UPDATE,
      PERMISSIONS.INVOICE_VALIDATE, PERMISSIONS.INVOICE_ISSUE, PERMISSIONS.INVOICE_EXPORT,
      PERMISSIONS.CUSTOMER_MANAGE, PERMISSIONS.AUDIT_READ
    ]
  },
  ACCOUNTANT: {
    label: "Comptable",
    scope: SCOPES.OWN_ESTABLISHMENT,
    permissions: [
      PERMISSIONS.INVOICE_READ, PERMISSIONS.INVOICE_CREATE, PERMISSIONS.INVOICE_UPDATE,
      PERMISSIONS.INVOICE_VALIDATE, PERMISSIONS.INVOICE_EXPORT, PERMISSIONS.AUDIT_READ
    ]
  },
  SALES: {
    label: "Commercial",
    scope: SCOPES.OWN_DOCUMENTS,
    permissions: [
      PERMISSIONS.INVOICE_READ, PERMISSIONS.INVOICE_CREATE, PERMISSIONS.CUSTOMER_MANAGE
    ]
  }
};

export function can(user, permission, targetResource = null) {
  const roleDef = ROLES_DEFINITION[user.role];
  if (!roleDef) return false;

  const hasPerm = roleDef.permissions.includes(permission) || (user.directPermissions && user.directPermissions.includes(permission));
  if (!hasPerm) return false;

  // Contrôle de Scope
  if (roleDef.scope === SCOPES.ALL_ORGANIZATION) return true;
  if (!targetResource) return true;

  if (roleDef.scope === SCOPES.OWN_ESTABLISHMENT) {
    return targetResource.establishmentId === user.establishmentId;
  }
  if (roleDef.scope === SCOPES.OWN_DOCUMENTS) {
    return targetResource.createdBy === user.id;
  }
  return false;
}
