// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/domain/repositories"
)

// BaseTransactionalRepository provides common transaction functionality for all repository implementations.
// It serves as a base struct that other repositories can embed to inherit transaction capabilities.
// This struct manages database connections and provides a consistent interface for transaction operations.
type BaseTransactionalRepository struct {
	// pool holds the PostgreSQL connection pool for database operations
	pool *pgxpool.Pool
	// transactionManager handles transaction lifecycle and provides transaction interfaces
	transactionManager repositories.TransactionManager
}

// NewBaseTransactionalRepository creates a new base transactional repository instance.
// It initializes the repository with a connection pool and creates a new transaction manager.
//
// Parameters:
//   - pool: The PostgreSQL connection pool to use for database operations
//
// Returns:
//   - *BaseTransactionalRepository: A new base transactional repository instance
func NewBaseTransactionalRepository(pool *pgxpool.Pool) *BaseTransactionalRepository {
	return &BaseTransactionalRepository{
		pool:               pool,
		transactionManager: NewPostgresTransactionManager(pool),
	}
}

// WithTransaction executes repository operations within a database transaction.
// If the function returns an error, the transaction is automatically rolled back.
// If the function completes successfully, the transaction is committed.
//
// Parameters:
//   - ctx: Context for the transaction operation
//   - fn: Function to execute within the transaction context
//
// Returns:
//   - error: Any error that occurred during transaction execution or rollback
func (b *BaseTransactionalRepository) WithTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	return b.transactionManager.WithTransaction(ctx, fn)
}

// WithReadOnlyTransaction executes repository operations within a read-only database transaction.
// This is useful for operations that only need to read data and don't require write locks.
// Read-only transactions can be more performant and have less impact on concurrent operations.
//
// Parameters:
//   - ctx: Context for the transaction operation
//   - fn: Function to execute within the read-only transaction context
//
// Returns:
//   - error: Any error that occurred during transaction execution
func (b *BaseTransactionalRepository) WithReadOnlyTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	return b.transactionManager.WithReadOnlyTransaction(ctx, fn)
}

// BeginTransaction starts a new transaction for this repository.
// The caller is responsible for managing the transaction lifecycle (commit/rollback).
// This method is useful when you need more control over transaction management.
//
// Parameters:
//   - ctx: Context for the transaction operation
//
// Returns:
//   - repositories.Transaction: A transaction interface for database operations
//   - error: Any error that occurred while starting the transaction
func (b *BaseTransactionalRepository) BeginTransaction(ctx context.Context) (repositories.Transaction, error) {
	return b.transactionManager.Begin(ctx)
}

// GetTransactionManager returns the transaction manager for advanced usage.
// This allows direct access to the transaction manager when custom transaction logic is needed.
//
// Returns:
//   - repositories.TransactionManager: The transaction manager instance
func (b *BaseTransactionalRepository) GetTransactionManager() repositories.TransactionManager {
	return b.transactionManager
}

// GetPool returns the underlying connection pool.
// This method provides access to the raw connection pool for advanced database operations
// that are not covered by the standard transaction methods.
//
// Returns:
//   - *pgxpool.Pool: The PostgreSQL connection pool
func (b *BaseTransactionalRepository) GetPool() *pgxpool.Pool {
	return b.pool
}

// ExecuteInTransaction is a helper method to execute multiple operations in a single transaction.
// All operations are executed sequentially within the same transaction context.
// If any operation fails, the entire transaction is rolled back.
//
// Parameters:
//   - ctx: Context for the transaction operation
//   - operations: Variable number of functions to execute within the transaction
//
// Returns:
//   - error: Any error that occurred during operation execution or transaction management
func (b *BaseTransactionalRepository) ExecuteInTransaction(ctx context.Context, operations ...func(repositories.Transaction) error) error {
	return b.WithTransaction(ctx, func(tx repositories.Transaction) error {
		for _, op := range operations {
			if err := op(tx); err != nil {
				return fmt.Errorf("operation failed in transaction: %w", err)
			}
		}
		return nil
	})
}

// ExecuteReadOnlyInTransaction is a helper method to execute multiple read operations in a single read-only transaction.
// This method is optimized for scenarios where multiple read operations need to be consistent
// but don't require write locks or transaction isolation.
//
// Parameters:
//   - ctx: Context for the transaction operation
//   - operations: Variable number of read-only functions to execute within the transaction
//
// Returns:
//   - error: Any error that occurred during operation execution
func (b *BaseTransactionalRepository) ExecuteReadOnlyInTransaction(ctx context.Context, operations ...func(repositories.Transaction) error) error {
	return b.WithReadOnlyTransaction(ctx, func(tx repositories.Transaction) error {
		for _, op := range operations {
			if err := op(tx); err != nil {
				return fmt.Errorf("read operation failed in transaction: %w", err)
			}
		}
		return nil
	})
}
