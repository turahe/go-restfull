package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// TaxonomyRepository provides the concrete implementation for taxonomy management operations
// This struct handles all taxonomy-related database operations including CRUD operations,
// hierarchical tree management using nested sets, and Redis caching for performance.
type TaxonomyRepository struct {
	pgxPool     *pgxpool.Pool               // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable               // Redis client for caching operations
	nestedSet   *nestedset.NestedSetManager // Manager for nested set tree operations
}

// NewTaxonomyRepository creates a new instance of TaxonomyRepository
// This constructor function initializes the repository with the required dependencies
// including the nested set manager for tree structure operations and Redis for caching.
//
// Parameters:
//   - db: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - repositories.TaxonomyRepository: interface implementation for taxonomy management
func NewTaxonomyRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TaxonomyRepository {
	return &TaxonomyRepository{
		pgxPool:     db,
		redisClient: redisClient,
		nestedSet:   nestedset.NewNestedSetManager(db),
	}
}

// getCacheKey generates a consistent cache key for taxonomy operations
// This helper method creates standardized cache keys for various repository operations
// to ensure consistent caching behavior across the application.
//
// Parameters:
//   - operation: string identifier for the operation type
//   - params: variadic parameters to include in the cache key
//
// Returns:
//   - string: formatted cache key for the operation
func (r *TaxonomyRepository) getCacheKey(operation string, params ...interface{}) string {
	key := fmt.Sprintf("taxonomy:%s", operation)
	for _, param := range params {
		key += fmt.Sprintf(":%v", param)
	}
	return key
}

// getFromCache retrieves data from Redis cache
func (r *TaxonomyRepository) getFromCache(ctx context.Context, key string, dest interface{}) bool {
	if r.redisClient == nil {
		return false
	}

	data, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return false
	}

	return true
}

// setCache stores data in Redis cache with TTL
func (r *TaxonomyRepository) setCache(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	if r.redisClient == nil {
		return
	}

	if jsonData, err := json.Marshal(data); err == nil {
		r.redisClient.Set(ctx, key, jsonData, ttl)
	}
}

// invalidateCache removes cached data for taxonomy operations
func (r *TaxonomyRepository) invalidateCache(ctx context.Context, pattern string) {
	if r.redisClient == nil {
		return
	}

	keys, err := r.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}

	if len(keys) > 0 {
		r.redisClient.Del(ctx, keys...)
	}
}

// Create adds a new taxonomy to the taxonomies table with nested set positioning
// This method calculates the appropriate tree position using nested set values
// and inserts the taxonomy record with all required fields including tree structure.
//
// Parameters:
//   - ctx: context for the database operation
//   - taxonomy: pointer to the taxonomy entity to create
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TaxonomyRepository) Create(ctx context.Context, taxonomy *entities.Taxonomy) error {
	// Calculate nested set values using the shared manager
	values, err := r.nestedSet.CreateNode(ctx, "taxonomies", taxonomy.ParentID, 1)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	taxonomy.RecordLeft = &values.Left
	taxonomy.RecordRight = &values.Right
	taxonomy.RecordDepth = &values.Depth

	// Insert the new taxonomy
	query := `
		INSERT INTO taxonomies (id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	parentIDStr := ""
	if taxonomy.ParentID != nil {
		parentIDStr = taxonomy.ParentID.String()
	}

	_, err = r.pgxPool.Exec(ctx, query,
		taxonomy.ID.String(), taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description,
		parentIDStr, taxonomy.RecordLeft, taxonomy.RecordRight, taxonomy.RecordDepth,
		taxonomy.CreatedAt, taxonomy.UpdatedAt, "", "", // created_by, updated_by
	)

	if err != nil {
		return fmt.Errorf("failed to insert taxonomy: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "taxonomy:*")

	return nil
}

func (r *TaxonomyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("by_id", id.String())
	var taxonomy entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomy) {
		return &taxonomy, nil
	}

	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
	`

	var parentIDStr *string
	var createdBy, updatedBy, deletedBy string

	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&parentIDStr, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &createdBy, &updatedBy, &deletedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get taxonomy by ID: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	// Cache the result
	r.setCache(ctx, cacheKey, taxonomy, 5*time.Minute)

	return &taxonomy, nil
}

