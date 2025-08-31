package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTaxonomyRepository struct {
	db        *pgxpool.Pool
	nestedSet *nestedset.NestedSetManager
}

func NewPostgresTaxonomyRepository(db *pgxpool.Pool) repositories.TaxonomyRepository {
	return &PostgresTaxonomyRepository{db: db, nestedSet: nestedset.NewNestedSetManager(db)}
}

func (r *PostgresTaxonomyRepository) Create(ctx context.Context, taxonomy *entities.Taxonomy) error {
	// Set current timestamps if not already set
	now := time.Now()
	if taxonomy.CreatedAt.IsZero() {
		taxonomy.CreatedAt = now
	}
	if taxonomy.UpdatedAt.IsZero() {
		taxonomy.UpdatedAt = now
	}

	// Calculate nested set values
	nestedSetValues, err := r.nestedSet.CreateNode(ctx, "taxonomies", taxonomy.ParentID, int64(1))
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign nested set values to taxonomy entity
	taxonomy.RecordLeft = &nestedSetValues.Left
	taxonomy.RecordRight = &nestedSetValues.Right
	taxonomy.RecordDepth = &nestedSetValues.Depth
	taxonomy.RecordOrdering = &nestedSetValues.Ordering

	query := `
		INSERT INTO taxonomies (id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`
	var parentID *string
	if taxonomy.ParentID != nil {
		parentIDStr := taxonomy.ParentID.String()
		parentID = &parentIDStr
	}
	// Handle user fields conditionally - use nil for empty UUIDs
	var createdBy, updatedBy *string
	if taxonomy.CreatedBy != uuid.Nil {
		createdByStr := taxonomy.CreatedBy.String()
		createdBy = &createdByStr
	}
	if taxonomy.UpdatedBy != uuid.Nil {
		updatedByStr := taxonomy.UpdatedBy.String()
		updatedBy = &updatedByStr
	}

	_, err = r.db.Exec(ctx, query,
		taxonomy.ID, taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		parentID, taxonomy.RecordLeft, taxonomy.RecordRight, taxonomy.RecordDepth,
		taxonomy.CreatedAt, taxonomy.UpdatedAt, createdBy, updatedBy,
	)
	return err
}

// createTaxonomyFallback creates a taxonomy with manual nested set values when the nested set manager fails
func (r *PostgresTaxonomyRepository) createTaxonomyFallback(ctx context.Context, taxonomy *entities.Taxonomy) error {
	// Set current timestamps if not already set
	now := time.Now()
	if taxonomy.CreatedAt.IsZero() {
		taxonomy.CreatedAt = now
	}
	if taxonomy.UpdatedAt.IsZero() {
		taxonomy.UpdatedAt = now
	}

	// Get the maximum right value from existing taxonomies
	var maxRight int
	query := `SELECT COALESCE(MAX(record_right), 0) FROM taxonomies`
	if err := r.db.QueryRow(ctx, query).Scan(&maxRight); err != nil {
		return err
	}

	// Calculate new nested set values
	left := int64(maxRight + 1)
	right := int64(maxRight + 2)
	depth := int64(0)

	// If this is a child taxonomy, calculate depth based on parent
	if taxonomy.ParentID != nil {
		var parentDepth int
		parentQuery := `SELECT record_depth FROM taxonomies WHERE id = $1`
		if err := r.db.QueryRow(ctx, parentQuery, taxonomy.ParentID.String()).Scan(&parentDepth); err != nil {
			// If parent not found, treat as root
			depth = 0
		} else {
			depth = int64(parentDepth + 1)
		}
	}

	// Set the calculated values
	taxonomy.RecordLeft = &left
	taxonomy.RecordRight = &right
	taxonomy.RecordDepth = &depth

	// Insert with calculated nested set values
	insertQuery := `
		INSERT INTO taxonomies (id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`
	var parentID *string
	if taxonomy.ParentID != nil {
		parentIDStr := taxonomy.ParentID.String()
		parentID = &parentIDStr
	}
	// Handle user fields conditionally - use nil for empty UUIDs
	var createdBy, updatedBy *string
	if taxonomy.CreatedBy != uuid.Nil {
		createdByStr := taxonomy.CreatedBy.String()
		createdBy = &createdByStr
	}
	if taxonomy.UpdatedBy != uuid.Nil {
		updatedByStr := taxonomy.UpdatedBy.String()
		updatedBy = &updatedByStr
	}

	_, err := r.db.Exec(ctx, insertQuery,
		taxonomy.ID, taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		parentID, taxonomy.RecordLeft, taxonomy.RecordRight, taxonomy.RecordDepth,
		taxonomy.CreatedAt, taxonomy.UpdatedAt, createdBy, updatedBy,
	)
	return err
}

