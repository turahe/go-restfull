package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresMenuRepository struct {
	*BaseTransactionalRepository
	db        *pgxpool.Pool
	nestedSet *nestedset.NestedSetManager
}

func NewPostgresMenuRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MenuRepository {
	return &PostgresMenuRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		nestedSet:                   nestedset.NewNestedSetManager(db),
	}
}

func (r *PostgresMenuRepository) Create(ctx context.Context, menu *entities.Menu) error {
	// Calculate nested set values
	nestedSetValues, err := r.nestedSet.CreateNode(ctx, "menus", menu.ParentID, int64(1))
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign nested set values to menu entity
	menu.RecordLeft = &nestedSetValues.Left
	menu.RecordRight = &nestedSetValues.Right
	menu.RecordDepth = &nestedSetValues.Depth
	menu.RecordOrdering = &nestedSetValues.Ordering

	query := `
		INSERT INTO menus (
			id, name, slug, description, url, icon, parent_id,
			record_left, record_right, record_depth, record_ordering,
			is_active, is_visible, target, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18
		)`
	var parentIDStr *string
	if menu.ParentID != nil {
		parentIDStr = func() *string { s := menu.ParentID.String(); return &s }()
	}

	// Handle created_by and updated_by as strings for VARCHAR fields
	createdByStr := "system"
	updatedByStr := "system"
	if menu.CreatedBy != uuid.Nil {
		createdByStr = menu.CreatedBy.String()
	}
	if menu.UpdatedBy != uuid.Nil {
		updatedByStr = menu.UpdatedBy.String()
	}

	_, err = r.db.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordLeft, menu.RecordRight, menu.RecordDepth, menu.RecordOrdering,
		menu.IsActive, menu.IsVisible, menu.Target, createdByStr, updatedByStr, menu.CreatedAt, menu.UpdatedAt,
	)
	return err
}

// createMenuFallback creates a menu with manual nested set values
// This is used when the nested set manager fails to create the first menu
func (r *PostgresMenuRepository) createMenuFallback(ctx context.Context, menu *entities.Menu) error {
	// For the first menu, set manual nested set values
	var recordLeft, recordRight, recordDepth, recordOrdering int64
	if menu.ParentID == nil {
		// Root menu - start with basic values
		recordLeft = 1
		recordRight = 2
		recordDepth = 0
		recordOrdering = 1
	} else {
		// Child menu - this shouldn't happen in fallback, but handle it
		recordLeft = 3
		recordRight = 4
		recordDepth = 1
		recordOrdering = 1
	}

	menu.RecordLeft = &recordLeft
	menu.RecordRight = &recordRight
	menu.RecordDepth = &recordDepth
	menu.RecordOrdering = &recordOrdering

	query := `
		INSERT INTO menus (
			id, name, slug, description, url, icon, parent_id,
			record_left, record_right, record_depth, record_ordering,
			is_active, is_visible, target, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18
		)`
	var parentIDStr *string
	if menu.ParentID != nil {
		parentIDStr = func() *string { s := menu.ParentID.String(); return &s }()
	}

	// Handle created_by and updated_by as strings for VARCHAR fields
	createdByStr := "system"
	updatedByStr := "system"
	if menu.CreatedBy != uuid.Nil {
		createdByStr = menu.CreatedBy.String()
	}
	if menu.UpdatedBy != uuid.Nil {
		updatedByStr = menu.UpdatedBy.String()
	}

	_, err := r.db.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordLeft, menu.RecordRight, menu.RecordDepth, menu.RecordOrdering,
		menu.IsActive, menu.IsVisible, menu.Target, createdByStr, updatedByStr, menu.CreatedAt, menu.UpdatedAt,
	)
	return err
}

