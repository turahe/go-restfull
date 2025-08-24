package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MenuEntitiesRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresMenuRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MenuEntitiesRepository {
	return &MenuEntitiesRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

func (r *MenuEntitiesRepository) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	query := `INSERT INTO menu_roles (id, menu_id, role_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		ON CONFLICT (menu_id, role_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, menuID, roleID)
	return err
}

func (r *MenuEntitiesRepository) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	query := `DELETE FROM menu_roles WHERE menu_id = $1 AND role_id = $2`
	_, err := r.db.Exec(ctx, query, menuID, roleID)
	return err
}

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
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				menu.ParentID = &parentID
			}
		}
		menus = append(menus, &menu)
	}
	return menus, nil
}

func (r *MenuEntitiesRepository) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menu_roles WHERE menu_id = $1 AND role_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, menuID, roleID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

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

func (r *MenuEntitiesRepository) CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM menu_roles WHERE role_id = $1`
	var count int64
	if err := r.db.QueryRow(ctx, query, roleID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