func (r *TaxonomyRepository) GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("by_slug", slug)
	var taxonomy entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomy) {
		return &taxonomy, nil
	}

	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL
	`

	var parentIDStr *string
	var createdBy, updatedBy, deletedBy string

	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&parentIDStr, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &createdBy, &updatedBy, &deletedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get taxonomy by slug: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	// Cache the result
	r.setCache(ctx, cacheKey, taxonomy, 5*time.Minute)

	return &taxonomy, nil
}

func (r *TaxonomyRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	// Try cache first for small result sets
	cacheKey := r.getCacheKey("all", limit, offset)
	var taxonomies []*entities.Taxonomy
	if limit <= 100 && r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get taxonomies: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0, limit)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache small result sets
	if limit <= 100 {
		r.setCache(ctx, cacheKey, taxonomies, 2*time.Minute)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetAllWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	// Try cache first for small result sets
	cacheKey := r.getCacheKey("search", query, limit, offset)
	var taxonomies []*entities.Taxonomy
	if limit <= 100 && r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	var sqlQuery string
	var args []interface{}

	if query == "" {
		sqlQuery = `
			SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
			FROM taxonomies WHERE deleted_at IS NULL
			ORDER BY record_left ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	} else {
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
		return nil, fmt.Errorf("failed to search taxonomies: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0, limit)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache small result sets
	if limit <= 100 {
		r.setCache(ctx, cacheKey, taxonomies, 2*time.Minute)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("roots")
	var taxonomies []*entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root taxonomies: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache root taxonomies for longer as they change less frequently
	r.setCache(ctx, cacheKey, taxonomies, 10*time.Minute)

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("children", parentID.String())
	var taxonomies []*entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, parentID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache children for medium duration
	r.setCache(ctx, cacheKey, taxonomies, 5*time.Minute)

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("hierarchy")
	var taxonomies []*entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	query := `
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies WHERE deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache hierarchy for longer as it's expensive to compute
	r.setCache(ctx, cacheKey, taxonomies, 15*time.Minute)

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("descendants", id.String())
	var taxonomies []*entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies, target_node
		WHERE record_left > target_node.record_left 
		AND record_right < target_node.record_right 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache descendants for medium duration
	r.setCache(ctx, cacheKey, taxonomies, 5*time.Minute)

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("ancestors", id.String())
	var taxonomies []*entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT id, name, slug, code, description, parent_id, record_left, record_right, record_depth, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by
		FROM taxonomies, target_node
		WHERE record_left < target_node.record_left 
		AND record_right > target_node.record_right 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache ancestors for medium duration
	r.setCache(ctx, cacheKey, taxonomies, 5*time.Minute)

	return taxonomies, nil
}

func (r *TaxonomyRepository) GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	// Try cache first
	cacheKey := r.getCacheKey("siblings", id.String())
	var taxonomies []*entities.Taxonomy
	if r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT parent_id FROM taxonomies WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT t1.id, t1.name, t1.slug, t1.code, t1.description, t1.parent_id, t1.record_left, t1.record_right, t1.record_depth, t1.created_at, t1.updated_at, t1.deleted_at, t1.created_by, t1.updated_by, t1.deleted_by
		FROM taxonomies t1, target_node
		WHERE t1.parent_id = target_node.parent_id AND t1.id != $1 AND t1.deleted_at IS NULL
		ORDER BY t1.record_left ASC
	`

	rows, err := r.pgxPool.Query(ctx, query, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache siblings for medium duration
	r.setCache(ctx, cacheKey, taxonomies, 5*time.Minute)

	return taxonomies, nil
}

func (r *TaxonomyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	// Try cache first for small result sets
	cacheKey := r.getCacheKey("search", query, limit, offset)
	var taxonomies []*entities.Taxonomy
	if limit <= 100 && r.getFromCache(ctx, cacheKey, &taxonomies) {
		return taxonomies, nil
	}

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
		return nil, fmt.Errorf("failed to search taxonomies: %w", err)
	}
	defer rows.Close()

	taxonomies = make([]*entities.Taxonomy, 0, limit)
	for rows.Next() {
		taxonomy, err := r.scanTaxonomyRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
		}
		taxonomies = append(taxonomies, taxonomy)
	}

	// Cache small result sets
	if limit <= 100 {
		r.setCache(ctx, cacheKey, taxonomies, 2*time.Minute)
	}

	return taxonomies, nil
}

func (r *TaxonomyRepository) Update(ctx context.Context, taxonomy *entities.Taxonomy) error {
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

	if err != nil {
		return fmt.Errorf("failed to update taxonomy: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "taxonomy:*")

	return nil
}

func (r *TaxonomyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete - mark as deleted
	query := `
		UPDATE taxonomies 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pgxPool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete taxonomy: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "taxonomy:*")

	return nil
}

func (r *TaxonomyRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM taxonomies WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check taxonomy existence: %w", err)
	}
	return exists, nil
}

func (r *TaxonomyRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM taxonomies WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count taxonomies: %w", err)
	}
	return count, nil
}

func (r *TaxonomyRepository) CountWithSearch(ctx context.Context, query string) (int64, error) {
	var sqlQuery string
	var args []interface{}

	if query == "" {
		sqlQuery = `SELECT COUNT(*) FROM taxonomies WHERE deleted_at IS NULL`
	} else {
		sqlQuery = `SELECT COUNT(*) FROM taxonomies WHERE (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL`
		searchTerm := "%" + query + "%"
		args = []interface{}{searchTerm}
	}

	var count int64
	err := r.pgxPool.QueryRow(ctx, sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count taxonomies with search: %w", err)
	}
	return count, nil
}

// scanTaxonomyRow is a helper function to scan a taxonomy row from database
func (r *TaxonomyRepository) scanTaxonomyRow(rows pgx.Rows) (*entities.Taxonomy, error) {
	var taxonomy entities.Taxonomy
	var parentIDStr *string
	var createdBy, updatedBy, deletedBy string

	err := rows.Scan(
		&taxonomy.ID, &taxonomy.Name, &taxonomy.Slug, &taxonomy.Code, &taxonomy.Description,
		&parentIDStr, &taxonomy.RecordLeft, &taxonomy.RecordRight, &taxonomy.RecordDepth,
		&taxonomy.CreatedAt, &taxonomy.UpdatedAt, &taxonomy.DeletedAt, &createdBy, &updatedBy, &deletedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan taxonomy row: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	return &taxonomy, nil
}
