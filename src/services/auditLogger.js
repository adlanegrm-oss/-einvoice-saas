const auditStore = [
  {
    id: "AUD-9021",
    timestamp: "2026-10-04 18:15:20",
    userId: "usr_003",
    userEmail: "m.lefebvre@company.fr",
    action: "VALIDATE_INVOICE",
    target: "INV-2026-00125",
    establishment: "Agence Paris Nord",
    before: { status: "PENDING_VALIDATION" },
    after: { status: "VALIDATED" },
    hash: "3a9f029c78d10b42f..."
  },
  {
    id: "AUD-9020",
    timestamp: "2026-10-04 17:42:01",
    userId: "usr_001",
    userEmail: "j.dupont@company.fr",
    action: "CREATE_CUSTOMER",
    target: "CLI-4902 (TechCorp SAS)",
    establishment: "Agence Lyon Centre",
    before: null,
    after: { siret: "84920194800012", vat: "FR49849201948" },
    hash: "b890ef41094a9d721..."
  }
];

export function getAuditLogs() {
  return [...auditStore];
}

export function logAuditEvent(event) {
  const entry = {
    id: `AUD-${Math.floor(1000 + Math.random() * 9000)}`,
    timestamp: new Date().toISOString().replace('T', ' ').substring(0, 19),
    ...event,
    hash: Math.random().toString(36).substring(2) + "..."
  };
  auditStore.unshift(entry);
}
