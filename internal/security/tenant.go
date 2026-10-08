package security

import (
	"context"
	"errors"
)

type ctxKey int

const claimsKey ctxKey = 1

var ErrTenantMismatch = errors.New("security: tenant mismatch")

func ContextWithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

// AssertTenant garantit que la ressource appartient au tenant du token.
func AssertTenant(ctx context.Context, resourceTenantID string) error {
	c, ok := ClaimsFromContext(ctx)
	if !ok || c == nil {
		return ErrTokenClaims
	}
	if c.TenantID != resourceTenantID {
		return ErrTenantMismatch
	}
	return nil
}

// RequirePermission combine authz RBAC.
func RequirePermission(ctx context.Context, perm Permission) error {
	c, ok := ClaimsFromContext(ctx)
	if !ok || c == nil {
		return ErrTokenClaims
	}
	if !HasPermission(c.Roles, perm) {
		return errors.New("security: forbidden")
	}
	return nil
}
