package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// TaxonomyRepository defines the interface for taxonomy operations
type TaxonomyRepository interface {
	Create(ctx context.Context, taxonomy *model.Taxonomy) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Taxonomy, error)
	GetAll(ctx context.Context) ([]*model.Taxonomy, error)
	Update(ctx context.Context, taxonomy *model.Taxonomy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.Taxonomy, error)
	GetDescendants(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error)
	GetAncestors(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error)
	GetSiblings(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error)
	GetRootNodes(ctx context.Context) ([]*model.Taxonomy, error)
	GetPaginated(ctx context.Context, query string, limit, page int) ([]*model.Taxonomy, int, error)
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

// Create inserts a new taxonomy into the database
func (r *TaxonomyRepositoryImpl) Create(ctx context.Context, taxonomy *model.Taxonomy) error {
	now := time.Now()
	taxonomy.ID = uuid.New()
	taxonomy.CreatedAt = now
	taxonomy.UpdatedAt = now

	// Start a transaction for nested set operations
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if taxonomy.ParentID != nil {
		// Insert as child of existing parent
		query := `
			INSERT INTO taxonomies (id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 
				(SELECT record_right FROM taxonomies WHERE id = $4),
				(SELECT record_right + 1 FROM taxonomies WHERE id = $4),
				(SELECT record_depth + 1 FROM taxonomies WHERE id = $4),
				$5, $6)
		`
		_, err = tx.Exec(ctx, query,
			taxonomy.ID,
			taxonomy.Name,
			taxonomy.Description,
			taxonomy.ParentID,
			taxonomy.CreatedAt,
			taxonomy.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create taxonomy: %w", err)
		}

		// Update the record_left and record_right values for all nodes to the right of the parent
		updateQuery := `
			UPDATE taxonomies 
			SET record_left = CASE 
				WHEN record_left > (SELECT record_right FROM taxonomies WHERE id = $1) THEN record_left + 2
				ELSE record_left
			END,
			record_right = CASE 
				WHEN record_right >= (SELECT record_right FROM taxonomies WHERE id = $1) THEN record_right + 2
				ELSE record_right
			END
			WHERE record_right >= (SELECT record_right FROM taxonomies WHERE id = $1)
		`
		_, err = tx.Exec(ctx, updateQuery, taxonomy.ParentID)
		if err != nil {
			return fmt.Errorf("failed to update nested set values: %w", err)
		}

		// Update the new node's record_left and record_right
		finalUpdateQuery := `
			UPDATE taxonomies 
			SET record_left = (SELECT record_right - 1 FROM taxonomies WHERE id = $1),
				record_right = (SELECT record_right FROM taxonomies WHERE id = $1)
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, finalUpdateQuery, taxonomy.ParentID, taxonomy.ID)
		if err != nil {
			return fmt.Errorf("failed to update new node values: %w", err)
		}
	} else {
		// Insert as root node
		// Find the maximum record_right value
		var maxRight int64
		err = tx.QueryRow(ctx, "SELECT COALESCE(MAX(record_right), 0) FROM taxonomies").Scan(&maxRight)
		if err != nil {
			return fmt.Errorf("failed to get max record_right: %w", err)
		}

		taxonomy.RecordLeft = uint64(maxRight + 1)
		taxonomy.RecordRight = uint64(maxRight + 2)
		taxonomy.RecordDepth = 0

		query := `
			INSERT INTO taxonomies (id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(ctx, query,
			taxonomy.ID,
			taxonomy.Name,
			taxonomy.Description,
			taxonomy.ParentID,
			taxonomy.RecordLeft,
			taxonomy.RecordRight,
			taxonomy.RecordDepth,
			taxonomy.CreatedAt,
			taxonomy.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create taxonomy: %w", err)
		}
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a taxonomy by its ID
func (r *TaxonomyRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Taxonomy, error) {
	query := `
		SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at
		FROM taxonomies
		WHERE id = $1
	`

	taxonomy := &model.Taxonomy{}
	err := r.pgxPool.QueryRow(ctx, query, id).Scan(
		&taxonomy.ID,
		&taxonomy.Name,
		&taxonomy.Description,
		&taxonomy.ParentID,
		&taxonomy.RecordLeft,
		&taxonomy.RecordRight,
		&taxonomy.RecordDepth,
		&taxonomy.CreatedAt,
		&taxonomy.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("taxonomy not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get taxonomy: %w", err)
	}

	return taxonomy, nil
}

// GetAll retrieves all taxonomies from the database
func (r *TaxonomyRepositoryImpl) GetAll(ctx context.Context) ([]*model.Taxonomy, error) {
	query := `
		SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at
		FROM taxonomies
		ORDER BY record_left
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get taxonomies: %w", err)
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over taxonomies: %w", err)
	}

	return taxonomies, nil
}

// Update updates an existing taxonomy in the database
func (r *TaxonomyRepositoryImpl) Update(ctx context.Context, taxonomy *model.Taxonomy) error {
	query := `
		UPDATE taxonomies
		SET name = $1, description = $2, parent_id = $3, updated_at = $4
		WHERE id = $5
	`

	taxonomy.UpdatedAt = time.Now()

	result, err := r.pgxPool.Exec(ctx, query,
		taxonomy.Name,
		taxonomy.Description,
		taxonomy.ParentID,
		taxonomy.UpdatedAt,
		taxonomy.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update taxonomy: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("taxonomy not found")
	}

	return nil
}

// Delete removes a taxonomy from the database
func (r *TaxonomyRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM taxonomies WHERE id = $1`

	result, err := r.pgxPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete taxonomy: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("taxonomy not found")
	}

	return nil
}

// GetByParentID retrieves all taxonomies that have a specific parent ID
func (r *TaxonomyRepositoryImpl) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.Taxonomy, error) {
	query := `
		SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at
		FROM taxonomies
		WHERE parent_id = $1
		ORDER BY record_left
	`

	rows, err := r.pgxPool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get taxonomies by parent ID: %w", err)
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over taxonomies: %w", err)
	}

	return taxonomies, nil
}

