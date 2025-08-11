package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// MenuEntitiesRepository defines the interface for managing menu-role relationships
// This repository handles the many-to-many relationship between menus and roles,
// allowing for role-based access control to different menu items in the system.
type MenuEntitiesRepository interface {
	// AssignRoleToMenu assigns a specific role to a menu, enabling access control
	AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error

	// RemoveRoleFromMenu removes a role assignment from a menu, revoking access
	RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error

	// GetMenuRoles retrieves all roles that have access to a specific menu
	GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error)

	// GetRoleMenus retrieves all menus accessible by a specific role
	GetRoleMenus(ctx context.Context, roleID uuid.UUID) ([]*entities.Menu, error)

	// HasRole checks if a specific role has access to a specific menu
	HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error)

	// GetMenuRoleIDs retrieves the IDs of all roles that have access to a specific menu
	GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error)

	// CountMenusByRole counts the total number of menus accessible by a specific role
	CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
}

// MenuRoleRepositoryImpl implements the MenuEntitiesRepository interface
// This struct provides concrete implementations for managing menu-role relationships
// using PostgreSQL for persistence and Redis for caching (if needed).
type MenuRoleRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewMenuRoleRepository creates a new instance of MenuRoleRepositoryImpl
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - MenuEntitiesRepository: interface implementation for menu-role management
func NewMenuRoleRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MenuEntitiesRepository {
	return &MenuRoleRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// AssignRoleToMenu assigns a specific role to a menu by inserting a record into the menu_roles table
// This method uses an INSERT ... ON CONFLICT DO NOTHING pattern to handle duplicate assignments gracefully.
//
// Parameters:
//   - ctx: context for the database operation
//   - menuID: UUID of the menu to assign the role to
//   - roleID: UUID of the role to assign to the menu
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	// Insert role-menu relationship with conflict handling
	// ON CONFLICT DO NOTHING prevents duplicate assignments
	query := `INSERT INTO menu_roles (id, menu_id, role_id, created_at, updated_at)
			  VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
			  ON CONFLICT (menu_id, role_id) DO NOTHING`

	_, err := r.pgxPool.Exec(ctx, query, menuID, roleID)
	return err
}

// RemoveRoleFromMenu removes a role assignment from a menu by deleting the corresponding record
// This method permanently removes the role-menu relationship.
//
// Parameters:
//   - ctx: context for the database operation
//   - menuID: UUID of the menu to remove the role from
//   - roleID: UUID of the role to remove from the menu
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	// Delete the specific role-menu relationship
	query := `DELETE FROM menu_roles WHERE menu_id = $1 AND role_id = $2`

	_, err := r.pgxPool.Exec(ctx, query, menuID, roleID)
	return err
}

// GetMenuRoles retrieves all roles that have access to a specific menu
// This method joins the roles and menu_roles tables to get complete role information
// for roles that are active and not deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - menuID: UUID of the menu to get roles for
//
// Returns:
//   - []*entities.Role: slice of role entities with access to the menu
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error) {
	// Join roles and menu_roles tables to get roles with menu access
	// Filter for active, non-deleted roles and order by creation date
	query := `SELECT r.id, r.name, r.slug, r.description, r.is_active, r.created_at, r.updated_at, r.created_by, r.updated_by
			  FROM roles r
			  INNER JOIN menu_roles mr ON r.id = mr.role_id
			  WHERE mr.menu_id = $1 AND r.deleted_at IS NULL AND r.is_active = true
			  ORDER BY r.created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build role entities
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		var createdBy, updatedBy string

		// Scan row data into role struct and string variables for created_by/updated_by
		err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
			&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// GetRoleMenus retrieves all menus accessible by a specific role
// This method joins the menus and menu_roles tables to get complete menu information
// for menus that are not deleted, ordered by record ordering and creation date.
//
// Parameters:
//   - ctx: context for the database operation
//   - roleID: UUID of the role to get menus for
//
// Returns:
//   - []*entities.Menu: slice of menu entities accessible by the role
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) GetRoleMenus(ctx context.Context, roleID uuid.UUID) ([]*entities.Menu, error) {
	// Join menus and menu_roles tables to get menus accessible by the role
	// Filter for non-deleted menus and order by record ordering then creation date
	query := `SELECT m.id, m.name, m.slug, m.description, m.url, m.icon, m.parent_id, m.record_ordering, m.is_active, m.is_visible, m.target, m.created_at, m.updated_at, m.deleted_at
			  FROM menus m
			  INNER JOIN menu_roles mr ON m.id = mr.menu_id
			  WHERE mr.role_id = $1 AND m.deleted_at IS NULL
			  ORDER BY m.record_ordering ASC, m.created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build menu entities
	var menus []*entities.Menu
	for rows.Next() {
		var menu entities.Menu
		var parentIDStr *string

		// Scan row data into menu struct and handle nullable parent_id
		err := rows.Scan(&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr,
			&menu.RecordOrdering, &menu.IsActive, &menu.IsVisible, &menu.Target, &menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt)
		if err != nil {
			return nil, err
		}

		// Convert parent ID string to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				menu.ParentID = &parentID
			}
		}

		menus = append(menus, &menu)
	}

	return menus, nil
}

// HasRole checks if a specific role has access to a specific menu
// This method uses an EXISTS subquery for efficient checking without retrieving full data.
//
// Parameters:
//   - ctx: context for the database operation
//   - menuID: UUID of the menu to check access for
//   - roleID: UUID of the role to check access for
//
// Returns:
//   - bool: true if the role has access to the menu, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	// Use EXISTS subquery for efficient role access checking
	query := `SELECT EXISTS(SELECT 1 FROM menu_roles WHERE menu_id = $1 AND role_id = $2)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, menuID, roleID).Scan(&exists)
	return exists, err
}

// GetMenuRoleIDs retrieves the IDs of all roles that have access to a specific menu
// This method returns only the role IDs without full role information for efficiency.
//
// Parameters:
//   - ctx: context for the database operation
//   - menuID: UUID of the menu to get role IDs for
//
// Returns:
//   - []uuid.UUID: slice of role UUIDs with access to the menu
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	// Select only role IDs for efficiency when full role data isn't needed
	query := `SELECT role_id FROM menu_roles WHERE menu_id = $1`
	rows, err := r.pgxPool.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Build slice of role UUIDs
	var roleIDs []uuid.UUID
	for rows.Next() {
		var roleID uuid.UUID
		err := rows.Scan(&roleID)
		if err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, nil
}

// CountMenusByRole counts the total number of menus accessible by a specific role
// This method provides a quick count without retrieving full menu data.
//
// Parameters:
//   - ctx: context for the database operation
//   - roleID: UUID of the role to count menus for
//
// Returns:
//   - int64: total count of menus accessible by the role
//   - error: nil if successful, or database error if the operation fails
func (r *MenuRoleRepositoryImpl) CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	// Use COUNT(*) for efficient counting without retrieving data
	query := `SELECT COUNT(*) FROM menu_roles WHERE role_id = $1`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, roleID).Scan(&count)
	return count, err
}
