package adapters

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// BaseRepository provides common CRUD operations for entities
type BaseRepository[T any] struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
	tableName   string
	entityType  reflect.Type
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](pgxPool *pgxpool.Pool, redisClient redis.Cmdable, tableName string) *BaseRepository[T] {
	var entity T
	return &BaseRepository[T]{
		pgxPool:     pgxPool,
		redisClient: redisClient,
		tableName:   tableName,
		entityType:  reflect.TypeOf(entity),
	}
}

// Create inserts a new entity into the database
func (r *BaseRepository[T]) Create(ctx context.Context, entity T) error {
	// This is a simplified version - in practice, you'd use reflection or code generation
	// to build the SQL dynamically based on the entity struct
	query := fmt.Sprintf("INSERT INTO %s (id, created_at, updated_at) VALUES ($1, $2, $3)", r.tableName)

	// For now, we'll use a generic approach - specific repositories should override this
	_, err := r.pgxPool.Exec(ctx, query)
	return err
}

// GetByID retrieves an entity by its ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND deleted_at IS NULL", r.tableName)

	var entity T
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(&entity)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, err
	}

	return &entity, nil
}

// GetAll retrieves all entities with pagination
func (r *BaseRepository[T]) GetAll(ctx context.Context, limit, offset int) ([]*T, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2", r.tableName)

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*T
	for rows.Next() {
		var entity T
		err := rows.Scan(&entity)
		if err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// Update updates an entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity T) error {
	query := fmt.Sprintf("UPDATE %s SET updated_at = $1 WHERE id = $2", r.tableName)

	_, err := r.pgxPool.Exec(ctx, query)
	return err
}

// Delete soft deletes an entity
func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id = $1", r.tableName)

	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

// Count returns the total number of entities
func (r *BaseRepository[T]) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", r.tableName)

	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// Exists checks if an entity exists by ID
func (r *BaseRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND deleted_at IS NULL)", r.tableName)

	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(&exists)
	return exists, err
}

// Search performs a text search on entities
func (r *BaseRepository[T]) Search(ctx context.Context, query string, limit, offset int) ([]*T, error) {
	searchQuery := fmt.Sprintf("SELECT * FROM %s WHERE deleted_at IS NULL AND (name ILIKE $1 OR description ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3", r.tableName)

	searchTerm := "%" + strings.ToLower(query) + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*T
	for rows.Next() {
		var entity T
		err := rows.Scan(&entity)
		if err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

// GetRedisClient returns the Redis client
func (r *BaseRepository[T]) GetRedisClient() redis.Cmdable {
	return r.redisClient
}

// GetPgxPool returns the PostgreSQL connection pool
func (r *BaseRepository[T]) GetPgxPool() *pgxpool.Pool {
	return r.pgxPool
}
