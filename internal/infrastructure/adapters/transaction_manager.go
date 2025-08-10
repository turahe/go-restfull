package adapters

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/domain/repositories"
)

// pgxTransaction wraps pgx.Tx to implement the Transaction interface
type pgxTransaction struct {
	tx pgx.Tx
}

func (t *pgxTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// PostgresTransactionManager implements TransactionManager using pgx
type PostgresTransactionManager struct {
	pool *pgxpool.Pool
}

// NewPostgresTransactionManager creates a new transaction manager
func NewPostgresTransactionManager(pool *pgxpool.Pool) repositories.TransactionManager {
	return &PostgresTransactionManager{
		pool: pool,
	}
}

// Begin starts a new transaction
func (tm *PostgresTransactionManager) Begin(ctx context.Context) (repositories.Transaction, error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &pgxTransaction{tx: tx}, nil
}

// WithTransaction executes a function within a transaction
func (tm *PostgresTransactionManager) WithTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Execute the function
	if err := fn(&pgxTransaction{tx: tx}); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithReadOnlyTransaction executes a function within a read-only transaction
func (tm *PostgresTransactionManager) WithReadOnlyTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read-only transaction: %w", err)
	}

	// Set transaction as read-only
	_, err = tx.Exec(ctx, "SET TRANSACTION READ ONLY")
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to set transaction as read-only: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Execute the function
	if err := fn(&pgxTransaction{tx: tx}); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("read-only transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit read-only transaction: %w", err)
	}

	return nil
}
