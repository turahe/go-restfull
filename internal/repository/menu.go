package repository

import (
	"context"
	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *model.Menu) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Menu, error)
	GetBySlug(ctx context.Context, slug string) (*model.Menu, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Menu, error)
	GetActive(ctx context.Context, limit, offset int) ([]*model.Menu, error)
	GetVisible(ctx context.Context, limit, offset int) ([]*model.Menu, error)
	GetRootMenus(ctx context.Context) ([]*model.Menu, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Menu, error)
	GetHierarchy(ctx context.Context) ([]*model.Menu, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*model.Menu, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*model.Menu, error)
	Update(ctx context.Context, menu *model.Menu) error
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	Show(ctx context.Context, id uuid.UUID) error
	Hide(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	CountVisible(ctx context.Context) (int64, error)
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

func (r *MenuRepositoryImpl) Create(ctx context.Context, menu *model.Menu) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// If this is a root menu (no parent)
	if menu.ParentID == nil {
		// Get the maximum right value and add 1 for the new node
		var maxRight int64
		err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(record_right), 0) FROM menus WHERE deleted_at IS NULL`).Scan(&maxRight)
		if err != nil {
			return err
		}

		menu.RecordLeft = maxRight + 1
		menu.RecordRight = maxRight + 2
	} else {
		// Get the parent's right value
		var parentRight int64
		err = tx.QueryRow(ctx, `SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL`, *menu.ParentID).Scan(&parentRight)
		if err != nil {
			return err
		}

		// Make space for the new node by shifting all nodes to the right
		_, err = tx.Exec(ctx, `
			UPDATE menus 
			SET record_left = CASE 
				WHEN record_left > $1 THEN record_left + 2 
				ELSE record_left 
			END,
			record_right = CASE 
				WHEN record_right >= $1 THEN record_right + 2 
				ELSE record_right 
			END
			WHERE deleted_at IS NULL
		`, parentRight)
		if err != nil {
			return err
		}

		menu.RecordLeft = parentRight
		menu.RecordRight = parentRight + 1
	}

	// Insert the new menu
	query := `
		INSERT INTO menus (id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err = tx.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon,
		menu.ParentID, menu.RecordLeft, menu.RecordRight, menu.RecordOrdering,
		menu.IsActive, menu.IsVisible, menu.Target,
		menu.CreatedAt, menu.UpdatedAt, menu.CreatedBy, menu.UpdatedBy,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *MenuRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus WHERE id = $1 AND deleted_at IS NULL
	`

	var menu model.Menu
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
		&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
		&menu.IsActive, &menu.IsVisible, &menu.Target,
		&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
	)

	if err != nil {
		return nil, err
	}

	return &menu, nil
}

func (r *MenuRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus WHERE slug = $1 AND deleted_at IS NULL
	`

	var menu model.Menu
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
		&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
		&menu.IsActive, &menu.IsVisible, &menu.Target,
		&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
	)

	if err != nil {
		return nil, err
	}

	return &menu, nil
}

func (r *MenuRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetActive(ctx context.Context, limit, offset int) ([]*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus WHERE is_active = true AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetVisible(ctx context.Context, limit, offset int) ([]*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus WHERE is_active = true AND is_visible = true AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetRootMenus(ctx context.Context) ([]*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus WHERE parent_id IS NULL AND is_active = true AND is_visible = true AND deleted_at IS NULL
		ORDER BY record_ordering ASC, name ASC
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus 
		WHERE parent_id = $1 AND is_active = true AND is_visible = true AND deleted_at IS NULL
		ORDER BY record_ordering ASC, name ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, parentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetHierarchy(ctx context.Context) ([]*model.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus 
		WHERE is_active = true AND is_visible = true AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*model.Menu, error) {
	query := `
		SELECT DISTINCT m.id, m.name, m.slug, m.description, m.url, m.icon, m.parent_id, m.record_left, m.record_right, m.record_ordering, m.is_active, m.is_visible, m.target, m.created_at, m.updated_at, m.deleted_at, m.created_by, m.updated_by, m.deleted_by
		FROM menus m
		INNER JOIN menu_roles mr ON m.id = mr.menu_id
		INNER JOIN user_roles ur ON mr.role_id = ur.role_id
		WHERE ur.user_id = $1 AND m.is_active = true AND m.is_visible = true AND m.deleted_at IS NULL
		ORDER BY m.record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*model.Menu, error) {
	searchQuery := `
		SELECT id, name, slug, description, url, icon, parent_id, record_left, record_right, record_ordering, is_active, is_visible, target, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM menus 
		WHERE (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1) AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*model.Menu
	for rows.Next() {
		var menu model.Menu
		err := rows.Scan(
			&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon,
			&menu.ParentID, &menu.RecordLeft, &menu.RecordRight, &menu.RecordOrdering,
			&menu.IsActive, &menu.IsVisible, &menu.Target,
			&menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt, &menu.CreatedBy, &menu.UpdatedBy, &menu.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *MenuRepositoryImpl) Update(ctx context.Context, menu *model.Menu) error {
	// For nested set, we need to handle parent changes carefully
	// This is a simplified update that doesn't change the tree structure
	query := `
		UPDATE menus 
		SET name = $2, slug = $3, description = $4, url = $5, icon = $6, record_ordering = $7, is_active = $8, is_visible = $9, target = $10, updated_at = $11, updated_by = $12
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pgxPool.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon,
		menu.RecordOrdering, menu.IsActive, menu.IsVisible, menu.Target,
		menu.UpdatedAt, menu.UpdatedBy,
	)

	return err
}

func (r *MenuRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the node's left and right values
	var left, right int64
	err = tx.QueryRow(ctx, `SELECT record_left, record_right FROM menus WHERE id = $1 AND deleted_at IS NULL`, id.String()).Scan(&left, &right)
	if err != nil {
		return err
	}

	// Calculate the width of the subtree
	width := right - left + 1

	// Delete the node and all its descendants
	_, err = tx.Exec(ctx, `UPDATE menus SET deleted_at = NOW() WHERE record_left >= $1 AND record_right <= $2 AND deleted_at IS NULL`, left, right)
	if err != nil {
		return err
	}

	// Close the gap by shifting all nodes to the left
	_, err = tx.Exec(ctx, `
		UPDATE menus 
		SET record_left = CASE 
			WHEN record_left > $1 THEN record_left - $2 
			ELSE record_left 
		END,
		record_right = CASE 
			WHEN record_right > $1 THEN record_right - $2 
			ELSE record_right 
		END
		WHERE deleted_at IS NULL
	`, right, width)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *MenuRepositoryImpl) Activate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_active = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *MenuRepositoryImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_active = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *MenuRepositoryImpl) Show(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_visible = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *MenuRepositoryImpl) Hide(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_visible = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
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

func (r *MenuRepositoryImpl) CountActive(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE is_active = true AND deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *MenuRepositoryImpl) CountVisible(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE is_active = true AND is_visible = true AND deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}
