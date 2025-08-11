package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// RoleRepository defines the interface for managing role entities
// This repository handles CRUD operations for roles, including activation/deactivation
// and search functionality with support for soft deletes.
type RoleRepository interface {
	// Create adds a new role to the system
	Create(ctx context.Context, role *entities.Role) error

	// GetByID retrieves a specific role by its UUID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error)

	// GetBySlug retrieves a role by its unique slug identifier
	GetBySlug(ctx context.Context, slug string) (*entities.Role, error)

	// GetAll retrieves all non-deleted roles with pagination support
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error)

	// GetActive retrieves only active, non-deleted roles with pagination
	GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error)

	// Search performs full-text search on role names, slugs, and descriptions
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error)

	// Update modifies an existing role's information
	Update(ctx context.Context, role *entities.Role) error

	// Delete performs a soft delete by setting deleted_at timestamp
	Delete(ctx context.Context, id uuid.UUID) error

	// Activate sets a role's is_active flag to true
	Activate(ctx context.Context, id uuid.UUID) error

	// Deactivate sets a role's is_active flag to false
	Deactivate(ctx context.Context, id uuid.UUID) error

	// ExistsBySlug checks if a role with the given slug exists and is not deleted
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// Count returns the total number of non-deleted roles
	Count(ctx context.Context) (int64, error)

	// CountActive returns the total number of active, non-deleted roles
	CountActive(ctx context.Context) (int64, error)
}

