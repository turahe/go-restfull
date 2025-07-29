package repository

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TaxonomyRepository interface {
	Create(ctx context.Context, taxonomy *entities.Taxonomy) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error)
	GetAllWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error)
	GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error)
	GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	Update(ctx context.Context, taxonomy *entities.Taxonomy) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountWithSearch(ctx context.Context, query string) (int64, error)
}

type TaxonomyRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewTaxonomyRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) TaxonomyRepository {
	return &TaxonomyRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *TaxonomyRepositoryImpl) Create(ctx context.Context, taxonomy *entities.Taxonomy) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// If this is a root taxonomy (no parent)
	if taxonomy.ParentID == nil {
		// Get the maximum right value and add 1 for the new node
		var maxRight int64
		err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(record_right), 0) FROM taxonomies WHERE deleted_at IS NULL`).Scan(&maxRight)
		if err != nil {
			return err
		}

		taxonomy.RecordLeft = maxRight + 1
		taxonomy.RecordRight = maxRight + 2
		taxonomy.RecordDepth = 0
	} else {
		// Get the parent's right value
		var parentRight int64
		err = tx.QueryRow(ctx, `SELECT record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL`, taxonomy.ParentID.String()).Scan(&parentRight)
		if err != nil {
			return err
		}

		// Make space for the new node by shifting all nodes to the right
		_, err = tx.Exec(ctx, `
			UPDATE taxonomies 
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

		taxonomy.RecordLeft = parentRight
		taxonomy.RecordRight = parentRight + 1
		taxonomy.RecordDepth = 0 // Will be calculated based on parent
	}

	// Insert the new taxonomy
	query := `
		INSERT INTO taxonomies (id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	parentIDStr := ""
	if taxonomy.ParentID != nil {
		parentIDStr = taxonomy.ParentID.String()
	}

	_, err = tx.Exec(ctx, query,
		taxonomy.ID.String(), taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		parentIDStr, taxonomy.RecordLeft, taxonomy.RecordRight, taxonomy.RecordDepth,
		taxonomy.CreatedAt, taxonomy.UpdatedAt, "", "", // created_by, updated_by
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *TaxonomyRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
	`

	var taxonomy entities.Taxonomy
	var parentIDStr *string
	var createdBy, updatedBy, deletedBy string

	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&parentIDStr, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &createdBy, &updatedBy, &deletedBy,
	)

	if err != nil {
		return nil, err
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	return &taxonomy, nil
}

func (r *TaxonomyRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL
	`

	var taxonomy entities.Taxonomy
	var parentIDStr *string
	var createdBy, updatedBy, deletedBy string

	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&parentIDStr, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &createdBy, &updatedBy, &deletedBy,
	)

	if err != nil {
		return nil, err
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	return &taxonomy, nil
}

func (r *TaxonomyRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

// scanTaxonomyRow is a helper function to scan a taxonomy row from database
func (r *TaxonomyRepositoryImpl) scanTaxonomyRow(rows pgx.Rows) (*entities.Taxonomy, error) {
	var taxonomy entities.Taxonomy
	var parentIDStr *string
	var createdBy, updatedBy, deletedBy string

	err := rows.Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&parentIDStr, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &createdBy, &updatedBy, &deletedBy,
	)
	if err != nil {
		return nil, err
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	return &taxonomy, nil
}

func (r *TaxonomyRepositoryImpl) GetAllWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	var sqlQuery string
	var args []interface{}

	if query == "" {
		// If no search query, return all taxonomies with pagination
		sqlQuery = `
			SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
			FROM taxonomies WHERE deleted_at IS NULL
			ORDER BY record_left ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	} else {
		// If search query provided, search with pagination
		sqlQuery = `
			SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
			FROM taxonomies 
			WHERE (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL
			ORDER BY record_left ASC
			LIMIT $2 OFFSET $3
		`
		searchTerm := "%" + query + "%"
		args = []interface{}{searchTerm, limit, offset}
	}

	rows, err := r.pgxPool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, parentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies 
		WHERE record_left > (SELECT record_left FROM taxonomies WHERE id = $1 AND deleted_at IS NULL)
		AND record_right < (SELECT record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL)
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies 
		WHERE record_left < (SELECT record_left FROM taxonomies WHERE id = $1 AND deleted_at IS NULL)
		AND record_right > (SELECT record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL)
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	query := `
		SELECT t1.id, t1.name, t1.slug, t1.code, t1.description, t1.parent_id, t1.record_left, t1.record_right, t1.record_depth, t1.created_at, t1.updated_at, t1.deleted_at, t1.created_by, t1.updated_by, t1.deleted_by
		FROM taxonomies t1
		JOIN taxonomies t2 ON t1.parent_id = t2.parent_id
		WHERE t2.id = $1 AND t1.id != $1 AND t1.deleted_at IS NULL
		ORDER BY t1.record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	searchQuery := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies 
		WHERE (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxonomies []*entities.Taxonomy
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) Update(ctx context.Context, taxonomy *entities.Taxonomy) error {
	// For nested set, we need to handle parent changes carefully
	// This is a simplified update that doesn't change the tree structure
	query := `
		UPDATE taxonomies 
		SET name = $2, slug = $3, code = $4, description = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pgxPool.Exec(ctx, query,
		taxonomy.ID.String(), taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		taxonomy.UpdatedAt,
	)

	return err
}

func (r *TaxonomyRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the node's left and right values
	var left, right int64
	err = tx.QueryRow(ctx, `SELECT record_left, record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL`, id.String()).Scan(&left, &right)
	if err != nil {
		return err
	}

	// Calculate the width of the subtree
	width := right - left + 1

	// Delete the node and all its descendants
	_, err = tx.Exec(ctx, `UPDATE taxonomies SET deleted_at = NOW() WHERE record_left >= $1 AND record_right <= $2 AND deleted_at IS NULL`, left, right)
	if err != nil {
		return err
	}

	// Close the gap by shifting all nodes to the left
	_, err = tx.Exec(ctx, `
		UPDATE taxonomies 
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

func (r *TaxonomyRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

func (r *TaxonomyRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM taxonomies WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *TaxonomyRepositoryImpl) CountWithSearch(ctx context.Context, query string) (int64, error) {
	var sqlQuery string
	var args []interface{}

	if query == "" {
		// If no search query, count all taxonomies
		sqlQuery = `SELECT COUNT(*) FROM taxonomies WHERE deleted_at IS NULL`
		args = []interface{}{}
	} else {
		// If search query provided, count matching taxonomies
		sqlQuery = `
			SELECT COUNT(*) 
			FROM taxonomies 
			WHERE (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL
		`
		searchTerm := "%" + query + "%"
		args = []interface{}{searchTerm}
	}

	var count int64
	err := r.pgxPool.QueryRow(ctx, sqlQuery, args...).Scan(&count)
	return count, err
}
