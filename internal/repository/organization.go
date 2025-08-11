package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// organizationRepository implements the OrganizationRepository interface using nested set model
// This struct provides concrete implementations for organization management operations
// using PostgreSQL for persistence and nested set for hierarchical tree structure.
type organizationRepository struct {
	db        *pgxpool.Pool               // PostgreSQL connection pool for database operations
	nestedSet *nestedset.NestedSetManager // Manager for nested set tree operations
}

// NewOrganizationRepository creates a new organization repository instance
// This constructor function initializes the repository with the required dependencies
// including the nested set manager for tree structure operations.
//
// Parameters:
//   - db: PostgreSQL connection pool for database operations
//
// Returns:
//   - repositories.OrganizationRepository: interface implementation for organization management
func NewOrganizationRepository(db *pgxpool.Pool) repositories.OrganizationRepository {
	return &organizationRepository{
		db:        db,
		nestedSet: nestedset.NewNestedSetManager(db),
	}
}

// getCacheKey generates a consistent cache key for organization operations
// This helper method creates standardized cache keys for various repository operations
// to ensure consistent caching behavior across the application.
//
// Parameters:
//   - operation: string identifier for the operation type
//   - params: variadic parameters to include in the cache key
//
// Returns:
//   - string: formatted cache key for the operation
func (r *organizationRepository) getCacheKey(operation string, params ...interface{}) string {
	key := fmt.Sprintf("organization:%s", operation)
	for _, param := range params {
		key += fmt.Sprintf(":%v", param)
	}
	return key
}

// getFromCache retrieves data from memory cache (simplified version without Redis)
func (r *organizationRepository) getFromCache(ctx context.Context, key string, dest interface{}) bool {
	// For now, return false to indicate no cache hit
	// In a real implementation, you would use Redis or in-memory cache
	return false
}

// setCache stores data in cache (simplified version without Redis)
func (r *organizationRepository) setCache(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	// For now, do nothing
	// In a real implementation, you would store in Redis or in-memory cache
}

// invalidateCache removes cached data for organization operations
func (r *organizationRepository) invalidateCache(ctx context.Context, pattern string) {
	// For now, do nothing
	// In a real implementation, you would invalidate Redis or in-memory cache
}

// Create adds a new organization to the organizations table with nested set positioning
// This method calculates the appropriate tree position using nested set values
// and inserts the organization record with all required fields including tree structure.
//
// Parameters:
//   - ctx: context for the database operation
//   - organization: pointer to the organization entity to create
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *organizationRepository) Create(ctx context.Context, organization *entities.Organization) error {
	// Calculate nested set values using the shared manager
	values, err := r.nestedSet.CreateNode(ctx, "organizations", organization.ParentID, 1)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign computed nested set values to the entity
	organization.RecordLeft = &values.Left
	organization.RecordRight = &values.Right
	organization.RecordDepth = &values.Depth
	organization.RecordOrdering = &values.Ordering

	// Insert the new organization
	query := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	parentIDStr := ""
	if organization.ParentID != nil {
		parentIDStr = organization.ParentID.String()
	}

	_, err = r.db.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description, organization.Code,
		organization.Type, organization.Status, parentIDStr,
		organization.RecordLeft, organization.RecordRight, organization.RecordDepth, organization.RecordOrdering,
		organization.CreatedAt, organization.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert organization: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("by_id", id.String())
	var organization entities.Organization
	if r.getFromCache(ctx, cacheKey, &organization) {
		return &organization, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL
	`

	var parentIDStr *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&organization.ID, &organization.Name, &organization.Description, &organization.Code,
		&organization.Type, &organization.Status, &parentIDStr,
		&organization.RecordLeft, &organization.RecordRight, &organization.RecordDepth, &organization.RecordOrdering,
		&organization.CreatedAt, &organization.UpdatedAt, &organization.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get organization by ID: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			organization.ParentID = &parentID
		}
	}

	// Cache the result
	r.setCache(ctx, cacheKey, organization, 5*time.Minute)

	return &organization, nil
}

func (r *organizationRepository) GetByCode(ctx context.Context, code string) (*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("by_code", code)
	var organization entities.Organization
	if r.getFromCache(ctx, cacheKey, &organization) {
		return &organization, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE code = $1 AND deleted_at IS NULL
	`

	var parentIDStr *string

	err := r.db.QueryRow(ctx, query, code).Scan(
		&organization.ID, &organization.Name, &organization.Description, &organization.Code,
		&organization.Type, &organization.Status, &parentIDStr,
		&organization.RecordLeft, &organization.RecordRight, &organization.RecordDepth, &organization.RecordOrdering,
		&organization.CreatedAt, &organization.UpdatedAt, &organization.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get organization by code: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			organization.ParentID = &parentID
		}
	}

	// Cache the result
	r.setCache(ctx, cacheKey, organization, 5*time.Minute)

	return &organization, nil
}

