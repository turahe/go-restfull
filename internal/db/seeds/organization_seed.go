package seeds

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/infrastructure/adapters"
)

// SeedOrganizations seeds the database with default organizations
// This function creates organizations with appropriate menu and role assignments
// using the organization repository with nested set operations
func SeedOrganizations() error {
	pool := pgx.GetPgxPool()
	ctx := context.Background()

	// Create organization repository instance
	orgRepo := adapters.NewPostgresOrganizationRepository(pool)

	// Default organizations to create
	defaultOrganizations := []struct {
		name        string
		description string
		code        string
		orgType     string
		status      string
		parentCode  *string
	}{
		{
			name:        "Main Company",
			description: "Primary company headquarters",
			code:        "MAIN",
			orgType:     "COMPANY",
			status:      "active",
			parentCode:  nil,
		},
		{
			name:        "IT Department",
			description: "Information Technology Department",
			code:        "IT",
			orgType:     "DEPARTMENT",
			status:      "active",
			parentCode:  stringPtr("MAIN"),
		},
		{
			name:        "HR Department",
			description: "Human Resources Department",
			code:        "HR",
			orgType:     "DEPARTMENT",
			status:      "active",
			parentCode:  stringPtr("MAIN"),
		},
		{
			name:        "Finance Department",
			description: "Finance and Accounting Department",
			code:        "FIN",
			orgType:     "DEPARTMENT",
			status:      "active",
			parentCode:  stringPtr("MAIN"),
		},
		{
			name:        "Marketing Department",
			description: "Marketing and Sales Department",
			code:        "MKT",
			orgType:     "DEPARTMENT",
			status:      "active",
			parentCode:  stringPtr("MAIN"),
		},
		{
			name:        "Branch Office - North",
			description: "Northern Regional Branch Office",
			code:        "BRANCH_NORTH",
			orgType:     "BRANCH_OFFICE",
			status:      "active",
			parentCode:  stringPtr("MAIN"),
		},
		{
			name:        "Branch Office - South",
			description: "Southern Regional Branch Office",
			code:        "BRANCH_SOUTH",
			orgType:     "BRANCH_OFFICE",
			status:      "active",
			parentCode:  stringPtr("MAIN"),
		},
		{
			name:        "Partner Company A",
			description: "Strategic Business Partner",
			code:        "PARTNER_A",
			orgType:     "PARTNER",
			status:      "active",
			parentCode:  nil,
		},
	}

	// Track created organizations for parent-child relationships
	createdOrgs := make(map[string]*entities.Organization)

	for _, orgData := range defaultOrganizations {
		// Check if organization already exists
		exists, err := organizationExists(ctx, pool, orgData.code)
		if err != nil {
			log.Printf("Error checking if organization %s exists: %v", orgData.code, err)
			continue
		}

		if exists {
			log.Printf("Organization %s already exists, skipping", orgData.code)
			continue
		}

		// Create organization using repository with nested set operations
		org, err := createOrganizationWithRepository(ctx, orgRepo, orgData, createdOrgs)
		if err != nil {
			log.Printf("Error creating organization %s: %v", orgData.code, err)
			continue
		}

		// Store created organization for parent-child relationships
		createdOrgs[orgData.code] = org

		log.Printf("Successfully created organization: %s (%s)", orgData.name, orgData.code)

		// Assign roles and menus based on organization type
		if err := assignOrganizationRolesAndMenus(ctx, pool, org.ID, orgData.orgType); err != nil {
			log.Printf("Error assigning roles and menus to organization %s: %v", orgData.code, err)
			continue
		}
	}

	return nil
}

