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

// MenuEntitiesRepository implements the MenuEntitiesRepository interface using PostgreSQL as the primary
// data store and Redis for caching. This repository manages the many-to-many relationship between
// menus and roles, providing functionality for role-based menu access control. It embeds the
// BaseTransactionalRepository to inherit transaction management capabilities.
type MenuEntitiesRepository struct {
	// BaseTransactionalRepository provides transaction management functionality
	*BaseTransactionalRepository
	// db holds the PostgreSQL connection pool for database operations
	db *pgxpool.Pool
	// redisClient holds the Redis client for caching operations
	redisClient redis.Cmdable
}

// NewPostgresMenuRoleRepository creates a new PostgreSQL menu role repository instance.
// This factory function initializes the repository with database and Redis connections,
// and sets up the base transactional repository for transaction management.
// Note: The function name suggests it creates a menu role repository, but it returns
// a MenuEntitiesRepository interface implementation.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations
//   - redisClient: The Redis client for caching operations
//
// Returns:
//   - repositories.MenuEntitiesRepository: A new menu entities repository instance
func NewPostgresMenuRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MenuEntitiesRepository {
	return &MenuEntitiesRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

// AssignRoleToMenu creates a role-menu association, granting the specified role
// access to the specified menu. This method uses an INSERT with ON CONFLICT DO NOTHING
// to prevent duplicate assignments and automatically generates UUIDs for new records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - menuID: The unique identifier of the menu to assign the role to
//   - roleID: The unique identifier of the role to assign to the menu
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	query := `INSERT INTO menu_roles (id, menu_id, role_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		ON CONFLICT (menu_id, role_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, menuID, roleID)
	return err
}

// RemoveRoleFromMenu removes a role-menu association, revoking the specified role's
// access to the specified menu. This method performs a direct DELETE operation
// on the junction table.
//
// Parameters:
//   - ctx: Context for the database operation
//   - menuID: The unique identifier of the menu to remove the role from
//   - roleID: The unique identifier of the role to remove from the menu
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	query := `DELETE FROM menu_roles WHERE menu_id = $1 AND role_id = $2`
	_, err := r.db.Exec(ctx, query, menuID, roleID)
	return err
}

// GetMenuRoles retrieves all roles that have access to a specific menu.
// This method performs a JOIN between the menu_roles junction table and the roles table
// to return complete role entities. It filters for active roles only and excludes
// soft-deleted roles. Results are ordered by role creation date.
//
// Parameters:
//   - ctx: Context for the database operation
//   - menuID: The unique identifier of the menu to get roles for
//
// Returns:
//   - []*entities.Role: List of role entities that have access to the menu
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error) {
	query := `SELECT r.id, r.name, r.slug, r.description, r.is_active, r.created_at, r.updated_at, r.created_by, r.updated_by
		FROM roles r
		INNER JOIN menu_roles mr ON r.id = mr.role_id
		WHERE mr.menu_id = $1 AND r.deleted_at IS NULL AND r.is_active = true
		ORDER BY r.created_at ASC`
	rows, err := r.db.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		var createdBy, updatedBy string
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// GetRoleMenus retrieves all menus that a specific role has access to.
// This method performs a JOIN between the menu_roles junction table and the menus table
// to return complete menu entities. It supports pagination and excludes soft-deleted menus.
// Results are ordered by menu ordering and creation date for consistent display.
//
// Parameters:
//   - ctx: Context for the database operation
//   - roleID: The unique identifier of the role to get menus for
//   - limit: Maximum number of menus to return
//   - offset: Number of menus to skip for pagination
//
// Returns:
//   - []*entities.Menu: List of menu entities that the role has access to
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) GetRoleMenus(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.Menu, error) {
	query := `SELECT m.id, m.name, m.slug, m.description, m.url, m.icon, m.parent_id, m.record_ordering, m.is_active, m.is_visible, m.target, m.created_at, m.updated_at, m.deleted_at
		FROM menus m
		INNER JOIN menu_roles mr ON m.id = mr.menu_id
		WHERE mr.role_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.record_ordering ASC, m.created_at ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, roleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var menu entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr, &menu.RecordOrdering, &menu.IsActive, &menu.IsVisible, &menu.Target, &menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt); err != nil {
			return nil, err
		}
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				menu.ParentID = &parentID
			}
		}
		menus = append(menus, &menu)
	}
	return menus, nil
}

// HasRole checks if a specific role has access to a specific menu.
// This method performs a simple existence check in the menu_roles junction table
// and is useful for permission validation in the application layer.
//
// Parameters:
//   - ctx: Context for the database operation
//   - menuID: The unique identifier of the menu to check
//   - roleID: The unique identifier of the role to check
//
// Returns:
//   - bool: True if the role has access to the menu, false otherwise
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menu_roles WHERE menu_id = $1 AND role_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, menuID, roleID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// HasAnyRole checks if a menu has access through any of the specified roles.
// This method iterates through the provided role IDs and returns true if any
// of them have access to the menu. It's useful for scenarios where a user
// has multiple roles and access should be granted if any role provides it.
//
// Parameters:
//   - ctx: Context for the database operation
//   - menuID: The unique identifier of the menu to check
//   - roleIDs: Slice of role IDs to check for access
//
// Returns:
//   - bool: True if any of the roles have access to the menu, false otherwise
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) HasAnyRole(ctx context.Context, menuID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	for _, roleID := range roleIDs {
		exists, err := r.HasRole(ctx, menuID, roleID)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

// GetMenuRoleIDs retrieves all role IDs that have access to a specific menu.
// This method returns just the role identifiers without the full role entities,
// which is useful for efficient permission checking and bulk operations.
//
// Parameters:
//   - ctx: Context for the database operation
//   - menuID: The unique identifier of the menu to get role IDs for
//
// Returns:
//   - []uuid.UUID: List of role IDs that have access to the menu
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT role_id FROM menu_roles WHERE menu_id = $1`
	rows, err := r.db.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roleIDs []uuid.UUID
	for rows.Next() {
		var roleID uuid.UUID
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}
	return roleIDs, nil
}

// CountMenusByRole returns the total number of menus that a specific role has access to.
// This method is useful for understanding role permissions and for pagination
// calculations when displaying role-specific menu lists.
//
// Parameters:
//   - ctx: Context for the database operation
//   - roleID: The unique identifier of the role to count menus for
//
// Returns:
//   - int64: The total count of menus accessible by the role
//   - error: Any error that occurred during the database operation
func (r *MenuEntitiesRepository) CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM menu_roles WHERE role_id = $1`
	var count int64
	if err := r.db.QueryRow(ctx, query, roleID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