func (r *organizationRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Organization, error) {
	// Try cache first for small result sets
	cacheKey := r.getCacheKey("all", limit, offset)
	var organizations []*entities.Organization
	if limit <= 100 && r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0, limit)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache small result sets
	if limit <= 100 {
		r.setCache(ctx, cacheKey, organizations, 2*time.Minute)
	}

	return organizations, nil
}

func (r *organizationRepository) Update(ctx context.Context, organization *entities.Organization) error {
	// For nested set, we need to handle parent changes carefully
	// This is a simplified update that doesn't change the tree structure
	query := `
		UPDATE organizations 
		SET name = $2, description = $3, code = $4, type = $5, status = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description, organization.Code,
		organization.Type, organization.Status, organization.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete - mark as deleted
	query := `
		UPDATE organizations 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

func (r *organizationRepository) GetRoots(ctx context.Context) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("roots")
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root organizations: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache root organizations for longer as they change less frequently
	r.setCache(ctx, cacheKey, organizations, 10*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("children", parentID.String())
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache children for medium duration
	r.setCache(ctx, cacheKey, organizations, 5*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetDescendants(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("descendants", parentID.String())
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations, target_node
		WHERE record_left > target_node.record_left 
		AND record_right < target_node.record_right 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache descendants for medium duration
	r.setCache(ctx, cacheKey, organizations, 5*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetAncestors(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("ancestors", organizationID.String())
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations, target_node
		WHERE record_left < target_node.record_left 
		AND record_right > target_node.record_right 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache ancestors for medium duration
	r.setCache(ctx, cacheKey, organizations, 5*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetSiblings(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("siblings", organizationID.String())
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT parent_id FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT t1.id, t1.name, t1.description, t1.code, t1.type, t1.status, t1.parent_id, 
		       t1.record_left, t1.record_right, t1.record_depth, t1.record_ordering,
		       t1.created_at, t1.updated_at, t1.deleted_at
		FROM organizations t1, target_node
		WHERE t1.parent_id = target_node.parent_id AND t1.id != $1 AND t1.deleted_at IS NULL
		ORDER BY t1.record_left ASC
	`

	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache siblings for medium duration
	r.setCache(ctx, cacheKey, organizations, 5*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("path", organizationID.String())
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations, target_node
		WHERE record_left <= target_node.record_left 
		AND record_right >= target_node.record_right 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get path: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache path for medium duration
	r.setCache(ctx, cacheKey, organizations, 5*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetTree(ctx context.Context) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("tree")
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache tree for longer as it's expensive to compute
	r.setCache(ctx, cacheKey, organizations, 15*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) GetSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error) {
	// Try cache first
	cacheKey := r.getCacheKey("subtree", rootID.String())
	var organizations []*entities.Organization
	if r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations, target_node
		WHERE record_left >= target_node.record_left 
		AND record_right <= target_node.record_right 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtree: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache subtree for medium duration
	r.setCache(ctx, cacheKey, organizations, 5*time.Minute)

	return organizations, nil
}

func (r *organizationRepository) AddChild(ctx context.Context, parentID, childID uuid.UUID) error {
	// Use the nested set manager to properly restructure the tree
	err := r.nestedSet.MoveSubtree(ctx, "organizations", childID, parentID)
	if err != nil {
		return fmt.Errorf("failed to add child using nested set: %w", err)
	}

	// Update the updated_at timestamp
	query := `UPDATE organizations SET updated_at = NOW() WHERE id = $1`
	_, err = r.db.Exec(ctx, query, childID)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

func (r *organizationRepository) MoveSubtree(ctx context.Context, organizationID, newParentID uuid.UUID) error {
	// Use the nested set manager to properly restructure the tree
	err := r.nestedSet.MoveSubtree(ctx, "organizations", organizationID, newParentID)
	if err != nil {
		return fmt.Errorf("failed to move subtree using nested set: %w", err)
	}

	// Update the updated_at timestamp
	query := `UPDATE organizations SET updated_at = NOW() WHERE id = $1`
	_, err = r.db.Exec(ctx, query, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

func (r *organizationRepository) DeleteSubtree(ctx context.Context, organizationID uuid.UUID) error {
	// Use the nested set manager to properly handle subtree deletion
	err := r.nestedSet.DeleteSubtree(ctx, "organizations", organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete subtree using nested set: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

func (r *organizationRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	// Try cache first for small result sets
	cacheKey := r.getCacheKey("search", query, limit, offset)
	var organizations []*entities.Organization
	if limit <= 100 && r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	searchQuery := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations 
		WHERE (name ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0, limit)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache small result sets
	if limit <= 100 {
		r.setCache(ctx, cacheKey, organizations, 2*time.Minute)
	}

	return organizations, nil
}

func (r *organizationRepository) GetByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error) {
	// Try cache first for small result sets
	cacheKey := r.getCacheKey("by_status", status, limit, offset)
	var organizations []*entities.Organization
	if limit <= 100 && r.getFromCache(ctx, cacheKey, &organizations) {
		return organizations, nil
	}

	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE status = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations by status: %w", err)
	}
	defer rows.Close()

	organizations = make([]*entities.Organization, 0, limit)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	// Cache small result sets
	if limit <= 100 {
		r.setCache(ctx, cacheKey, organizations, 2*time.Minute)
	}

	return organizations, nil
}

func (r *organizationRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE code = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}
	return exists, nil
}

func (r *organizationRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}
	return exists, nil
}

func (r *organizationRepository) IsDescendant(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM organizations o1, organizations o2
			WHERE o1.id = $1 AND o2.id = $2 
			AND o1.record_left < o2.record_left 
			AND o1.record_right > o2.record_right
			AND o1.deleted_at IS NULL AND o2.deleted_at IS NULL
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, ancestorID, descendantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check descendant relationship: %w", err)
	}
	return exists, nil
}

func (r *organizationRepository) IsAncestor(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	return r.IsDescendant(ctx, ancestorID, descendantID)
}

func (r *organizationRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations: %w", err)
	}
	return count, nil
}

func (r *organizationRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
	var sqlQuery string
	var args []interface{}

	if query == "" {
		sqlQuery = `SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL`
	} else {
		sqlQuery = `SELECT COUNT(*) FROM organizations WHERE (name ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL`
		searchTerm := "%" + query + "%"
		args = []interface{}{searchTerm}
	}

	var count int64
	err := r.db.QueryRow(ctx, sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations with search: %w", err)
	}
	return count, nil
}

func (r *organizationRepository) CountByStatus(ctx context.Context, status entities.OrganizationStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE status = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations by status: %w", err)
	}
	return count, nil
}

func (r *organizationRepository) CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE parent_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, parentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count children: %w", err)
	}
	return count, nil
}

func (r *organizationRepository) CountDescendants(ctx context.Context, parentID uuid.UUID) (int64, error) {
	// Use CTE for better performance
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT COUNT(*) FROM organizations, target_node
		WHERE record_left > target_node.record_left 
		AND record_right < target_node.record_right 
		AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, parentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count descendants: %w", err)
	}
	return count, nil
}

// scanOrganizationRow is a helper function to scan an organization row from database
func (r *organizationRepository) scanOrganizationRow(rows pgx.Rows) (*entities.Organization, error) {
	var organization entities.Organization
	var parentIDStr *string

	err := rows.Scan(
		&organization.ID, &organization.Name, &organization.Description, &organization.Code,
		&organization.Type, &organization.Status, &parentIDStr,
		&organization.RecordLeft, &organization.RecordRight, &organization.RecordDepth, &organization.RecordOrdering,
		&organization.CreatedAt, &organization.UpdatedAt, &organization.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan organization row: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			organization.ParentID = &parentID
		}
	}

	return &organization, nil
}

// Advanced Nested Set Operations

// ValidateTree validates the nested set tree structure and returns any inconsistencies
func (r *organizationRepository) ValidateTree(ctx context.Context) ([]string, error) {
	errors, err := r.nestedSet.ValidateTree(ctx, "organizations")
	if err != nil {
		return nil, fmt.Errorf("failed to validate tree: %w", err)
	}
	return errors, nil
}

// RebuildTree rebuilds the entire nested set tree structure from parent_id relationships
func (r *organizationRepository) RebuildTree(ctx context.Context) error {
	err := r.nestedSet.RebuildTree(ctx, "organizations")
	if err != nil {
		return fmt.Errorf("failed to rebuild tree: %w", err)
	}

	// Invalidate all caches after tree rebuild
	r.invalidateCache(ctx, "organization:*")

	return nil
}

// GetTreeStatistics returns comprehensive statistics about the tree structure
func (r *organizationRepository) GetTreeStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats, err := r.nestedSet.GetTreeStatistics(ctx, "organizations")
	if err != nil {
		return nil, fmt.Errorf("failed to get tree statistics: %w", err)
	}
	return stats, nil
}

// GetTreeHeight returns the maximum depth of the tree
func (r *organizationRepository) GetTreeHeight(ctx context.Context) (uint64, error) {
	height, err := r.nestedSet.GetTreeHeight(ctx, "organizations")
	if err != nil {
		return 0, fmt.Errorf("failed to get tree height: %w", err)
	}
	return height, nil
}

// GetLevelWidth returns the number of nodes at a specific depth level
func (r *organizationRepository) GetLevelWidth(ctx context.Context, level uint64) (int64, error) {
	width, err := r.nestedSet.GetLevelWidth(ctx, "organizations", level)
	if err != nil {
		return 0, fmt.Errorf("failed to get level width: %w", err)
	}
	return width, nil
}

// GetSubtreeSize returns the total number of nodes in a subtree
func (r *organizationRepository) GetSubtreeSize(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	size, err := r.nestedSet.GetSubtreeSize(ctx, "organizations", organizationID)
	if err != nil {
		return 0, fmt.Errorf("failed to get subtree size: %w", err)
	}
	return size, nil
}

// InsertBetween inserts a new organization between two existing siblings
func (r *organizationRepository) InsertBetween(ctx context.Context, organization *entities.Organization, leftSiblingID, rightSiblingID *uuid.UUID) error {
	// Calculate nested set values for insertion between siblings
	var leftSiblingRight, rightSiblingLeft uint64

	if leftSiblingID != nil {
		err := r.db.QueryRow(ctx, `
			SELECT record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		`, leftSiblingID.String()).Scan(&leftSiblingRight)
		if err != nil {
			return fmt.Errorf("failed to get left sibling info: %w", err)
		}
	} else {
		leftSiblingRight = 0
	}

	if rightSiblingID != nil {
		err := r.db.QueryRow(ctx, `
			SELECT record_left FROM organizations WHERE id = $1 AND deleted_at IS NULL
		`, rightSiblingID.String()).Scan(&rightSiblingLeft)
		if err != nil {
			return fmt.Errorf("failed to get right sibling info: %w", err)
		}
	} else {
		// Get the maximum right value if inserting at the end
		err := r.db.QueryRow(ctx, `
			SELECT COALESCE(MAX(record_right), 0) FROM organizations WHERE deleted_at IS NULL
		`).Scan(&rightSiblingLeft)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
		rightSiblingLeft++
	}

	// Calculate new nested set values
	newLeft := leftSiblingRight + 1
	newRight := rightSiblingLeft - 1

	// Ensure we have space between siblings
	if newLeft >= newRight {
		// Need to shift existing nodes to make space
		shiftAmount := uint64(2)
		_, err := r.db.Exec(ctx, `
			UPDATE organizations 
			SET record_left = CASE 
				WHEN record_left > $1 THEN record_left + $2 
				ELSE record_left 
			END,
			record_right = CASE 
				WHEN record_right >= $1 THEN record_right + $2 
				ELSE record_right 
			END
			WHERE deleted_at IS NULL
		`, leftSiblingRight, shiftAmount)
		if err != nil {
			return fmt.Errorf("failed to shift nodes: %w", err)
		}
		newLeft = leftSiblingRight + 1
		newRight = newLeft + 1
	}

	// Assign computed nested set values
	organization.RecordLeft = &newLeft
	organization.RecordRight = &newRight
	organization.RecordDepth = &[]uint64{0}[0] // Root level
	organization.RecordOrdering = &[]uint64{newLeft}[0]

	// Insert the organization
	query := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	parentIDStr := ""
	if organization.ParentID != nil {
		parentIDStr = organization.ParentID.String()
	}

	_, err := r.db.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description, organization.Code,
		organization.Type, organization.Status, parentIDStr,
		organization.RecordLeft, organization.RecordRight, organization.RecordDepth, organization.RecordOrdering,
		organization.CreatedAt, organization.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert organization: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

// SwapPositions swaps the positions of two organizations in the tree
func (r *organizationRepository) SwapPositions(ctx context.Context, org1ID, org2ID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current positions
	var org1Left, org1Right, org2Left, org2Right uint64
	err = tx.QueryRow(ctx, `
		SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
	`, org1ID.String()).Scan(&org1Left, &org1Right)
	if err != nil {
		return fmt.Errorf("failed to get first organization info: %w", err)
	}

	err = tx.QueryRow(ctx, `
		SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
	`, org2ID.String()).Scan(&org2Left, &org2Right)
	if err != nil {
		return fmt.Errorf("failed to get second organization info: %w", err)
	}

	// Check if organizations are in the same subtree (can't swap if one is ancestor of the other)
	if (org1Left < org2Left && org1Right > org2Right) || (org2Left < org1Left && org2Right > org1Right) {
		return fmt.Errorf("cannot swap positions: organizations are in ancestor-descendant relationship")
	}

	// Use temporary values to avoid conflicts during swap
	tempLeft := uint64(999999999) // Large temporary value

	// Move first organization to temporary position
	_, err = tx.Exec(ctx, `
		UPDATE organizations 
		SET record_left = $1, record_right = $2
		WHERE id = $3
	`, tempLeft, tempLeft+org1Right-org1Left, org1ID.String())
	if err != nil {
		return fmt.Errorf("failed to move first organization to temp position: %w", err)
	}

	// Move second organization to first organization's position
	_, err = tx.Exec(ctx, `
		UPDATE organizations 
		SET record_left = $1, record_right = $2
		WHERE id = $3
	`, org1Left, org1Right, org2ID.String())
	if err != nil {
		return fmt.Errorf("failed to move second organization: %w", err)
	}

	// Move first organization to second organization's position
	_, err = tx.Exec(ctx, `
		UPDATE organizations 
		SET record_left = $1, record_right = $2
		WHERE id = $3
	`, org2Left, org2Right, org1ID.String())
	if err != nil {
		return fmt.Errorf("failed to move first organization to final position: %w", err)
	}

	// Update timestamps
	_, err = tx.Exec(ctx, `
		UPDATE organizations SET updated_at = NOW() WHERE id IN ($1, $2)
	`, org1ID.String(), org2ID.String())
	if err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate relevant caches
	r.invalidateCache(ctx, "organization:*")

	return nil
}

// GetLeafNodes returns all leaf nodes (nodes without children) in the tree
func (r *organizationRepository) GetLeafNodes(ctx context.Context) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations o1
		WHERE deleted_at IS NULL 
		AND NOT EXISTS (
			SELECT 1 FROM organizations o2 
			WHERE o2.parent_id = o1.id AND o2.deleted_at IS NULL
		)
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf nodes: %w", err)
	}
	defer rows.Close()

	organizations := make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	return organizations, nil
}

// GetInternalNodes returns all internal nodes (nodes with children) in the tree
func (r *organizationRepository) GetInternalNodes(ctx context.Context) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations o1
		WHERE deleted_at IS NULL 
		AND EXISTS (
			SELECT 1 FROM organizations o2 
			WHERE o2.parent_id = o1.id AND o2.deleted_at IS NULL
		)
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal nodes: %w", err)
	}
	defer rows.Close()

	organizations := make([]*entities.Organization, 0)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}

	return organizations, nil
}

// Batch operations for nested set

// BatchMoveSubtrees moves multiple subtrees in a single transaction for better performance
func (r *organizationRepository) BatchMoveSubtrees(ctx context.Context, moves []struct {
	OrganizationID uuid.UUID
	NewParentID    uuid.UUID
}) error {
	if len(moves) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, move := range moves {
		// Use the nested set manager for each move
		err := r.nestedSet.MoveSubtree(ctx, "organizations", move.OrganizationID, move.NewParentID)
		if err != nil {
			return fmt.Errorf("failed to move subtree %s: %w", move.OrganizationID.String(), err)
		}

		// Update timestamp
		_, err = tx.Exec(ctx, `UPDATE organizations SET updated_at = NOW() WHERE id = $1`, move.OrganizationID.String())
		if err != nil {
			return fmt.Errorf("failed to update timestamp for %s: %w", move.OrganizationID.String(), err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit batch move transaction: %w", err)
	}

	// Invalidate all caches after batch operation
	r.invalidateCache(ctx, "organization:*")

	return nil
}

// BatchInsertBetween inserts multiple organizations between siblings in a single transaction
func (r *organizationRepository) BatchInsertBetween(ctx context.Context, insertions []struct {
	Organization   *entities.Organization
	LeftSiblingID  *uuid.UUID
	RightSiblingID *uuid.UUID
}) error {
	if len(insertions) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, insertion := range insertions {
		err := r.insertBetweenInTx(ctx, tx, insertion.Organization, insertion.LeftSiblingID, insertion.RightSiblingID)
		if err != nil {
			return fmt.Errorf("failed to insert organization %s: %w", insertion.Organization.ID.String(), err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit batch insert transaction: %w", err)
	}

	// Invalidate all caches after batch operation
	r.invalidateCache(ctx, "organization:*")

	return nil
}

// insertBetweenInTx is a helper method for batch operations
func (r *organizationRepository) insertBetweenInTx(ctx context.Context, tx pgx.Tx, organization *entities.Organization, leftSiblingID, rightSiblingID *uuid.UUID) error {
	// Calculate nested set values for insertion between siblings
	var leftSiblingRight, rightSiblingLeft uint64

	if leftSiblingID != nil {
		err := tx.QueryRow(ctx, `
			SELECT record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		`, leftSiblingID.String()).Scan(&leftSiblingRight)
		if err != nil {
			return fmt.Errorf("failed to get left sibling info: %w", err)
		}
	} else {
		leftSiblingRight = 0
	}

	if rightSiblingID != nil {
		err := tx.QueryRow(ctx, `
			SELECT record_left FROM organizations WHERE id = $1 AND deleted_at IS NULL
		`, rightSiblingID.String()).Scan(&rightSiblingLeft)
		if err != nil {
			return fmt.Errorf("failed to get right sibling info: %w", err)
		}
	} else {
		// Get the maximum right value if inserting at the end
		err := tx.QueryRow(ctx, `
			SELECT COALESCE(MAX(record_right), 0) FROM organizations WHERE deleted_at IS NULL
		`).Scan(&rightSiblingLeft)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
		rightSiblingLeft++
	}

	// Calculate new nested set values
	newLeft := leftSiblingRight + 1
	newRight := rightSiblingLeft - 1

	// Ensure we have space between siblings
	if newLeft >= newRight {
		// Need to shift existing nodes to make space
		shiftAmount := uint64(2)
		_, err := tx.Exec(ctx, `
			UPDATE organizations 
			SET record_left = CASE 
				WHEN record_left > $1 THEN record_left + $2 
				ELSE record_left 
			END,
			record_right = CASE 
				WHEN record_right >= $1 THEN record_right + $2 
				ELSE record_right 
			END
			WHERE deleted_at IS NULL
		`, leftSiblingRight, shiftAmount)
		if err != nil {
			return fmt.Errorf("failed to shift nodes: %w", err)
		}
		newLeft = leftSiblingRight + 1
		newRight = newLeft + 1
	}

	// Assign computed nested set values
	organization.RecordLeft = &newLeft
	organization.RecordRight = &newRight
	organization.RecordDepth = &[]uint64{0}[0] // Root level
	organization.RecordOrdering = &[]uint64{newLeft}[0]

	// Insert the organization
	query := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	parentIDStr := ""
	if organization.ParentID != nil {
		parentIDStr = organization.ParentID.String()
	}

	_, err := tx.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description, organization.Code,
		organization.Type, organization.Status, parentIDStr,
		organization.RecordLeft, organization.RecordRight, organization.RecordDepth, organization.RecordOrdering,
		organization.CreatedAt, organization.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert organization: %w", err)
	}

	return nil
}

// Tree optimization and maintenance

// OptimizeTree performs tree optimization by compacting nested set values
func (r *organizationRepository) OptimizeTree(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get all organizations ordered by their current left values
	rows, err := tx.Query(ctx, `
		SELECT id, record_left, record_right, record_depth 
		FROM organizations 
		WHERE deleted_at IS NULL 
		ORDER BY record_left ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to get organizations for optimization: %w", err)
	}
	defer rows.Close()

	var organizations []struct {
		ID    uuid.UUID
		Left  uint64
		Right uint64
		Depth uint64
	}

	for rows.Next() {
		var org struct {
			ID    uuid.UUID
			Left  uint64
			Right uint64
			Depth uint64
		}
		err := rows.Scan(&org.ID, &org.Left, &org.Right, &org.Depth)
		if err != nil {
			return fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, org)
	}

	// Recalculate nested set values with optimal spacing
	newLeft := uint64(1)
	for _, org := range organizations {
		subtreeSize := org.Right - org.Left + 1
		newRight := newLeft + subtreeSize - 1

		// Update the organization with new values
		_, err := tx.Exec(ctx, `
			UPDATE organizations 
			SET record_left = $1, record_right = $2, record_depth = $3
			WHERE id = $4
		`, newLeft, newRight, org.Depth, org.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update organization %s: %w", org.ID.String(), err)
		}

		newLeft = newRight + 1
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit optimization transaction: %w", err)
	}

	// Invalidate all caches after optimization
	r.invalidateCache(ctx, "organization:*")

	return nil
}

// GetTreePerformanceMetrics returns performance metrics for tree operations
func (r *organizationRepository) GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get tree size
	count, err := r.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree size: %w", err)
	}
	metrics["total_nodes"] = count

	// Get tree height
	height, err := r.GetTreeHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree height: %w", err)
	}
	metrics["tree_height"] = height

	// Get average children per node
	var avgChildren float64
	err = r.db.QueryRow(ctx, `
		SELECT AVG(child_count) FROM (
			SELECT COUNT(*) as child_count 
			FROM organizations o1
			LEFT JOIN organizations o2 ON o2.parent_id = o1.id AND o2.deleted_at IS NULL
			WHERE o1.deleted_at IS NULL
			GROUP BY o1.id
		) as children_stats
	`).Scan(&avgChildren)
	if err != nil {
		avgChildren = 0
	}
	metrics["average_children_per_node"] = avgChildren

	// Get leaf node percentage
	var leafCount int64
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM organizations o1
		WHERE o1.deleted_at IS NULL 
		AND NOT EXISTS (
			SELECT 1 FROM organizations o2 
			WHERE o2.parent_id = o1.id AND o2.deleted_at IS NULL
		)
	`).Scan(&leafCount)
	if err != nil {
		leafCount = 0
	}
	leafPercentage := float64(0)
	if count > 0 {
		leafPercentage = float64(leafCount) / float64(count) * 100
	}
	metrics["leaf_node_percentage"] = leafPercentage

	// Get tree balance (standard deviation of depths)
	var depthStdDev float64
	err = r.db.QueryRow(ctx, `
		SELECT STDDEV(record_depth) FROM organizations WHERE deleted_at IS NULL
	`).Scan(&depthStdDev)
	if err != nil {
		depthStdDev = 0
	}
	metrics["depth_standard_deviation"] = depthStdDev

	return metrics, nil
}

// ValidateTreeIntegrity performs comprehensive tree integrity validation
func (r *organizationRepository) ValidateTreeIntegrity(ctx context.Context) ([]string, error) {
	var errors []string

	// Check for orphaned nodes (nodes with parent_id pointing to non-existent nodes)
	var orphanedCount int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM organizations o1
		WHERE o1.deleted_at IS NULL 
		AND o1.parent_id IS NOT NULL 
		AND NOT EXISTS (
			SELECT 1 FROM organizations o2 
			WHERE o2.id::text = o1.parent_id AND o2.deleted_at IS NULL
		)
	`).Scan(&orphanedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check for orphaned nodes: %w", err)
	}
	if orphanedCount > 0 {
		errors = append(errors, fmt.Sprintf("Found %d orphaned nodes", orphanedCount))
	}

	// Check for circular references
	var circularCount int64
	err = r.db.QueryRow(ctx, `
		WITH RECURSIVE cycle_check AS (
			SELECT id, parent_id, ARRAY[id] as path
			FROM organizations 
			WHERE deleted_at IS NULL AND parent_id IS NOT NULL
			UNION ALL
			SELECT o.id, o.parent_id, cc.path || o.id
			FROM organizations o
			JOIN cycle_check cc ON o.id::text = cc.parent_id
			WHERE o.deleted_at IS NULL 
			AND NOT (o.id = ANY(cc.path))
		)
		SELECT COUNT(*) FROM cycle_check
	`).Scan(&circularCount)
	if err != nil {
		// This query might fail if there are cycles, which is what we're checking for
		circularCount = 1
	}
	if circularCount > 0 {
		errors = append(errors, "Found circular references in tree structure")
	}

	// Check nested set consistency
	var inconsistentCount int64
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM organizations o1
		WHERE o1.deleted_at IS NULL 
		AND EXISTS (
			SELECT 1 FROM organizations o2
			WHERE o2.deleted_at IS NULL 
			AND o2.id != o1.id
			AND (
				(o2.record_left >= o1.record_left AND o2.record_left <= o1.record_right) OR
				(o2.record_right >= o1.record_left AND o2.record_right <= o1.record_right)
			)
			AND NOT (
				o2.record_left < o1.record_left AND o2.record_right > o1.record_right
			)
		)
	`).Scan(&inconsistentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check nested set consistency: %w", err)
	}
	if inconsistentCount > 0 {
		errors = append(errors, fmt.Sprintf("Found %d nodes with inconsistent nested set values", inconsistentCount))
	}

	return errors, nil
}
