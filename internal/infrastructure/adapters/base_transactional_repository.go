package adapters

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/domain/repositories"
)

// BaseTransactionalRepository provides common transaction functionality
type BaseTransactionalRepository struct {
	pool               *pgxpool.Pool
	transactionManager repositories.TransactionManager
}

// NewBaseTransactionalRepository creates a new base transactional repository
func NewBaseTransactionalRepository(pool *pgxpool.Pool) *BaseTransactionalRepository {
	return &BaseTransactionalRepository{
		pool:               pool,
		transactionManager: NewPostgresTransactionManager(pool),
	}
}

// WithTransaction executes repository operations within a transaction
func (b *BaseTransactionalRepository) WithTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	return b.transactionManager.WithTransaction(ctx, fn)
}

// WithReadOnlyTransaction executes repository operations within a read-only transaction
func (b *BaseTransactionalRepository) WithReadOnlyTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	return b.transactionManager.WithReadOnlyTransaction(ctx, fn)
}

// BeginTransaction starts a new transaction for this repository
func (b *BaseTransactionalRepository) BeginTransaction(ctx context.Context) (repositories.Transaction, error) {
	return b.transactionManager.Begin(ctx)
}

// GetTransactionManager returns the transaction manager for advanced usage
func (b *BaseTransactionalRepository) GetTransactionManager() repositories.TransactionManager {
	return b.transactionManager
}

// GetPool returns the underlying connection pool
func (b *BaseTransactionalRepository) GetPool() *pgxpool.Pool {
	return b.pool
}

// ExecuteInTransaction is a helper method to execute multiple operations in a single transaction
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

// ExecuteReadOnlyInTransaction is a helper method to execute multiple read operations in a single read-only transaction
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
