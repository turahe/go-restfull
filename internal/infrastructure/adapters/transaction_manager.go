// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/domain/repositories"
)

// pgxTransaction wraps pgx.Tx to implement the Transaction interface.
// This adapter provides a clean abstraction over the pgx transaction implementation,
// allowing the domain layer to work with a generic transaction interface.
type pgxTransaction struct {
	// tx holds the underlying pgx transaction instance
	tx pgx.Tx
}

// Commit commits the current transaction, making all changes permanent.
// This method delegates to the underlying pgx transaction's Commit method.
//
// Parameters:
//   - ctx: Context for the commit operation
//
// Returns:
//   - error: Any error that occurred during the commit operation
func (t *pgxTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the current transaction, undoing all changes.
// This method delegates to the underlying pgx transaction's Rollback method.
//
// Parameters:
//   - ctx: Context for the rollback operation
//
// Returns:
//   - error: Any error that occurred during the rollback operation
func (t *pgxTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// PostgresTransactionManager implements TransactionManager using pgx for PostgreSQL.
// This struct manages the lifecycle of database transactions, providing both
// automatic and manual transaction management capabilities.
type PostgresTransactionManager struct {
	// pool holds the PostgreSQL connection pool for creating new transactions
	pool *pgxpool.Pool
}

// NewPostgresTransactionManager creates a new PostgreSQL transaction manager instance.
// It initializes the manager with a connection pool for creating transactions.
//
// Parameters:
//   - pool: The PostgreSQL connection pool to use for transaction creation
//
// Returns:
//   - repositories.TransactionManager: A new transaction manager instance
func NewPostgresTransactionManager(pool *pgxpool.Pool) repositories.TransactionManager {
	return &PostgresTransactionManager{
		pool: pool,
	}
}

// Begin starts a new database transaction.
// The caller is responsible for managing the transaction lifecycle (commit/rollback).
// This method is useful when you need fine-grained control over transaction management.
//
// Parameters:
//   - ctx: Context for the transaction operation
//
// Returns:
//   - repositories.Transaction: A transaction interface for database operations
//   - error: Any error that occurred while starting the transaction
func (tm *PostgresTransactionManager) Begin(ctx context.Context) (repositories.Transaction, error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &pgxTransaction{tx: tx}, nil
}

// WithTransaction executes a function within a database transaction.
// This method provides automatic transaction management - if the function returns an error,
// the transaction is automatically rolled back. If the function completes successfully,
// the transaction is committed. The method also handles panic recovery and cleanup.
//
// Parameters:
//   - ctx: Context for the transaction operation
//   - fn: Function to execute within the transaction context
//
// Returns:
//   - error: Any error that occurred during transaction execution, rollback, or commit
func (tm *PostgresTransactionManager) WithTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on panic recovery
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Execute the function within the transaction context
	if err := fn(&pgxTransaction{tx: tx}); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction if all operations succeeded
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithReadOnlyTransaction executes a function within a read-only database transaction.
// This method is optimized for operations that only need to read data and don't require
// write locks. Read-only transactions can be more performant and have less impact on
// concurrent operations. The method automatically sets the transaction as read-only
// and provides the same automatic management as WithTransaction.
//
// Parameters:
//   - ctx: Context for the transaction operation
//   - fn: Function to execute within the read-only transaction context
//
// Returns:
//   - error: Any error that occurred during transaction execution, rollback, or commit
func (tm *PostgresTransactionManager) WithReadOnlyTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read-only transaction: %w", err)
	}

	// Set transaction as read-only for performance optimization
	_, err = tx.Exec(ctx, "SET TRANSACTION READ ONLY")
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to set transaction as read-only: %w", err)
	}

	// Ensure rollback on panic recovery
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Execute the function within the read-only transaction context
	if err := fn(&pgxTransaction{tx: tx}); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("read-only transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit the read-only transaction if all operations succeeded
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit read-only transaction: %w", err)
	}

	return nil
}
