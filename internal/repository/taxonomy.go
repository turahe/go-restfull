package repository

import (
	"context"
	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TaxonomyRepository interface {
	Create(ctx context.Context, taxonomy *model.Taxonomy) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Taxonomy, error)
	GetBySlug(ctx context.Context, slug string) (*model.Taxonomy, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Taxonomy, error)
	GetRootTaxonomies(ctx context.Context) ([]*model.Taxonomy, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Taxonomy, error)
	GetHierarchy(ctx context.Context) ([]*model.Taxonomy, error)
	GetDescendants(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error)
	GetAncestors(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error)
	GetSiblings(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*model.Taxonomy, error)
	Update(ctx context.Context, taxonomy *model.Taxonomy) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
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

func (r *TaxonomyRepositoryImpl) Create(ctx context.Context, taxonomy *model.Taxonomy) error {
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
		err = tx.QueryRow(ctx, `SELECT record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL`, *taxonomy.ParentID).Scan(&parentRight)
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

	_, err = tx.Exec(ctx, query,
		taxonomy.ID, taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		taxonomy.ParentID, taxonomy.RecordLeft, taxonomy.RecordRight, taxonomy.RecordDepth,
		taxonomy.CreatedAt, taxonomy.UpdatedAt, taxonomy.CreatedBy, taxonomy.UpdatedBy,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *TaxonomyRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
	`

	var taxonomy model.Taxonomy
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
	)

	if err != nil {
		return nil, err
	}

	return &taxonomy, nil
}

func (r *TaxonomyRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*model.Taxonomy, error) {
	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL
	`

	var taxonomy model.Taxonomy
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
	)

	if err != nil {
		return nil, err
	}

	return &taxonomy, nil
}

func (r *TaxonomyRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetRootTaxonomies(ctx context.Context) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetHierarchy(ctx context.Context) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetDescendants(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetAncestors(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetSiblings(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*model.Taxonomy, error) {
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

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		var taxonomy model.Taxonomy
		err := rows.Scan(
			&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
			&taxonomy.ParentID, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
			&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &taxonomy.CreatedBy, &taxonomy.UpdatedBy, &taxonomy.DeletedBy,
		)
		if err != nil {
			return nil, err
		}
		taxonomies = append(taxonomies, &taxonomy)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) Update(ctx context.Context, taxonomy *model.Taxonomy) error {
	// For nested set, we need to handle parent changes carefully
	// This is a simplified update that doesn't change the tree structure
	query := `
		UPDATE taxonomies 
		SET name = $2, slug = $3, code = $4, description = $5, updated_at = $6, updated_by = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pgxPool.Exec(ctx, query,
		taxonomy.ID, taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		taxonomy.UpdatedAt, taxonomy.UpdatedBy,
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
