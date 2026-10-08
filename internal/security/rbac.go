package security

import "strings"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleAuditor  Role = "auditor"
)

// Permission actions métier.
type Permission string

const (
	PermInvoiceRead     Permission = "invoice:read"
	PermInvoiceWrite    Permission = "invoice:write"
	PermInvoiceTransmit Permission = "invoice:transmit"
	PermAuditRead       Permission = "audit:read"
	PermAdmin           Permission = "admin:*"
)

var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermInvoiceRead, PermInvoiceWrite, PermInvoiceTransmit, PermAuditRead, PermAdmin,
	},
	RoleOperator: {
		PermInvoiceRead, PermInvoiceWrite, PermInvoiceTransmit,
	},
	RoleAuditor: {
		PermInvoiceRead, PermAuditRead,
	},
}

func HasPermission(roles []string, perm Permission) bool {
	for _, r := range roles {
		role := Role(strings.ToLower(strings.TrimSpace(r)))
		for _, p := range rolePermissions[role] {
			if p == perm || p == PermAdmin {
				return true
			}
		}
	}
	return false
}

func HasAnyRole(roles []string, want ...Role) bool {
	set := map[string]struct{}{}
	for _, r := range roles {
		set[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[string(w)]; ok {
			return true
		}
	}
	return false
}
