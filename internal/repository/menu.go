package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

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

type menuRepository struct {
	db          *pgxpool.Pool
	redisClient redis.Cmdable
	nestedSet   *nestedset.NestedSetManager
}

func NewMenuRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MenuRepository {
	return &menuRepository{
		db:          pgxPool,
		redisClient: redisClient,
		nestedSet:   nestedset.NewNestedSetManager(pgxPool),
	}
}

func (r *menuRepository) Create(ctx context.Context, menu *entities.Menu) error {
	// Calculate nested set values using the shared manager
	values, err := r.nestedSet.CreateNode(ctx, "menus", menu.ParentID, 1)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign computed nested set values to the entity
	menu.RecordLeft = &values.Left
	menu.RecordRight = &values.Right
	menu.RecordDepth = &values.Depth
	menu.RecordOrdering = &values.Ordering

	// Insert the new menu
	query := `
		INSERT INTO menus (
			id, name, slug, description, url, icon, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			is_active, is_visible, target, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	parentIDStr := ""
	if menu.ParentID != nil {
		parentIDStr = menu.ParentID.String()
	}

	_, err = r.db.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordLeft, menu.RecordRight, menu.RecordDepth, menu.RecordOrdering,
		menu.IsActive, menu.IsVisible, menu.Target, menu.CreatedBy, menu.UpdatedBy, menu.CreatedAt, menu.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create menu: %w", err)
	}

	return nil
}

func (r *menuRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE id = $1 AND deleted_at IS NULL
	`

	var menu entities.Menu
	var parentIDStr *string

	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr,
		&menu.RecordLeft, &menu.RecordRight, &menu.RecordDepth, &menu.RecordOrdering,
		&menu.IsActive, &menu.IsVisible, &menu.Target, &menu.CreatedBy, &menu.UpdatedBy, &menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt)
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

func (r *menuRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE deleted_at IS NULL
		ORDER BY record_left ASC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
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

func (r *menuRepository) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_ordering ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, parentID.String())
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

func (r *menuRepository) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY record_ordering ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
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

func (r *menuRepository) Update(ctx context.Context, menu *entities.Menu) error {
	// For updates, we only update basic fields, not the tree structure
	// Tree restructuring would require complex operations and is not implemented here
	query := `
		UPDATE menus SET name = $1, slug = $2, description = $3, url = $4, icon = $5, 
		                is_active = $6, is_visible = $7, target = $8, updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon,
		menu.IsActive, menu.IsVisible, menu.Target, menu.UpdatedAt, menu.ID.String())
	return err
}

func (r *menuRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	query := `UPDATE menus SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id.String())
	return err
}

