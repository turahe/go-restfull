package adapters

import (
	"context"
	"fmt"

	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// organizationRepository implements the OrganizationRepository interface
type organizationRepository struct {
	db *pgxpool.Pool
}

// NewOrganizationRepository creates a new organization repository instance
func NewOrganizationRepository(db *pgxpool.Pool) repositories.OrganizationRepository {
	return &organizationRepository{
		db: db,
	}
}

func (r *organizationRepository) Create(ctx context.Context, organization *entities.Organization) error {
	query := `
		INSERT INTO organizations (
			id, name, description, code, email, phone, address, website, logo_url, 
			status, parent_id, record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	// Calculate nested set values
	var left, right, depth, ordering int
	if organization.ParentID != nil {
		// Get parent's nested set values
		var parentLeft, parentRight, parentDepth int
		err := r.db.QueryRow(ctx, `
			SELECT record_left, record_right, record_depth 
			FROM organizations WHERE id = $1
		`, organization.ParentID).Scan(&parentLeft, &parentRight, &parentDepth)
		if err != nil {
			return fmt.Errorf("failed to get parent nested set values: %w", err)
		}

		// Make space for the new node
		_, err = r.db.Exec(ctx, `
			UPDATE organizations 
			SET record_right = record_right + 2 
			WHERE record_right >= $1
		`, parentRight)
		if err != nil {
			return fmt.Errorf("failed to update nested set values: %w", err)
		}

		_, err = r.db.Exec(ctx, `
			UPDATE organizations 
			SET record_left = record_left + 2 
			WHERE record_left > $1
		`, parentRight)
		if err != nil {
			return fmt.Errorf("failed to update nested set values: %w", err)
		}

		left = parentRight
		right = parentRight + 1
		depth = parentDepth + 1
		ordering = 1 // Default ordering
	} else {
		// Root node
		var maxRight int
		err := r.db.QueryRow(ctx, `
			SELECT COALESCE(MAX(record_right), 0) FROM organizations
		`).Scan(&maxRight)
		if err != nil {
			return fmt.Errorf("failed to get max right value: %w", err)
		}

		left = maxRight + 1
		right = maxRight + 2
		depth = 0
		ordering = 1
	}

	organization.SetNestedSetValues(&left, &right, &depth, &ordering)

	_, err := r.db.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description,
		organization.Code, organization.Email, organization.Phone,
		organization.Address, organization.Website, organization.LogoURL,
		organization.Status, organization.ParentID,
		organization.RecordLeft, organization.RecordRight,
		organization.RecordDepth, organization.RecordOrdering,
		organization.CreatedAt, organization.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE id = $1
	`

	var org entities.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
		&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
		&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
		&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

func (r *organizationRepository) GetByCode(ctx context.Context, code string) (*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE code = $1
	`

	var org entities.Organization
	err := r.db.QueryRow(ctx, query, code).Scan(
		&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
		&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
		&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
		&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

func (r *organizationRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE deleted_at IS NULL
		ORDER BY record_left
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) Update(ctx context.Context, organization *entities.Organization) error {
	query := `
		UPDATE organizations
		SET name = $2, description = $3, code = $4, email = $5, phone = $6,
			address = $7, website = $8, logo_url = $9, status = $10,
			parent_id = $11, record_left = $12, record_right = $13,
			record_depth = $14, record_ordering = $15, updated_at = $16
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		organization.ID, organization.Name, organization.Description,
		organization.Code, organization.Email, organization.Phone,
		organization.Address, organization.Website, organization.LogoURL,
		organization.Status, organization.ParentID,
		organization.RecordLeft, organization.RecordRight,
		organization.RecordDepth, organization.RecordOrdering,
		organization.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE organizations
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

func (r *organizationRepository) GetRoots(ctx context.Context) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY record_left
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_ordering, record_left
	`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetDescendants(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE record_left > (SELECT record_left FROM organizations WHERE id = $1)
		  AND record_right < (SELECT record_right FROM organizations WHERE id = $1)
		  AND deleted_at IS NULL
		ORDER BY record_left
	`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendant organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetAncestors(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE record_left < (SELECT record_left FROM organizations WHERE id = $1)
		  AND record_right > (SELECT record_right FROM organizations WHERE id = $1)
		  AND deleted_at IS NULL
		ORDER BY record_left
	`

	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestor organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetSiblings(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE parent_id = (SELECT parent_id FROM organizations WHERE id = $1)
		  AND id != $1
		  AND deleted_at IS NULL
		ORDER BY record_ordering, record_left
	`

	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sibling organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		WITH RECURSIVE org_path AS (
			SELECT id, name, description, code, email, phone, address, website, logo_url,
				   status, parent_id, record_left, record_right, record_depth, record_ordering,
				   created_at, updated_at, deleted_at, 1 as level
			FROM organizations
			WHERE id = $1
			
			UNION ALL
			
			SELECT o.id, o.name, o.description, o.code, o.email, o.phone, o.address, o.website, o.logo_url,
				   o.status, o.parent_id, o.record_left, o.record_right, o.record_depth, o.record_ordering,
				   o.created_at, o.updated_at, o.deleted_at, op.level + 1
			FROM organizations o
			INNER JOIN org_path op ON o.id = op.parent_id
		)
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM org_path
		ORDER BY level DESC
	`

	rows, err := r.db.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization path: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetTree(ctx context.Context) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE deleted_at IS NULL
		ORDER BY record_left
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization tree: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE record_left >= (SELECT record_left FROM organizations WHERE id = $1)
		  AND record_right <= (SELECT record_right FROM organizations WHERE id = $1)
		  AND deleted_at IS NULL
		ORDER BY record_left
	`

	rows, err := r.db.Query(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization subtree: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) AddChild(ctx context.Context, parentID, childID uuid.UUID) error {
	// This operation is handled during creation, so we just need to update the parent_id
	query := `
		UPDATE organizations
		SET parent_id = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, childID, parentID)
	if err != nil {
		return fmt.Errorf("failed to add child organization: %w", err)
	}

	return nil
}

func (r *organizationRepository) MoveSubtree(ctx context.Context, organizationID, newParentID uuid.UUID) error {
	// This is a complex operation that requires recalculating nested set values
	// For now, we'll implement a simplified version that just updates the parent_id
	// A full implementation would require recalculating all nested set values

	query := `
		UPDATE organizations
		SET parent_id = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, organizationID, newParentID)
	if err != nil {
		return fmt.Errorf("failed to move organization: %w", err)
	}

	return nil
}

func (r *organizationRepository) DeleteSubtree(ctx context.Context, organizationID uuid.UUID) error {
	query := `
		UPDATE organizations
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE record_left >= (SELECT record_left FROM organizations WHERE id = $1)
		  AND record_right <= (SELECT record_right FROM organizations WHERE id = $1)
	`

	_, err := r.db.Exec(ctx, query, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete organization subtree: %w", err)
	}

	return nil
}

func (r *organizationRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	searchQuery := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR description ILIKE $1 OR code ILIKE $1 OR email ILIKE $1)
		ORDER BY record_left
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) GetByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error) {
	query := `
		SELECT id, name, description, code, email, phone, address, website, logo_url,
			   status, parent_id, record_left, record_right, record_depth, record_ordering,
			   created_at, updated_at, deleted_at
		FROM organizations
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY record_left
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations by status: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		err := rows.Scan(
			&org.ID, &org.Name, &org.Description, &org.Code, &org.Email,
			&org.Phone, &org.Address, &org.Website, &org.LogoURL, &org.Status,
			&org.ParentID, &org.RecordLeft, &org.RecordRight, &org.RecordDepth,
			&org.RecordOrdering, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE code = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check organization code existence: %w", err)
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
			  AND o2.record_left > o1.record_left
			  AND o2.record_right < o1.record_right
		)
	`

	var isDescendant bool
	err := r.db.QueryRow(ctx, query, ancestorID, descendantID).Scan(&isDescendant)
	if err != nil {
		return false, fmt.Errorf("failed to check descendant relationship: %w", err)
	}

	return isDescendant, nil
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
	searchQuery := `
		SELECT COUNT(*) FROM organizations
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR description ILIKE $1 OR code ILIKE $1 OR email ILIKE $1)
	`

	searchTerm := "%" + query + "%"
	var count int64
	err := r.db.QueryRow(ctx, searchQuery, searchTerm).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations by search: %w", err)
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
	query := `
		SELECT COUNT(*) FROM organizations
		WHERE record_left > (SELECT record_left FROM organizations WHERE id = $1)
		  AND record_right < (SELECT record_right FROM organizations WHERE id = $1)
		  AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, parentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count descendants: %w", err)
	}

	return count, nil
}
