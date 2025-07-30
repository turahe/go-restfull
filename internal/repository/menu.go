package repository

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *entities.Menu) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error)
	GetRootMenus(ctx context.Context) ([]*entities.Menu, error)
	Update(ctx context.Context, menu *entities.Menu) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

type MenuRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewMenuRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MenuRepository {
	return &MenuRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *MenuRepositoryImpl) Create(ctx context.Context, menu *entities.Menu) error {
	query := `INSERT INTO menus (id, name, slug, description, url, icon, parent_id, record_ordering, is_active, is_visible, target, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	parentIDStr := ""
	if menu.ParentID != nil {
		parentIDStr = menu.ParentID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query,
		menu.ID.String(), menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordOrdering, menu.IsActive, menu.IsVisible, menu.Target, menu.CreatedAt, menu.UpdatedAt)
	return err
}

func (r *MenuRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	query := `SELECT id, name, slug, description, url, icon, parent_id, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at
			  FROM menus WHERE id = $1 AND deleted_at IS NULL`

	var menu entities.Menu
	var parentIDStr *string

	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr,
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

	return &menu, nil
}

func (r *MenuRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	query := `SELECT id, name, slug, description, url, icon, parent_id, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at
			  FROM menus WHERE deleted_at IS NULL
			  ORDER BY record_ordering ASC, created_at ASC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*entities.Menu
	for rows.Next() {
		menu, err := r.scanMenuRow(rows)
		if err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error) {
	query := `SELECT id, name, slug, description, url, icon, parent_id, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at
			  FROM menus WHERE parent_id = $1 AND deleted_at IS NULL
			  ORDER BY record_ordering ASC, created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, parentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*entities.Menu
	for rows.Next() {
		menu, err := r.scanMenuRow(rows)
		if err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	query := `SELECT id, name, slug, description, url, icon, parent_id, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at
			  FROM menus WHERE parent_id IS NULL AND deleted_at IS NULL
			  ORDER BY record_ordering ASC, created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*entities.Menu
	for rows.Next() {
		menu, err := r.scanMenuRow(rows)
		if err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) Update(ctx context.Context, menu *entities.Menu) error {
	query := `UPDATE menus SET name = $1, slug = $2, description = $3, url = $4, icon = $5, parent_id = $6, record_ordering = $7, is_active = $8, is_visible = $9, target = $10, updated_at = $11
			  WHERE id = $12 AND deleted_at IS NULL`

	parentIDStr := ""
	if menu.ParentID != nil {
		parentIDStr = menu.ParentID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordOrdering, menu.IsActive, menu.IsVisible, menu.Target, menu.UpdatedAt, menu.ID.String())
	return err
}

func (r *MenuRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *MenuRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menus WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

func (r *MenuRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// scanMenuRow is a helper function to scan a menu row from database
func (r *MenuRepositoryImpl) scanMenuRow(rows pgx.Rows) (*entities.Menu, error) {
	var menu entities.Menu
	var parentIDStr *string

	err := rows.Scan(
		&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr,
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

	return &menu, nil
}
