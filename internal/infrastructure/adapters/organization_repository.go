package adapters

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

type PostgresOrganizationRepository struct {
	*BaseTransactionalRepository
	db        *pgxpool.Pool
	nestedSet *nestedset.NestedSetManager
}

func NewPostgresOrganizationRepository(db *pgxpool.Pool) repositories.OrganizationRepository {
	return &PostgresOrganizationRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		nestedSet:                   nestedset.NewNestedSetManager(db),
	}
}

func (r *PostgresOrganizationRepository) Create(ctx context.Context, organization *entities.Organization) error {
	// Set current timestamps if not already set
	now := time.Now()
	if organization.CreatedAt.IsZero() {
		organization.CreatedAt = now
	}
	if organization.UpdatedAt.IsZero() {
		organization.UpdatedAt = now
	}

	// Calculate nested set values
	nestedSetValues, err := r.nestedSet.CreateNode(ctx, "organizations", organization.ParentID, int64(1))
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign nested set values to organization entity
	organization.RecordLeft = &nestedSetValues.Left
	organization.RecordRight = &nestedSetValues.Right
	organization.RecordDepth = &nestedSetValues.Depth
	organization.RecordOrdering = &nestedSetValues.Ordering

	query := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

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
	return nil
}

// createOrganizationFallback creates an organization with manual nested set values
// This is used when the nested set manager fails to create the first organization
func (r *PostgresOrganizationRepository) createOrganizationFallback(ctx context.Context, organization *entities.Organization) error {
	// Set current timestamps if not already set
	now := time.Now()
	if organization.CreatedAt.IsZero() {
		organization.CreatedAt = now
	}
	if organization.UpdatedAt.IsZero() {
		organization.UpdatedAt = now
	}

	// For the first organization, set manual nested set values
	var recordLeft, recordRight, recordDepth, recordOrdering int64
	if organization.ParentID == nil {
		// Root organization - start with basic values
		recordLeft = 1
		recordRight = 2
		recordDepth = 0
		recordOrdering = 1
	} else {
		// Child organization - this shouldn't happen in fallback, but handle it
		recordLeft = 3
		recordRight = 4
		recordDepth = 1
		recordOrdering = 1
	}

	organization.RecordLeft = &recordLeft
	organization.RecordRight = &recordRight
	organization.RecordDepth = &recordDepth
	organization.RecordOrdering = &recordOrdering

	query := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	var parentIDStr *string
	if organization.ParentID != nil {
		parentIDStr = func() *string { s := organization.ParentID.String(); return &s }()
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
	return nil
}

func (r *PostgresOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL`
	var organization entities.Organization
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
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			organization.ParentID = &parentID
		}
	}
	return &organization, nil
}

func (r *PostgresOrganizationRepository) GetByCode(ctx context.Context, code string) (*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE code = $1 AND deleted_at IS NULL`
	var organization entities.Organization
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
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			organization.ParentID = &parentID
		}
	}
	return &organization, nil
}