// createOrganizationWithRepository creates an organization using the repository with nested set operations
func createOrganizationWithRepository(ctx context.Context, orgRepo repositories.OrganizationRepository, orgData struct {
	name        string
	description string
	code        string
	orgType     string
	status      string
	parentCode  *string
}, createdOrgs map[string]*entities.Organization) (*entities.Organization, error) {

	// Create organization entity
	org := &entities.Organization{
		ID:          uuid.New(),
		Name:        orgData.name,
		Description: &orgData.description,
		Code:        &orgData.code,
		Type:        func() *entities.OrganizationType { t := entities.OrganizationType(orgData.orgType); return &t }(),
		Status:      entities.OrganizationStatus(orgData.status),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set parent ID if parent code is provided and parent exists
	if orgData.parentCode != nil {
		if parentOrg, exists := createdOrgs[*orgData.parentCode]; exists {
			org.ParentID = &parentOrg.ID
		}
	}

	// Try to create organization using repository (this will handle nested set operations)
	if err := orgRepo.Create(ctx, org); err != nil {
		// If nested set creation fails, try fallback approach for the first organization
		if orgData.parentCode == nil && len(createdOrgs) == 0 {
			log.Printf("Nested set creation failed for first organization %s, trying fallback approach: %v", orgData.code, err)
			return createOrganizationFallback(ctx, orgData, createdOrgs)
		}
		return nil, err
	}

	return org, nil
}

// createOrganizationFallback creates the first organization with manual nested set values
// This is used when the nested set manager fails to create the first organization
func createOrganizationFallback(ctx context.Context, orgData struct {
	name        string
	description string
	code        string
	orgType     string
	status      string
	parentCode  *string
}, createdOrgs map[string]*entities.Organization) (*entities.Organization, error) {
	pool := pgx.GetPgxPool()

	orgID := uuid.New()
	var parentID *uuid.UUID

	// Get parent organization ID if parent code is provided
	if orgData.parentCode != nil {
		var existingParentID uuid.UUID
		parentQuery := `SELECT id FROM organizations WHERE code = $1 AND deleted_at IS NULL`
		err := pool.QueryRow(ctx, parentQuery, *orgData.parentCode).Scan(&existingParentID)
		if err == nil {
			parentID = &existingParentID
		}
	}

	// For the first organization, set manual nested set values
	var recordLeft, recordRight, recordDepth, recordOrdering int64
	if parentID == nil {
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

	insertQuery := `
		INSERT INTO organizations (
			id, name, description, code, type, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err := pool.Exec(ctx, insertQuery,
		orgID,
		orgData.name,
		orgData.description,
		orgData.code,
		orgData.orgType,
		orgData.status,
		parentID,
		recordLeft,
		recordRight,
		recordDepth,
		recordOrdering,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return nil, err
	}

	// Create and return the organization entity
	org := &entities.Organization{
		ID:             orgID,
		Name:           orgData.name,
		Description:    &orgData.description,
		Code:           &orgData.code,
		Type:           func() *entities.OrganizationType { t := entities.OrganizationType(orgData.orgType); return &t }(),
		Status:         entities.OrganizationStatus(orgData.status),
		ParentID:       parentID,
		RecordLeft:     &recordLeft,
		RecordRight:    &recordRight,
		RecordDepth:    &recordDepth,
		RecordOrdering: &recordOrdering,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return org, nil
}

// organizationExists checks if an organization with the given code exists
func organizationExists(ctx context.Context, pool *pgxpool.Pool, code string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE code = $1 AND deleted_at IS NULL)`
	err := pool.QueryRow(ctx, query, code).Scan(&exists)
	return exists, err
}

// assignOrganizationRolesAndMenus assigns appropriate roles and menus based on organization type
func assignOrganizationRolesAndMenus(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, orgType string) error {
	// Define role and menu assignments based on organization type
	var roleSlugs []string
	var menuSlugs []string

	switch orgType {
	case "COMPANY":
		// Main company gets all roles and menus
		roleSlugs = []string{"admin", "user", "moderator", "editor", "viewer"}
		menuSlugs = []string{"dashboard", "users", "roles", "menus", "posts", "media", "settings"}

	case "DEPARTMENT":
		// Departments get limited roles and menus
		roleSlugs = []string{"user", "moderator", "editor", "viewer"}
		menuSlugs = []string{"dashboard", "posts", "media"}

	case "BRANCH_OFFICE":
		// Branch offices get moderate access
		roleSlugs = []string{"user", "moderator", "editor", "viewer"}
		menuSlugs = []string{"dashboard", "users", "posts", "media", "settings"}

	case "PARTNER":
		// Partners get limited access
		roleSlugs = []string{"user", "viewer"}
		menuSlugs = []string{"dashboard", "posts", "media"}

	default:
		// Default access for other types
		roleSlugs = []string{"user", "viewer"}
		menuSlugs = []string{"dashboard"}
	}

	// Assign roles to organization
	if err := assignRolesToOrganization(ctx, pool, orgID, roleSlugs); err != nil {
		return err
	}

	// Assign menus to organization
	if err := assignMenusToOrganization(ctx, pool, orgID, menuSlugs); err != nil {
		return err
	}

	log.Printf("Assigned %d roles and %d menus to organization type: %s", len(roleSlugs), len(menuSlugs), orgType)
	return nil
}

// assignRolesToOrganization assigns roles to an organization
func assignRolesToOrganization(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, roleSlugs []string) error {
	for _, roleSlug := range roleSlugs {
		// Get role ID
		var roleID uuid.UUID
		roleQuery := `SELECT id FROM roles WHERE slug = $1 AND deleted_at IS NULL`
		err := pool.QueryRow(ctx, roleQuery, roleSlug).Scan(&roleID)
		if err != nil {
			log.Printf("Error getting role ID for %s: %v", roleSlug, err)
			continue
		}

		// Check if organization-role relationship already exists
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM organization_roles WHERE organization_id = $1 AND role_id = $2)`
		err = pool.QueryRow(ctx, checkQuery, orgID, roleID).Scan(&exists)
		if err != nil {
			// If organization_roles table doesn't exist, skip role assignment
			log.Printf("Organization roles table not found, skipping role assignment")
			return nil
		}

		if exists {
			continue
		}

		// Create organization-role relationship
		insertQuery := `
			INSERT INTO organization_roles (id, organization_id, role_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`

		_, err = pool.Exec(ctx, insertQuery,
			uuid.New(),
			orgID,
			roleID,
			time.Now(),
			time.Now(),
		)

		if err != nil {
			log.Printf("Error assigning role %s to organization: %v", roleSlug, err)
			continue
		}
	}

	return nil
}

// assignMenusToOrganization assigns menus to an organization
func assignMenusToOrganization(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, menuSlugs []string) error {
	for _, menuSlug := range menuSlugs {
		// Get menu ID
		var menuID uuid.UUID
		menuQuery := `SELECT id FROM menus WHERE slug = $1 AND deleted_at IS NULL`
		err := pool.QueryRow(ctx, menuQuery, menuSlug).Scan(&menuID)
		if err != nil {
			log.Printf("Error getting menu ID for %s: %v", menuSlug, err)
			continue
		}

		// Check if organization-menu relationship already exists
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM organization_menus WHERE organization_id = $1 AND menu_id = $2)`
		err = pool.QueryRow(ctx, checkQuery, orgID, menuID).Scan(&exists)
		if err != nil {
			// If organization_menus table doesn't exist, skip menu assignment
			log.Printf("Organization menus table not found, skipping menu assignment")
			return nil
		}

		if exists {
			continue
		}

		// Create organization-menu relationship
		insertQuery := `
			INSERT INTO organization_menus (id, organization_id, menu_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`

		_, err = pool.Exec(ctx, insertQuery,
			uuid.New(),
			orgID,
			menuID,
			time.Now(),
			time.Now(),
		)

		if err != nil {
			log.Printf("Error assigning menu %s to organization: %v", menuSlug, err)
			continue
		}
	}

	return nil
}

// Helper function to convert string to pointer
func stringPtr(s string) *string {
	return &s
}
