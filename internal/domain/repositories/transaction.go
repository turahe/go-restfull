package repositories

import (
	"context"
)

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error
}

// TransactionManager defines the interface for managing database transactions
type TransactionManager interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) (Transaction, error)

	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(Transaction) error) error

	// WithReadOnlyTransaction executes a function within a read-only transaction
	WithReadOnlyTransaction(ctx context.Context, fn func(Transaction) error) error
}

// TransactionalRepository defines the interface for repositories that support transactions
type TransactionalRepository interface {
	// WithTransaction executes repository operations within a transaction
	WithTransaction(ctx context.Context, fn func(Transaction) error) error

	// BeginTransaction starts a new transaction for this repository
	BeginTransaction(ctx context.Context) (Transaction, error)
}
