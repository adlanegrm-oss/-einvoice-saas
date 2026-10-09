package repository

import "context"

// TransactionRunner exécute une opération dans une transaction atomique.
// Si lockKey est non vide, l'implémentation peut sérialiser les opérations
// qui utilisent la même clé.
type TransactionRunner interface {
	WithinTransaction(
		ctx context.Context,
		lockKey string,
		fn func(context.Context) error,
	) error
}