// RoleRepositoryImpl implements the RoleRepository interface
// This struct provides concrete implementations for role management operations
// using PostgreSQL for persistence and Redis for caching (if needed).
type RoleRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewRoleRepository creates a new instance of RoleRepositoryImpl
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - RoleRepository: interface implementation for role management
func NewRoleRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) RoleRepository {
	return &RoleRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// Create adds a new role to the roles table
// This method inserts a new role record with all required fields including
// generated timestamps and user tracking information.
//
// Parameters:
//   - ctx: context for the database operation
//   - role: pointer to the role entity to create
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Create(ctx context.Context, role *entities.Role) error {
	// Insert new role with generated timestamps
	query := `INSERT INTO roles (id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), $6, $7)`

	_, err := r.pgxPool.Exec(ctx, query,
		role.ID, role.Name, role.Slug, role.Description, role.IsActive, role.CreatedBy, role.UpdatedBy)
	return err
}

// GetByID retrieves a specific role by its UUID from the database
// This method performs a soft-delete aware query, only returning roles that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the role to retrieve
//
// Returns:
//   - *entities.Role: pointer to the found role entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	// Query for role by ID, excluding soft-deleted roles
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE id = $1 AND deleted_at IS NULL`

	var role entities.Role
	var createdBy, updatedBy *string

	err := r.pgxPool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetBySlug retrieves a role by its unique slug identifier from the database
// This method performs a soft-delete aware query, only returning roles that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - slug: string slug identifier of the role to retrieve
//
// Returns:
//   - *entities.Role: pointer to the found role entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	// Query for role by slug, excluding soft-deleted roles
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE slug = $1 AND deleted_at IS NULL`

	var role entities.Role
	var createdBy, updatedBy *string

	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetAll retrieves all non-deleted roles from the database with pagination support
// This method returns roles ordered by creation date (newest first) and supports
// limit/offset for efficient pagination.
//
// Parameters:
//   - ctx: context for the database operation
//   - limit: maximum number of roles to return
//   - offset: number of roles to skip (for pagination)
//
// Returns:
//   - []*entities.Role: slice of role entities with pagination
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	// Query for all non-deleted roles with pagination, ordered by creation date descending
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build role entities
	var roles []*entities.Role
	for rows.Next() {
		role, err := r.scanRoleRow(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// scanRoleRow is a helper function to scan a role row from database result set
// This method extracts role data from a database row and constructs a Role entity.
// It's used by methods that return multiple roles to avoid code duplication.
//
// Parameters:
//   - rows: pgx.Rows containing the database result set
//
// Returns:
//   - *entities.Role: pointer to the scanned role entity
//   - error: nil if successful, or error if scanning fails
func (r *RoleRepositoryImpl) scanRoleRow(rows pgx.Rows) (*entities.Role, error) {
	var role entities.Role
	var createdBy, updatedBy *string

	err := rows.Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetActive retrieves only active, non-deleted roles from the database with pagination
// This method filters for roles where is_active = true and orders by creation date.
//
// Parameters:
//   - ctx: context for the database operation
//   - limit: maximum number of active roles to return
//   - offset: number of active roles to skip (for pagination)
//
// Returns:
//   - []*entities.Role: slice of active role entities with pagination
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	// Query for active, non-deleted roles with pagination
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE is_active = true AND deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build role entities
	var roles []*entities.Role
	for rows.Next() {
		role, err := r.scanRoleRow(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Search performs full-text search on role names, slugs, and descriptions
// This method uses ILIKE pattern matching for case-insensitive search with wildcards.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search term to look for in role fields
//   - limit: maximum number of results to return
//   - offset: number of results to skip (for pagination)
//
// Returns:
//   - []*entities.Role: slice of role entities matching the search query
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	// Search query using ILIKE for case-insensitive pattern matching
	// Search across name, slug, and description fields
	searchQuery := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
					FROM roles WHERE deleted_at IS NULL AND
					(name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1)
					ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	// Add wildcards for pattern matching
	searchTerm := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build role entities
	var roles []*entities.Role
	for rows.Next() {
		role, err := r.scanRoleRow(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Update modifies an existing role's information in the database
// This method updates the role's name, slug, and description fields.
// Only non-deleted roles can be updated.
//
// Parameters:
//   - ctx: context for the database operation
//   - role: pointer to the role entity with updated information
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Update(ctx context.Context, role *entities.Role) error {
	// Update role fields, excluding soft-deleted roles
	query := `UPDATE roles SET name = $1, slug = $2, description = $3, updated_at = NOW()
			  WHERE id = $4 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, role.Name, role.Slug, role.Description, role.ID)
	return err
}

// Delete performs a soft delete by setting the deleted_at timestamp
// This method doesn't physically remove the record but marks it as deleted
// for data integrity and potential recovery purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the role to soft delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete by setting deleted_at timestamp and updating updated_at
	query := `UPDATE roles SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

// Activate sets a role's is_active flag to true
// This method enables a previously deactivated role.
// Only non-deleted roles can be activated.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the role to activate
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Activate(ctx context.Context, id uuid.UUID) error {
	// Set is_active to true and update timestamp
	query := `UPDATE roles SET is_active = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

// Deactivate sets a role's is_active flag to false
// This method disables a previously active role.
// Only non-deleted roles can be deactivated.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the role to deactivate
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	// Set is_active to false and update timestamp
	query := `UPDATE roles SET is_active = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

// ExistsBySlug checks if a role with the given slug exists and is not deleted
// This method uses an EXISTS subquery for efficient checking without retrieving full data.
//
// Parameters:
//   - ctx: context for the database operation
//   - slug: string slug identifier to check for existence
//
// Returns:
//   - bool: true if the role exists, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	// Use EXISTS subquery for efficient slug existence checking
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE slug = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

// Count returns the total number of non-deleted roles in the database
// This method provides a quick count for pagination and reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of non-deleted roles
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) Count(ctx context.Context) (int64, error) {
	// Count all non-deleted roles
	query := `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountActive returns the total number of active, non-deleted roles in the database
// This method provides a quick count of enabled roles for reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of active, non-deleted roles
//   - error: nil if successful, or database error if the operation fails
func (r *RoleRepositoryImpl) CountActive(ctx context.Context) (int64, error) {
	// Count only active, non-deleted roles
	query := `SELECT COUNT(*) FROM roles WHERE is_active = true AND deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}
