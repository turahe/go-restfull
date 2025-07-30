package repository

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MenuEntitiesRepository interface {
	AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error
	RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error
	GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error)
	GetRoleMenus(ctx context.Context, roleID uuid.UUID) ([]*entities.Menu, error)
	HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error)
	GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error)
	CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
}

type MenuRoleRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewMenuRoleRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MenuEntitiesRepository {
	return &MenuRoleRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *MenuRoleRepositoryImpl) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	query := `INSERT INTO menu_roles (id, menu_id, role_id, created_at, updated_at)
			  VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
			  ON CONFLICT (menu_id, role_id) DO NOTHING`

	_, err := r.pgxPool.Exec(ctx, query, menuID, roleID)
	return err
}

func (r *MenuRoleRepositoryImpl) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	query := `DELETE FROM menu_roles WHERE menu_id = $1 AND role_id = $2`

	_, err := r.pgxPool.Exec(ctx, query, menuID, roleID)
	return err
}

func (r *MenuRoleRepositoryImpl) GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error) {
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

	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		var createdBy, updatedBy string

		err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
			&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

func (r *MenuRoleRepositoryImpl) GetRoleMenus(ctx context.Context, roleID uuid.UUID) ([]*entities.Menu, error) {
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

	var menus []*entities.Menu
	for rows.Next() {
		var menu entities.Menu
		var parentIDStr *string

		err := rows.Scan(&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr,
			&menu.RecordOrdering, &menu.IsActive, &menu.IsVisible, &menu.Target, &menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt)
		if err != nil {
			return nil, err
		}

		// Convert parent ID string to UUID
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				menu.ParentID = &parentID
			}
		}

		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRoleRepositoryImpl) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menu_roles WHERE menu_id = $1 AND role_id = $2)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, menuID, roleID).Scan(&exists)
	return exists, err
}

func (r *MenuRoleRepositoryImpl) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT role_id FROM menu_roles WHERE menu_id = $1`
	rows, err := r.pgxPool.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r *MenuRoleRepositoryImpl) CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM menu_roles WHERE role_id = $1`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, roleID).Scan(&count)
	return count, err
}