func (r *PostgresMenuRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE id = $1 AND deleted_at IS NULL`
	var m entities.Menu
	var parentIDStr *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr,
		&m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering,
		&m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	if parentIDStr != nil {
		if p, err := uuid.Parse(*parentIDStr); err == nil {
			m.ParentID = &p
		}
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetBySlug(ctx context.Context, slug string) (*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE slug = $1 AND deleted_at IS NULL`
	var m entities.Menu
	var parentIDStr *string
	if err := r.db.QueryRow(ctx, query, slug).Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
		return nil, err
	}
	if parentIDStr != nil {
		if p, err := uuid.Parse(*parentIDStr); err == nil {
			m.ParentID = &p
		}
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE deleted_at IS NULL ORDER BY record_left ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		var createdByStr, updatedByStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &createdByStr, &updatedByStr, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				m.ParentID = &p
			}
		}
		// Convert string fields to UUID
		if createdByStr != "" {
			if createdBy, err := uuid.Parse(createdByStr); err == nil {
				m.CreatedBy = createdBy
			}
		}
		if updatedByStr != "" {
			if updatedBy, err := uuid.Parse(updatedByStr); err == nil {
				m.UpdatedBy = updatedBy
			}
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE is_active = true AND deleted_at IS NULL ORDER BY record_left ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		var createdByStr, updatedByStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &createdByStr, &updatedByStr, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				m.ParentID = &p
			}
		}
		// Convert string fields to UUID
		if createdByStr != "" {
			if createdBy, err := uuid.Parse(createdByStr); err == nil {
				m.CreatedBy = createdBy
			}
		}
		if updatedByStr != "" {
			if updatedBy, err := uuid.Parse(updatedByStr); err == nil {
				m.UpdatedBy = updatedBy
			}
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetVisible(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE is_visible = true AND deleted_at IS NULL ORDER BY record_left ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		var createdByStr, updatedByStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &createdByStr, &updatedByStr, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				m.ParentID = &p
			}
		}
		// Convert string fields to UUID
		if createdByStr != "" {
			if createdBy, err := uuid.Parse(createdByStr); err == nil {
				m.CreatedBy = createdBy
			}
		}
		if updatedByStr != "" {
			if updatedBy, err := uuid.Parse(updatedByStr); err == nil {
				m.UpdatedBy = updatedBy
			}
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE parent_id IS NULL AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		var createdByStr, updatedByStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &createdByStr, &updatedByStr, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		// Convert string fields to UUID
		if createdByStr != "" {
			if createdBy, err := uuid.Parse(createdByStr); err == nil {
				m.CreatedBy = createdBy
			}
		}
		if updatedByStr != "" {
			if updatedBy, err := uuid.Parse(updatedByStr); err == nil {
				m.UpdatedBy = updatedBy
			}
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE parent_id = $1 AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		var createdByStr, updatedByStr string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &createdByStr, &updatedByStr, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				m.ParentID = &p
			}
		}
		// Convert string fields to UUID
		if createdByStr != "" {
			if createdBy, err := uuid.Parse(createdByStr); err == nil {
				m.CreatedBy = createdBy
			}
		}
		if updatedByStr != "" {
			if updatedBy, err := uuid.Parse(updatedByStr); err == nil {
				m.UpdatedBy = updatedBy
			}
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetHierarchy(ctx context.Context) ([]*entities.Menu, error) {
	// Build hierarchy using a single ordered query
	menus, err := r.GetAll(ctx, 10000, 0)
	if err != nil {
		return nil, err
	}
	menuMap := make(map[uuid.UUID]*entities.Menu)
	var rootMenus []*entities.Menu
	for _, menu := range menus {
		menuMap[menu.ID] = menu
		menu.Children = []*entities.Menu{}
	}
	for _, menu := range menus {
		if menu.ParentID == nil {
			rootMenus = append(rootMenus, menu)
		} else {
			if parent, exists := menuMap[*menu.ParentID]; exists {
				parent.Children = append(parent.Children, menu)
			}
		}
	}
	return rootMenus, nil
}

func (r *PostgresMenuRepository) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error) {
	// For now, return all visible menus
	return r.GetVisible(ctx, 1000, 0)
}