func (r *PostgresOrganizationRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer rows.Close()
	organizations := make([]*entities.Organization, 0, limit)
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) Update(ctx context.Context, organization *entities.Organization) error {
	// Set current timestamp for updated_at
	organization.UpdatedAt = time.Now()

	query := `
		UPDATE organizations 
		SET name = $2, description = $3, code = $4, type = $5, status = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description, organization.Code,
		organization.Type, organization.Status, organization.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE organizations 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) GetRoots(ctx context.Context) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root organizations: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetDescendants(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
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
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetAncestors(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
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
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetSiblings(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		WITH target_node AS (
			SELECT parent_id FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT t1.id, t1.name, t1.description, t1.code, t1.type, t1.status, t1.parent_id, 
		       t1.record_left, t1.record_right, t1.record_depth, t1.record_ordering,
		       t1.created_at, t1.updated_at, t1.deleted_at
		FROM organizations t1, target_node
		WHERE t1.parent_id = target_node.parent_id AND t1.id != $1 AND t1.deleted_at IS NULL
		ORDER BY t1.record_left ASC`
	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
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
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get path: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetTree(ctx context.Context) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE deleted_at IS NULL
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error) {
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
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtree: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) AddChild(ctx context.Context, parentID, childID uuid.UUID) error {
	if err := r.nestedSet.MoveSubtree(ctx, "organizations", childID, parentID); err != nil {
		return fmt.Errorf("failed to add child using nested set: %w", err)
	}
	_, err := r.db.Exec(ctx, `UPDATE organizations SET updated_at = NOW() WHERE id = $1`, childID)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) MoveSubtree(ctx context.Context, organizationID, newParentID uuid.UUID) error {
	if err := r.nestedSet.MoveSubtree(ctx, "organizations", organizationID, newParentID); err != nil {
		return fmt.Errorf("failed to move subtree using nested set: %w", err)
	}
	_, err := r.db.Exec(ctx, `UPDATE organizations SET updated_at = NOW() WHERE id = $1`, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) DeleteSubtree(ctx context.Context, organizationID uuid.UUID) error {
	if err := r.nestedSet.DeleteSubtree(ctx, "organizations", organizationID); err != nil {
		return fmt.Errorf("failed to delete subtree using nested set: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	searchQuery := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations 
		WHERE (name ILIKE $1 OR description ILIKE $1 OR code ILIKE $1) AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3`
	searchTerm := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, type, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_at, updated_at, deleted_at
		FROM organizations WHERE status = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations by status: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE code = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresOrganizationRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresOrganizationRepository) IsDescendant(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM organizations o1, organizations o2
			WHERE o1.id = $1 AND o2.id = $2 
			AND o1.record_left < o2.record_left 
			AND o1.record_right > o2.record_right
			AND o1.deleted_at IS NULL AND o2.deleted_at IS NULL
		)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, ancestorID, descendantID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check descendant relationship: %w", err)
	}
	return exists, nil
}

func (r *PostgresOrganizationRepository) IsAncestor(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	return r.IsDescendant(ctx, ancestorID, descendantID)
}

func (r *PostgresOrganizationRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count organizations: %w", err)
	}
	return count, nil
}

func (r *PostgresOrganizationRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
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
	if err := r.db.QueryRow(ctx, sqlQuery, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count organizations with search: %w", err)
	}
	return count, nil
}

func (r *PostgresOrganizationRepository) CountByStatus(ctx context.Context, status entities.OrganizationStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE status = $1 AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query, status).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count organizations by status: %w", err)
	}
	return count, nil
}

func (r *PostgresOrganizationRepository) CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE parent_id = $1 AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query, parentID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count children: %w", err)
	}
	return count, nil
}

func (r *PostgresOrganizationRepository) CountDescendants(ctx context.Context, parentID uuid.UUID) (int64, error) {
	query := `
		WITH target_node AS (
			SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL
		)
		SELECT COUNT(*) FROM organizations, target_node
		WHERE record_left > target_node.record_left 
		AND record_right < target_node.record_right 
		AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query, parentID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count descendants: %w", err)
	}
	return count, nil
}

func (r *PostgresOrganizationRepository) ValidateTree(ctx context.Context) ([]string, error) {
	return r.nestedSet.ValidateTree(ctx, "organizations")
}

func (r *PostgresOrganizationRepository) RebuildTree(ctx context.Context) error {
	return r.nestedSet.RebuildTree(ctx, "organizations")
}

func (r *PostgresOrganizationRepository) GetTreeStatistics(ctx context.Context) (map[string]interface{}, error) {
	return r.nestedSet.GetTreeStatistics(ctx, "organizations")
}

func (r *PostgresOrganizationRepository) GetTreeHeight(ctx context.Context) (int64, error) {
	return r.nestedSet.GetTreeHeight(ctx, "organizations")
}

func (r *PostgresOrganizationRepository) GetLevelWidth(ctx context.Context, level uint64) (int64, error) {
	return r.nestedSet.GetLevelWidth(ctx, "organizations", int64(level))
}

func (r *PostgresOrganizationRepository) GetSubtreeSize(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	return r.nestedSet.GetSubtreeSize(ctx, "organizations", organizationID)
}

func (r *PostgresOrganizationRepository) InsertBetween(ctx context.Context, organization *entities.Organization, leftSiblingID, rightSiblingID *uuid.UUID) error {
	// Fallback: simple create
	return r.Create(ctx, organization)
}

func (r *PostgresOrganizationRepository) SwapPositions(ctx context.Context, org1ID, org2ID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var org1Left, org1Right, org2Left, org2Right int64
	err = tx.QueryRow(ctx, `SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL`, org1ID.String()).Scan(&org1Left, &org1Right)
	if err != nil {
		return fmt.Errorf("failed to get first organization info: %w", err)
	}
	err = tx.QueryRow(ctx, `SELECT record_left, record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL`, org2ID.String()).Scan(&org2Left, &org2Right)
	if err != nil {
		return fmt.Errorf("failed to get second organization info: %w", err)
	}
	if (org1Left < org2Left && org1Right > org2Right) || (org2Left < org1Left && org2Right > org1Right) {
		return fmt.Errorf("cannot swap positions: organizations are in ancestor-descendant relationship")
	}
	tempLeft := int64(999999999)
	_, err = tx.Exec(ctx, `UPDATE organizations SET record_left = $1, record_right = $2 WHERE id = $3`, tempLeft, tempLeft+org1Right-org1Left, org1ID.String())
	if err != nil {
		return fmt.Errorf("failed to move first organization to temp position: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE organizations SET record_left = $1, record_right = $2 WHERE id = $3`, org1Left, org1Right, org2ID.String())
	if err != nil {
		return fmt.Errorf("failed to move second organization: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE organizations SET record_left = $1, record_right = $2 WHERE id = $3`, org2Left, org2Right, org1ID.String())
	if err != nil {
		return fmt.Errorf("failed to move first organization to final position: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE organizations SET updated_at = NOW() WHERE id IN ($1, $2)`, org1ID.String(), org2ID.String())
	if err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) GetLeafNodes(ctx context.Context) ([]*entities.Organization, error) {
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
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf nodes: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) GetInternalNodes(ctx context.Context) ([]*entities.Organization, error) {
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
		ORDER BY record_left ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal nodes: %w", err)
	}
	defer rows.Close()
	var organizations []*entities.Organization
	for rows.Next() {
		organization, err := r.scanOrganizationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}
		organizations = append(organizations, organization)
	}
	return organizations, nil
}

func (r *PostgresOrganizationRepository) BatchMoveSubtrees(ctx context.Context, moves []struct {
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
		if err := r.nestedSet.MoveSubtree(ctx, "organizations", move.OrganizationID, move.NewParentID); err != nil {
			return fmt.Errorf("failed to move subtree %s: %w", move.OrganizationID.String(), err)
		}
		if _, err = tx.Exec(ctx, `UPDATE organizations SET updated_at = NOW() WHERE id = $1`, move.OrganizationID.String()); err != nil {
			return fmt.Errorf("failed to update timestamp for %s: %w", move.OrganizationID.String(), err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit batch move transaction: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) BatchInsertBetween(ctx context.Context, insertions []struct {
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
		if err := r.insertBetweenInTx(ctx, tx, insertion.Organization, insertion.LeftSiblingID, insertion.RightSiblingID); err != nil {
			return fmt.Errorf("failed to insert organization %s: %w", insertion.Organization.ID.String(), err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit batch insert transaction: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) insertBetweenInTx(ctx context.Context, tx pgx.Tx, organization *entities.Organization, leftSiblingID, rightSiblingID *uuid.UUID) error {
	var leftSiblingRight, rightSiblingLeft int64
	if leftSiblingID != nil {
		if err := tx.QueryRow(ctx, `SELECT record_right FROM organizations WHERE id = $1 AND deleted_at IS NULL`, leftSiblingID.String()).Scan(&leftSiblingRight); err != nil {
			return fmt.Errorf("failed to get left sibling info: %w", err)
		}
	} else {
		leftSiblingRight = 0
	}
	if rightSiblingID != nil {
		if err := tx.QueryRow(ctx, `SELECT record_left FROM organizations WHERE id = $1 AND deleted_at IS NULL`, rightSiblingID.String()).Scan(&rightSiblingLeft); err != nil {
			return fmt.Errorf("failed to get right sibling info: %w", err)
		}
	} else {
		if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(record_right), 0) FROM organizations WHERE deleted_at IS NULL`).Scan(&rightSiblingLeft); err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}
		rightSiblingLeft++
	}
	newLeft := leftSiblingRight + 1
	newRight := rightSiblingLeft - 1
	if newLeft >= newRight {
		shiftAmount := int64(2)
		if _, err := tx.Exec(ctx, `
			UPDATE organizations 
			SET record_left = CASE 
				WHEN record_left > $1 THEN record_left + $2 
				ELSE record_left 
			END,
			record_right = CASE 
				WHEN record_right >= $1 THEN record_right + $2 
				ELSE record_right 
			END
			WHERE deleted_at IS NULL`, leftSiblingRight, shiftAmount); err != nil {
			return fmt.Errorf("failed to shift nodes: %w", err)
		}
		newLeft = leftSiblingRight + 1
		newRight = newLeft + 1
	}
	organization.RecordLeft = &newLeft
	organization.RecordRight = &newRight
	organization.RecordDepth = &[]int64{0}[0]
	organization.RecordOrdering = &[]int64{newLeft}[0]
	query := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`
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

func (r *PostgresOrganizationRepository) OptimizeTree(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `
		SELECT id, record_left, record_right, record_depth 
		FROM organizations 
		WHERE deleted_at IS NULL 
		ORDER BY record_left ASC`)
	if err != nil {
		return fmt.Errorf("failed to get organizations for optimization: %w", err)
	}
	defer rows.Close()
	var organizations []struct {
		ID    uuid.UUID
		Left  int64
		Right int64
		Depth int64
	}
	for rows.Next() {
		var org struct {
			ID    uuid.UUID
			Left  int64
			Right int64
			Depth int64
		}
		if err := rows.Scan(&org.ID, &org.Left, &org.Right, &org.Depth); err != nil {
			return fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, org)
	}
	newLeft := int64(1)
	for _, org := range organizations {
		subtreeSize := org.Right - org.Left + 1
		newRight := newLeft + subtreeSize - 1
		if _, err := tx.Exec(ctx, `
			UPDATE organizations 
			SET record_left = $1, record_right = $2, record_depth = $3
			WHERE id = $4`, newLeft, newRight, org.Depth, org.ID.String()); err != nil {
			return fmt.Errorf("failed to update organization %s: %w", org.ID.String(), err)
		}
		newLeft = newRight + 1
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit optimization transaction: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	count, err := r.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree size: %w", err)
	}
	metrics["total_nodes"] = count
	height, err := r.GetTreeHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree height: %w", err)
	}
	metrics["tree_height"] = height
	var avgChildren float64
	_ = r.db.QueryRow(ctx, `
		SELECT AVG(child_count) FROM (
			SELECT COUNT(*) as child_count 
			FROM organizations o1
			LEFT JOIN organizations o2 ON o2.parent_id = o1.id AND o2.deleted_at IS NULL
			WHERE o1.deleted_at IS NULL
			GROUP BY o1.id
		) as children_stats`).Scan(&avgChildren)
	metrics["average_children_per_node"] = avgChildren
	var leafCount int64
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM organizations o1
		WHERE o1.deleted_at IS NULL 
		AND NOT EXISTS (
			SELECT 1 FROM organizations o2 
			WHERE o2.parent_id = o1.id AND o2.deleted_at IS NULL
		)`).Scan(&leafCount)
	leafPercentage := float64(0)
	if count > 0 {
		leafPercentage = float64(leafCount) / float64(count) * 100
	}
	metrics["leaf_node_percentage"] = leafPercentage
	var depthStdDev float64
	_ = r.db.QueryRow(ctx, `SELECT STDDEV(record_depth) FROM organizations WHERE deleted_at IS NULL`).Scan(&depthStdDev)
	metrics["depth_standard_deviation"] = depthStdDev
	return metrics, nil
}

func (r *PostgresOrganizationRepository) ValidateTreeIntegrity(ctx context.Context) ([]string, error) {
	var errors []string
	var orphanedCount int64
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM organizations o1
		WHERE o1.deleted_at IS NULL 
		AND o1.parent_id IS NOT NULL 
		AND NOT EXISTS (
			SELECT 1 FROM organizations o2 
			WHERE o2.id::text = o1.parent_id AND o2.deleted_at IS NULL
		)`).Scan(&orphanedCount); err != nil {
		return nil, fmt.Errorf("failed to check for orphaned nodes: %w", err)
	}
	if orphanedCount > 0 {
		errors = append(errors, fmt.Sprintf("Found %d orphaned nodes", orphanedCount))
	}
	var circularCount int64
	err := r.db.QueryRow(ctx, `
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
		SELECT COUNT(*) FROM cycle_check`).Scan(&circularCount)
	if err != nil {
		circularCount = 1
	}
	if circularCount > 0 {
		errors = append(errors, "Found circular references in tree structure")
	}
	var inconsistentCount int64
	if err := r.db.QueryRow(ctx, `
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
		)`).Scan(&inconsistentCount); err != nil {
		return nil, fmt.Errorf("failed to check nested set consistency: %w", err)
	}
	if inconsistentCount > 0 {
		errors = append(errors, fmt.Sprintf("Found %d nodes with inconsistent nested set values", inconsistentCount))
	}
	return errors, nil
}

func (r *PostgresOrganizationRepository) scanOrganizationRow(rows pgx.Rows) (*entities.Organization, error) {
	var organization entities.Organization
	var parentIDStr *string
	if err := rows.Scan(
		&organization.ID, &organization.Name, &organization.Description, &organization.Code,
		&organization.Type, &organization.Status, &parentIDStr,
		&organization.RecordLeft, &organization.RecordRight, &organization.RecordDepth, &organization.RecordOrdering,
		&organization.CreatedAt, &organization.UpdatedAt, &organization.DeletedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to scan organization row: %w", err)
	}
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			organization.ParentID = &parentID
		}
	}
	return &organization, nil
}