// GetDescendants retrieves all descendants of a taxonomy using nested set
func (r *TaxonomyRepositoryImpl) GetDescendants(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error) {
	query := `
		SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at
		FROM taxonomies
		WHERE record_left > (SELECT record_left FROM taxonomies WHERE id = $1)
		AND record_right < (SELECT record_right FROM taxonomies WHERE id = $1)
		ORDER BY record_left
	`

	rows, err := r.pgxPool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over taxonomies: %w", err)
	}

	return taxonomies, nil
}

// GetAncestors retrieves all ancestors of a taxonomy using nested set
func (r *TaxonomyRepositoryImpl) GetAncestors(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error) {
	query := `
		SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at
		FROM taxonomies
		WHERE record_left < (SELECT record_left FROM taxonomies WHERE id = $1)
		AND record_right > (SELECT record_right FROM taxonomies WHERE id = $1)
		ORDER BY record_left
	`

	rows, err := r.pgxPool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over taxonomies: %w", err)
	}

	return taxonomies, nil
}

// GetSiblings retrieves all siblings of a taxonomy
func (r *TaxonomyRepositoryImpl) GetSiblings(ctx context.Context, id uuid.UUID) ([]*model.Taxonomy, error) {
	query := `
		SELECT t1.id, t1.name, t1.description, t1.parent_id, t1.record_left, t1.record_right, t1.record_depth, t1.created_at, t1.updated_at
		FROM taxonomies t1
		JOIN taxonomies t2 ON t1.parent_id = t2.parent_id
		WHERE t2.id = $1 AND t1.id != $1
		ORDER BY t1.record_left
	`

	rows, err := r.pgxPool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over taxonomies: %w", err)
	}

	return taxonomies, nil
}

// GetRootNodes retrieves all root taxonomies (nodes without parents)
func (r *TaxonomyRepositoryImpl) GetRootNodes(ctx context.Context) ([]*model.Taxonomy, error) {
	query := `
		SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at
		FROM taxonomies
		WHERE parent_id IS NULL
		ORDER BY record_left
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root nodes: %w", err)
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over taxonomies: %w", err)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepositoryImpl) GetPaginated(ctx context.Context, query string, limit, page int) ([]*model.Taxonomy, int, error) {
	var where string
	var args []interface{}
	if query != "" {
		where = "WHERE name ILIKE $1 OR description ILIKE $1"
		args = append(args, "%"+query+"%")
	}
	countQuery := "SELECT COUNT(*) FROM taxonomies " + where
	var total int
	if err := r.pgxPool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	listQuery := "SELECT id, name, description, parent_id, record_left, record_right, record_depth, created_at, updated_at FROM taxonomies " + where + " ORDER BY record_left LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)
	rows, err := r.pgxPool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var taxonomies []*model.Taxonomy
	for rows.Next() {
		taxonomy := &model.Taxonomy{}
		err := rows.Scan(
			&taxonomy.ID,
			&taxonomy.Name,
			&taxonomy.Description,
			&taxonomy.ParentID,
			&taxonomy.RecordLeft,
			&taxonomy.RecordRight,
			&taxonomy.RecordDepth,
			&taxonomy.CreatedAt,
			&taxonomy.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		taxonomies = append(taxonomies, taxonomy)
	}
	return taxonomies, total, nil
}
