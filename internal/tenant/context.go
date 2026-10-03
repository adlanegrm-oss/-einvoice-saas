package tenant

import (
	"context"
	"errors"
)

type contextKey string

const tenantContextKey contextKey = "tenant_id"
const scopesContextKey contextKey = "token_scopes"

var (
	ErrUnauthorizedTenant = errors.New("aucun tenant valide identifié dans le contexte")
	ErrForbiddenScope     = errors.New("portée d'autorisation insuffisante pour cette opération")
)

func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey, tenantID)
}

func FromContext(ctx context.Context) (string, error) {
	val, ok := ctx.Value(tenantContextKey).(string)
	if !ok || val == "" {
		return "", ErrUnauthorizedTenant
	}
	return val, nil
}

func WithScopes(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, scopesContextKey, scopes)
}

func RequireScope(ctx context.Context, required string) error {
	scopes, ok := ctx.Value(scopesContextKey).([]string)
	if !ok {
		return ErrForbiddenScope
	}
	for _, s := range scopes {
		if s == required || s == "admin" {
			return nil
		}
	}
	return ErrForbiddenScope
}
