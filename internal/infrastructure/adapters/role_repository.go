// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresRoleRepository implements the RoleRepository interface using PostgreSQL as the primary
// data store and Redis for caching. This repository provides CRUD operations for role entities
// with support for soft deletes, role activation/deactivation, search functionality, and pagination.
// It embeds the BaseTransactionalRepository to inherit transaction management capabilities.
type PostgresRoleRepository struct {
	// BaseTransactionalRepository provides transaction management functionality
	*BaseTransactionalRepository
	// db holds the PostgreSQL connection pool for database operations
	db *pgxpool.Pool
	// redisClient holds the Redis client for caching operations
	redisClient redis.Cmdable
}

// NewPostgresRoleRepository creates a new PostgreSQL role repository instance.
// This factory function initializes the repository with database and Redis connections,
// and sets up the base transactional repository for transaction management.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations
//   - redisClient: The Redis client for caching operations
//
// Returns:
//   - repositories.RoleRepository: A new role repository instance
func NewPostgresRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.RoleRepository {
	return &PostgresRoleRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

// Create persists a new role entity to the database.
// This method inserts a new role record with all required fields including
// audit information (created_by, updated_by, timestamps) and sets the
// creation and update timestamps to the current time.
//
// Parameters:
//   - ctx: Context for the database operation
//   - role: The role entity to create
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	query := `INSERT INTO roles (id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,NOW(),NOW(),$6,$7)`
	_, err := r.db.Exec(ctx, query, role.ID, role.Name, role.Slug, role.Description, role.IsActive, role.CreatedBy, role.UpdatedBy)
	return err
}

// GetByID retrieves a role entity by its unique identifier.
// This method performs a soft-delete aware query, excluding records that have been
// marked as deleted. It returns the complete role entity with all fields populated.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the role to retrieve
//
// Returns:
//   - *entities.Role: The found role entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE id = $1 AND deleted_at IS NULL`
	var role entities.Role
	if err := r.db.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
		return nil, err
	}
	return &role, nil
}

// GetBySlug retrieves a role entity by its slug identifier.
// This method is useful for URL-friendly role lookups and performs a soft-delete
// aware query. The slug is typically a URL-safe version of the role name.
//
// Parameters:
//   - ctx: Context for the database operation
//   - slug: The slug identifier of the role to retrieve
//
// Returns:
//   - *entities.Role: The found role entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE slug = $1 AND deleted_at IS NULL`
	var role entities.Role
	if err := r.db.QueryRow(ctx, query, slug).Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
		return nil, err
	}
	return &role, nil
}

// GetAll retrieves a paginated list of all active roles.
// This method returns roles ordered by creation date (newest first) and supports
// pagination through limit and offset parameters. It excludes soft-deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of roles to return
//   - offset: Number of roles to skip for pagination
//
// Returns:
//   - []*entities.Role: List of role entities
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// GetActive retrieves a paginated list of only active roles.
// This method filters roles by their active status and is useful for scenarios
// where only enabled roles should be available for assignment or display.
// Results are ordered by creation date (newest first) and support pagination.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of active roles to return
//   - offset: Number of active roles to skip for pagination
//
// Returns:
//   - []*entities.Role: List of active role entities
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE is_active = true AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// Search performs a case-insensitive search for roles by name, slug, or description.
// This method uses ILIKE for pattern matching and supports partial matches across
// multiple fields. Results are ordered by creation date (newest first) and support
// pagination. The search is soft-delete aware, excluding deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - query: The search term to match against role names, slugs, and descriptions
//   - limit: Maximum number of matching roles to return
//   - offset: Number of matching roles to skip for pagination
//
// Returns:
//   - []*entities.Role: List of matching role entities
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	q := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE deleted_at IS NULL AND (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// Update modifies an existing role entity in the database.
// This method updates the role's name, slug, and description while preserving
// the original ID and creation information. It only updates non-deleted roles
// and automatically sets the updated_at timestamp to the current time.
//
// Parameters:
//   - ctx: Context for the database operation
//   - role: The role entity with updated values
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	query := `UPDATE roles SET name=$1, slug=$2, description=$3, updated_at=NOW() WHERE id=$4 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, role.Name, role.Slug, role.Description, role.ID)
	return err
}

// Delete performs a soft delete of a role entity.
// This method marks the role as deleted by setting the deleted_at timestamp
// and updates the updated_at timestamp. This preserves data integrity
// and allows for potential recovery while maintaining audit trails.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the role to delete
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Activate enables a role by setting its is_active flag to true.
// This method is useful for re-enabling previously deactivated roles.
// It only affects non-deleted roles and updates the updated_at timestamp.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the role to activate
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Activate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET is_active = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Deactivate disables a role by setting its is_active flag to false.
// This method is useful for temporarily disabling roles without deleting them.
// It only affects non-deleted roles and updates the updated_at timestamp.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the role to deactivate
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET is_active = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ExistsBySlug checks if a role with the specified slug already exists.
// This method is useful for validation purposes, ensuring role slugs are unique
// before creation. It performs a soft-delete aware check.
//
// Parameters:
//   - ctx: Context for the database operation
//   - slug: The slug to check for existence
//
// Returns:
//   - bool: True if a role with the slug exists, false otherwise
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// Count returns the total number of active roles in the system.
// This method excludes soft-deleted records and is useful for pagination
// calculations and system statistics.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of active roles
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountActive returns the total number of active (enabled) roles in the system.
// This method excludes soft-deleted records and only counts roles where is_active is true.
// It's useful for system statistics and monitoring active role usage.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of active roles
//   - error: Any error that occurred during the database operation
func (r *PostgresRoleRepository) CountActive(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM roles WHERE is_active = true AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