func (r *menuRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menus WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

func (r *menuRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// Nested Set Tree Traversal Methods

func (r *menuRepository) GetDescendants(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE record_left > (SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND record_right < (SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, menuID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
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

func (r *menuRepository) GetAncestors(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE record_left < (SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND record_right > (SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, menuID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
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

func (r *menuRepository) GetSiblings(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE parent_id = (SELECT parent_id FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND id != $1
		  AND deleted_at IS NULL
		ORDER BY record_ordering ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, menuID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
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

func (r *menuRepository) GetPath(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE record_left <= (SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND record_right >= (SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, menuID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get path: %w", err)
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

func (r *menuRepository) GetTree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	// Get the complete tree starting from the specified menu
	return r.GetSubtree(ctx, menuID)
}

func (r *menuRepository) GetSubtree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE record_left >= (SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND record_right <= (SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, menuID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get subtree: %w", err)
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

// Advanced Nested Set Operations

func (r *menuRepository) AddChild(ctx context.Context, parentID, childID uuid.UUID) error {
	// Use the nested set manager to properly restructure the tree
	err := r.nestedSet.MoveSubtree(ctx, "menus", childID, parentID)
	if err != nil {
		return fmt.Errorf("failed to add child using nested set: %w", err)
	}

	// Update the updated_at timestamp
	query := `UPDATE menus SET updated_at = NOW() WHERE id = $1`
	_, err = r.db.Exec(ctx, query, childID)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}

	return nil
}

func (r *menuRepository) MoveSubtree(ctx context.Context, menuID, newParentID uuid.UUID) error {
	// Use the nested set manager to properly restructure the tree
	err := r.nestedSet.MoveSubtree(ctx, "menus", menuID, newParentID)
	if err != nil {
		return fmt.Errorf("failed to move subtree using nested set: %w", err)
	}

	// Update the updated_at timestamp
	query := `UPDATE menus SET updated_at = NOW() WHERE id = $1`
	_, err = r.db.Exec(ctx, query, menuID)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}

	return nil
}

func (r *menuRepository) DeleteSubtree(ctx context.Context, menuID uuid.UUID) error {
	// Use the nested set manager to properly handle subtree deletion
	err := r.nestedSet.DeleteSubtree(ctx, "menus", menuID)
	if err != nil {
		return fmt.Errorf("failed to delete subtree using nested set: %w", err)
	}

	return nil
}

func (r *menuRepository) InsertBetween(ctx context.Context, menu *entities.Menu, leftSiblingID, rightSiblingID *uuid.UUID) error {
	// Calculate nested set values for insertion between siblings
	var leftValue, rightValue int64

	if leftSiblingID != nil {
		query := `SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL`
		err := r.db.QueryRow(ctx, query, leftSiblingID.String()).Scan(&leftValue)
		if err != nil {
			return fmt.Errorf("failed to get left sibling right value: %w", err)
		}
	} else {
		leftValue = 0
	}

	if rightSiblingID != nil {
		query := `SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL`
		err := r.db.QueryRow(ctx, query, rightSiblingID.String()).Scan(&rightValue)
		if err != nil {
			return fmt.Errorf("failed to get right sibling left value: %w", err)
		}
	} else {
		// Get the maximum right value and add 1
		query := `SELECT COALESCE(MAX(record_right), 0) + 1 FROM menus WHERE deleted_at IS NULL`
		err := r.db.QueryRow(ctx, query).Scan(&rightValue)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
	}

	// Calculate new values
	newLeft := uint64(leftValue + 1)
	newRight := uint64(newLeft + 1)
	newDepth := uint64(0) // Root level

	// Assign computed nested set values to the entity
	menu.RecordLeft = &newLeft
	menu.RecordRight = &newRight
	menu.RecordDepth = &newDepth
	menu.RecordOrdering = &newLeft

	// Insert the new menu
	query := `
		INSERT INTO menus (
			id, name, slug, description, url, icon, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			is_active, is_visible, target, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	parentIDStr := ""
	if menu.ParentID != nil {
		parentIDStr = menu.ParentID.String()
	}

	_, err := r.db.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordLeft, menu.RecordRight, menu.RecordDepth, menu.RecordOrdering,
		menu.IsActive, menu.IsVisible, menu.Target, menu.CreatedBy, menu.UpdatedBy, menu.CreatedAt, menu.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create menu: %w", err)
	}

	return nil
}

func (r *menuRepository) SwapPositions(ctx context.Context, menu1ID, menu2ID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current positions
	var left1, right1, left2, right2 int64
	query := `SELECT record_left, record_right FROM menus WHERE id = $1 AND deleted_at IS NULL`

	err = tx.QueryRow(ctx, query, menu1ID.String()).Scan(&left1, &right1)
	if err != nil {
		return fmt.Errorf("failed to get menu1 position: %w", err)
	}

	err = tx.QueryRow(ctx, query, menu2ID.String()).Scan(&left2, &right2)
	if err != nil {
		return fmt.Errorf("failed to get menu2 position: %w", err)
	}

	// Calculate size of each subtree
	size1 := right1 - left1 + 1
	size2 := right2 - left1 + 1

	// Swap positions
	_, err = tx.Exec(ctx, `UPDATE menus SET record_left = record_left + $1, record_right = record_right + $1 WHERE record_left >= $2 AND record_right <= $3 AND deleted_at IS NULL`, size2, left1, right1)
	if err != nil {
		return fmt.Errorf("failed to move menu1: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE menus SET record_left = record_left - $1, record_right = record_right - $1 WHERE record_left >= $2 AND record_right <= $3 AND deleted_at IS NULL`, size1, left2, right2)
	if err != nil {
		return fmt.Errorf("failed to move menu2: %w", err)
	}

	// Update timestamps
	_, err = tx.Exec(ctx, `UPDATE menus SET updated_at = NOW() WHERE id IN ($1, $2)`, menu1ID.String(), menu2ID.String())
	if err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *menuRepository) GetLeafNodes(ctx context.Context) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE record_left + 1 = record_right AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf nodes: %w", err)
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

func (r *menuRepository) GetInternalNodes(ctx context.Context) ([]*entities.Menu, error) {
	query := `
		SELECT id, name, slug, description, url, icon, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       is_active, is_visible, target, created_by, updated_by, created_at, updated_at, deleted_at
		FROM menus 
		WHERE record_left + 1 < record_right AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal nodes: %w", err)
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

// Batch Operations

func (r *menuRepository) BatchMoveSubtrees(ctx context.Context, moves []struct {
	MenuID      uuid.UUID
	NewParentID uuid.UUID
}) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, move := range moves {
		err := r.nestedSet.MoveSubtree(ctx, "menus", move.MenuID, move.NewParentID)
		if err != nil {
			return fmt.Errorf("failed to move subtree %s: %w", move.MenuID, err)
		}

		// Update timestamp
		_, err = tx.Exec(ctx, `UPDATE menus SET updated_at = NOW() WHERE id = $1`, move.MenuID.String())
		if err != nil {
			return fmt.Errorf("failed to update timestamp for %s: %w", move.MenuID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *menuRepository) BatchInsertBetween(ctx context.Context, insertions []struct {
	Menu           *entities.Menu
	LeftSiblingID  *uuid.UUID
	RightSiblingID *uuid.UUID
}) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, insertion := range insertions {
		err := r.insertBetweenInTx(ctx, tx, insertion.Menu, insertion.LeftSiblingID, insertion.RightSiblingID)
		if err != nil {
			return fmt.Errorf("failed to insert menu %s: %w", insertion.Menu.ID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *menuRepository) insertBetweenInTx(ctx context.Context, tx pgx.Tx, menu *entities.Menu, leftSiblingID, rightSiblingID *uuid.UUID) error {
	// Calculate nested set values for insertion between siblings
	var leftValue, rightValue int64

	if leftSiblingID != nil {
		query := `SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL`
		err := tx.QueryRow(ctx, query, leftSiblingID.String()).Scan(&leftValue)
		if err != nil {
			return fmt.Errorf("failed to get left sibling right value: %w", err)
		}
	} else {
		leftValue = 0
	}

	if rightSiblingID != nil {
		query := `SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL`
		err := tx.QueryRow(ctx, query, rightSiblingID.String()).Scan(&rightValue)
		if err != nil {
			return fmt.Errorf("failed to get right sibling left value: %w", err)
		}
	} else {
		// Get the maximum right value and add 1
		query := `SELECT COALESCE(MAX(record_right), 0) + 1 FROM menus WHERE deleted_at IS NULL`
		err := tx.QueryRow(ctx, query).Scan(&rightValue)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
	}

	// Calculate new values
	newLeft := uint64(leftValue + 1)
	newRight := uint64(newLeft + 1)
	newDepth := uint64(0) // Root level

	// Assign computed nested set values to the entity
	menu.RecordLeft = &newLeft
	menu.RecordRight = &newRight
	menu.RecordDepth = &newDepth
	menu.RecordOrdering = &newLeft

	// Insert the new menu
	query := `
		INSERT INTO menus (
			id, name, slug, description, url, icon, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			is_active, is_visible, target, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	parentIDStr := ""
	if menu.ParentID != nil {
		parentIDStr = menu.ParentID.String()
	}

	_, err := tx.Exec(ctx, query,
		menu.ID, menu.Name, menu.Slug, menu.Description, menu.URL, menu.Icon, parentIDStr,
		menu.RecordLeft, menu.RecordRight, menu.RecordDepth, menu.RecordOrdering,
		menu.IsActive, menu.IsVisible, menu.Target, menu.CreatedBy, menu.UpdatedBy, menu.CreatedAt, menu.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create menu: %w", err)
	}

	return nil
}

// Tree Maintenance and Optimization

func (r *menuRepository) ValidateTree(ctx context.Context) ([]string, error) {
	var issues []string

	// Check for overlapping intervals
	query := `
		SELECT COUNT(*) FROM menus m1, menus m2 
		WHERE m1.id != m2.id 
		  AND m1.deleted_at IS NULL 
		  AND m2.deleted_at IS NULL
		  AND m1.record_left <= m2.record_left 
		  AND m1.record_right >= m2.record_right
	`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to validate tree: %w", err)
	}

	if count > 0 {
		issues = append(issues, "Found overlapping nested set intervals")
	}

	// Check for gaps in left values
	query = `
		SELECT COUNT(*) FROM (
			SELECT record_left, LAG(record_left) OVER (ORDER BY record_left) as prev_left
			FROM menus WHERE deleted_at IS NULL
		) gaps WHERE record_left - prev_left > 1
	`
	err = r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to validate tree gaps: %w", err)
	}

	if count > 0 {
		issues = append(issues, "Found gaps in nested set left values")
	}

	return issues, nil
}

func (r *menuRepository) RebuildTree(ctx context.Context) error {
	// This is a complex operation that would require rebuilding the entire tree
	// For now, we'll implement a basic version that validates and reports issues
	issues, err := r.ValidateTree(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate tree before rebuild: %w", err)
	}

	if len(issues) > 0 {
		return fmt.Errorf("tree has validation issues that need to be resolved before rebuild: %v", issues)
	}

	return nil
}

func (r *menuRepository) OptimizeTree(ctx context.Context) error {
	// Analyze table to update statistics
	_, err := r.db.Exec(ctx, `ANALYZE menus`)
	if err != nil {
		return fmt.Errorf("failed to analyze menus table: %w", err)
	}

	// Reindex if needed
	_, err = r.db.Exec(ctx, `REINDEX TABLE menus`)
	if err != nil {
		return fmt.Errorf("failed to reindex menus table: %w", err)
	}

	return nil
}

func (r *menuRepository) GetTreeStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var totalCount int64
	query := `SELECT COUNT(*) FROM menus WHERE deleted_at IS NULL`
	err := r.db.QueryRow(ctx, query).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_count"] = totalCount

	// Tree height
	height, err := r.GetTreeHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree height: %w", err)
	}
	stats["tree_height"] = height

	// Leaf nodes count
	var leafCount int64
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM menus WHERE record_left + 1 = record_right AND deleted_at IS NULL`).Scan(&leafCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf count: %w", err)
	}
	stats["leaf_count"] = leafCount

	// Internal nodes count
	var internalCount int64
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM menus WHERE record_left + 1 < record_right AND deleted_at IS NULL`).Scan(&internalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal count: %w", err)
	}
	stats["internal_count"] = internalCount

	return stats, nil
}

func (r *menuRepository) GetTreeHeight(ctx context.Context) (uint64, error) {
	query := `SELECT COALESCE(MAX(record_depth), 0) FROM menus WHERE deleted_at IS NULL`
	var height uint64
	err := r.db.QueryRow(ctx, query).Scan(&height)
	if err != nil {
		return 0, fmt.Errorf("failed to get tree height: %w", err)
	}
	return height, nil
}

func (r *menuRepository) GetLevelWidth(ctx context.Context, level uint64) (int64, error) {
	query := `SELECT COUNT(*) FROM menus WHERE record_depth = $1 AND deleted_at IS NULL`
	var width int64
	err := r.db.QueryRow(ctx, query, level).Scan(&width)
	if err != nil {
		return 0, fmt.Errorf("failed to get level width: %w", err)
	}
	return width, nil
}

func (r *menuRepository) GetSubtreeSize(ctx context.Context, menuID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*) FROM menus 
		WHERE record_left >= (SELECT record_left FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND record_right <= (SELECT record_right FROM menus WHERE id = $1 AND deleted_at IS NULL)
		  AND deleted_at IS NULL
	`
	var size int64
	err := r.db.QueryRow(ctx, query, menuID.String()).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("failed to get subtree size: %w", err)
	}
	return size, nil
}

func (r *menuRepository) GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Query execution time for common operations
	start := time.Now()
	_, err := r.GetRootMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to measure root menus query: %w", err)
	}
	metrics["root_menus_query_time"] = time.Since(start).Milliseconds()

	start = time.Now()
	_, err = r.GetAll(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to measure get all query: %w", err)
	}
	metrics["get_all_query_time"] = time.Since(start).Milliseconds()

	// Table size information
	var tableSize, indexSize int64
	query := `
		SELECT 
			pg_total_relation_size('menus') as table_size,
			pg_indexes_size('menus') as index_size
	`
	err = r.db.QueryRow(ctx, query).Scan(&tableSize, &indexSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get table size info: %w", err)
	}
	metrics["table_size_bytes"] = tableSize
	metrics["index_size_bytes"] = indexSize

	return metrics, nil
}

func (r *menuRepository) ValidateTreeIntegrity(ctx context.Context) ([]string, error) {
	var issues []string

	// Check for orphaned nodes (nodes with parent_id pointing to non-existent nodes)
	query := `
		SELECT COUNT(*) FROM menus m1
		LEFT JOIN menus m2 ON m1.parent_id = m2.id
		WHERE m1.parent_id IS NOT NULL 
		  AND m2.id IS NULL 
		  AND m1.deleted_at IS NULL
	`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check orphaned nodes: %w", err)
	}

	if count > 0 {
		issues = append(issues, fmt.Sprintf("Found %d orphaned nodes", count))
	}

	// Check for circular references
	query = `
		WITH RECURSIVE cycle_check AS (
			SELECT id, parent_id, ARRAY[id] as path
			FROM menus WHERE parent_id IS NOT NULL AND deleted_at IS NULL
			UNION ALL
			SELECT m.id, m.parent_id, cc.path || m.id
			FROM menus m
			JOIN cycle_check cc ON m.id = cc.parent_id
			WHERE m.deleted_at IS NULL AND NOT (m.id = ANY(cc.path))
		)
		SELECT COUNT(*) FROM cycle_check WHERE parent_id = ANY(path)
	`
	err = r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check circular references: %w", err)
	}

	if count > 0 {
		issues = append(issues, "Found circular references in tree structure")
	}

	return issues, nil
}

// scanMenuRow is a helper function to scan a menu row from database
func (r *menuRepository) scanMenuRow(rows pgx.Rows) (*entities.Menu, error) {
	var menu entities.Menu
	var parentIDStr *string

	err := rows.Scan(
		&menu.ID, &menu.Name, &menu.Slug, &menu.Description, &menu.URL, &menu.Icon, &parentIDStr,
		&menu.RecordLeft, &menu.RecordRight, &menu.RecordDepth, &menu.RecordOrdering,
		&menu.IsActive, &menu.IsVisible, &menu.Target, &menu.CreatedBy, &menu.UpdatedBy, &menu.CreatedAt, &menu.UpdatedAt, &menu.DeletedAt)
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