func (r *PostgresTaxonomyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
	`
	var t entities.Taxonomy
	var parentIDStr *string
	var cb, ub, dbs *string
	if err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
		return nil, err
	}
	if parentIDStr != nil {
		if p, err := uuid.Parse(*parentIDStr); err == nil {
			t.ParentID = &p
		}
	}
	return &t, nil
}

func (r *PostgresTaxonomyRepository) GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL
	`
	var t entities.Taxonomy
	var parentIDStr *string
	var cb, ub, dbs *string
	if err := r.db.QueryRow(ctx, query, slug).Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
		return nil, err
	}
	if parentIDStr != nil {
		if p, err := uuid.Parse(*parentIDStr); err == nil {
			t.ParentID = &p
		}
	}
	return &t, nil
}

func (r *PostgresTaxonomyRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL ORDER BY record_left ASC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetAllWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	q := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL AND (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1)
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`
	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE parent_id IS NULL AND deleted_at IS NULL ORDER BY record_left ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE parent_id = $1 AND deleted_at IS NULL ORDER BY record_left ASC
	`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL ORDER BY record_left ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	ids, err := r.nestedSet.GetDescendants(ctx, "taxonomies", id, 10000, 0)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Taxonomy{}, nil
	}
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	ids, err := r.nestedSet.GetAncestors(ctx, "taxonomies", id)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Taxonomy{}, nil
	}
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	ids, err := r.nestedSet.GetSiblings(ctx, "taxonomies", id, 10000, 0)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entities.Taxonomy{}, nil
	}
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = ANY($1) AND deleted_at IS NULL ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	q := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL AND (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1)
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`
	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Taxonomy
	for rows.Next() {
		var t entities.Taxonomy
		var parentIDStr *string
		var cb, ub, dbs *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Code, &t.Description, &parentIDStr, &t.RecordLeft, &t.RecordRight, &t.RecordDepth, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &cb, &ub, &dbs); err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				t.ParentID = &p
			}
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresTaxonomyRepository) Update(ctx context.Context, taxonomy *entities.Taxonomy) error {
	// Set current timestamp for updated_at
	taxonomy.UpdatedAt = time.Now()

	// Handle updated_by field conditionally - use nil for empty UUIDs
	var updatedBy *string
	if taxonomy.UpdatedBy != uuid.Nil {
		updatedByStr := taxonomy.UpdatedBy.String()
		updatedBy = &updatedByStr
	}

	query := `UPDATE taxonomies SET name=$2, slug=$3, code=$4, description=$5, updated_by=$6, updated_at=$7 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, taxonomy.ID, taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description, updatedBy, taxonomy.UpdatedAt)
	return err
}

func (r *PostgresTaxonomyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE taxonomies SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTaxonomyRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresTaxonomyRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM taxonomies WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresTaxonomyRepository) CountWithSearch(ctx context.Context, query string) (int64, error) {
	var sql string
	var args []interface{}
	if query == "" {
		sql = `SELECT COUNT(*) FROM taxonomies WHERE deleted_at IS NULL`
	} else {
		sql = `SELECT COUNT(*) FROM taxonomies WHERE (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL`
		args = []interface{}{"%" + query + "%"}
	}
	var count int64
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