func (r *PostgresMenuRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	q := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE deleted_at IS NULL AND (
			name ILIKE $1 OR description ILIKE $1 OR slug ILIKE $1
		) ORDER BY record_left ASC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, "%"+strings.ToLower(query)+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				m.ParentID = &p
			}
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) Update(ctx context.Context, menu *entities.Menu) error {
	query := `
		UPDATE menus SET
			name = $2, slug = $3, description = $4, url = $5, icon = $6,
			is_active = $7, is_visible = $8, target = $9, updated_by = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon,
		menu.IsActive, menu.IsVisible, menu.Target, menu.UpdatedBy, menu.UpdatedAt,
	)
	return err
}

func (r *PostgresMenuRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresMenuRepository) Activate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_active = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresMenuRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_active = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresMenuRepository) Show(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_visible = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresMenuRepository) Hide(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE menus SET is_visible = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresMenuRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menus WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresMenuRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresMenuRepository) CountActive(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE is_active = true AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresMenuRepository) CountVisible(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE is_visible = true AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Nested Set Tree Traversal methods
func (r *PostgresMenuRepository) GetDescendants(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	ids, err := r.nestedSet.GetDescendants(ctx, "menus", menuID, 10000, 0)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Menu{}, nil
	}
	query := `SELECT id, name, slug, description, url, icon, parent_id,
	       record_left, record_right, record_depth, record_ordering,
	       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
	FROM menus WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetAncestors(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	ids, err := r.nestedSet.GetAncestors(ctx, "menus", menuID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Menu{}, nil
	}
	query := `SELECT id, name, slug, description, url, icon, parent_id,
	       record_left, record_right, record_depth, record_ordering,
	       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
	FROM menus WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetSiblings(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	ids, err := r.nestedSet.GetSiblings(ctx, "menus", menuID, 10000, 0)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Menu{}, nil
	}
	query := `SELECT id, name, slug, description, url, icon, parent_id,
	       record_left, record_right, record_depth, record_ordering,
	       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
	FROM menus WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetPath(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	ids, err := r.nestedSet.GetPath(ctx, "menus", menuID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Menu{}, nil
	}
	query := `SELECT id, name, slug, description, url, icon, parent_id,
	       record_left, record_right, record_depth, record_ordering,
	       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
	FROM menus WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetTree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	return r.GetSubtree(ctx, menuID)
}

func (r *PostgresMenuRepository) GetSubtree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus
		WHERE record_left >= (SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND record_right <= (SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND deleted_at IS NULL
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		var parentIDStr *string
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &parentIDStr, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

// Advanced Nested Set Operations
func (r *PostgresMenuRepository) AddChild(ctx context.Context, parentID, childID uuid.UUID) error {
	return r.nestedSet.MoveSubtree(ctx, "menus", childID, parentID)
}

func (r *PostgresMenuRepository) MoveSubtree(ctx context.Context, menuID, newParentID uuid.UUID) error {
	return r.nestedSet.MoveSubtree(ctx, "menus", menuID, newParentID)
}

func (r *PostgresMenuRepository) DeleteSubtree(ctx context.Context, menuID uuid.UUID) error {
	return r.nestedSet.DeleteSubtree(ctx, "menus", menuID)
}

func (r *PostgresMenuRepository) InsertBetween(ctx context.Context, menu *entities.Menu, leftSiblingID, rightSiblingID *uuid.UUID) error {
	// Not implemented in NestedSetManager; use Create for now
	return r.Create(ctx, menu)
}

func (r *PostgresMenuRepository) SwapPositions(ctx context.Context, menu1ID, menu2ID uuid.UUID) error {
	return fmt.Errorf("SwapPositions not implemented")
}

func (r *PostgresMenuRepository) GetLeafNodes(ctx context.Context) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus m
		WHERE deleted_at IS NULL AND NOT EXISTS (
			SELECT 1 FROM menus c WHERE c.parent_id = m.id AND c.deleted_at IS NULL
		)
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &m.ParentID, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetInternalNodes(ctx context.Context) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id,
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus m
		WHERE deleted_at IS NULL AND EXISTS (
			SELECT 1 FROM menus c WHERE c.parent_id = m.id AND c.deleted_at IS NULL
		)
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var menus []*entities.Menu
	for rows.Next() {
		var m entities.Menu
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.Description, &m.URL, &m.Icon, &m.ParentID, &m.RecordLeft, &m.RecordRight, &m.RecordDepth, &m.RecordOrdering, &m.IsActive, &m.IsVisible, &m.Target, &m.CreatedBy, &m.UpdatedBy, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

// Batch Operations
func (r *PostgresMenuRepository) BatchMoveSubtrees(ctx context.Context, moves []struct {
	MenuID      uuid.UUID
	NewParentID uuid.UUID
}) error {
	for _, mv := range moves {
		if err := r.nestedSet.MoveSubtree(ctx, "menus", mv.MenuID, mv.NewParentID); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresMenuRepository) BatchInsertBetween(ctx context.Context, insertions []struct {
	Menu           *entities.Menu
	LeftSiblingID  *uuid.UUID
	RightSiblingID *uuid.UUID
}) error {
	for _, ins := range insertions {
		if err := r.Create(ctx, ins.Menu); err != nil {
			return err
		}
	}
	return nil
}

// Tree Maintenance and Optimization (use helper as available)
func (r *PostgresMenuRepository) ValidateTree(ctx context.Context) ([]string, error) {
	return r.nestedSet.ValidateTree(ctx, "menus")
}

func (r *PostgresMenuRepository) RebuildTree(ctx context.Context) error {
	return r.nestedSet.RebuildTree(ctx, "menus")
}

func (r *PostgresMenuRepository) OptimizeTree(ctx context.Context) error {
	return r.nestedSet.RebuildTree(ctx, "menus")
}

func (r *PostgresMenuRepository) GetTreeStatistics(ctx context.Context) (map[string]interface{}, error) {
	return r.nestedSet.GetTreeStatistics(ctx, "menus")
}

func (r *PostgresMenuRepository) GetTreeHeight(ctx context.Context) (int64, error) {
	return r.nestedSet.GetTreeHeight(ctx, "menus")
}

func (r *PostgresMenuRepository) GetLevelWidth(ctx context.Context, level uint64) (int64, error) {
	return r.nestedSet.GetLevelWidth(ctx, "menus", int64(level))
}

func (r *PostgresMenuRepository) GetSubtreeSize(ctx context.Context, menuID uuid.UUID) (int64, error) {
	return r.nestedSet.GetSubtreeSize(ctx, "menus", menuID)
}

func (r *PostgresMenuRepository) GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	return r.nestedSet.GetTreeStatistics(ctx, "menus")
}

func (r *PostgresMenuRepository) ValidateTreeIntegrity(ctx context.Context) ([]string, error) {
	return r.nestedSet.ValidateTree(ctx, "menus")
}
