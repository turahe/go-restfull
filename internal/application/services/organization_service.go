// Package services provides application-level business logic for organization management.
// This package contains the organization service implementation that handles hierarchical
// organization structures, organization lifecycle, and complex organizational relationships
// while ensuring proper data integrity and business rules.
package services

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// organizationService implements the OrganizationService interface and provides comprehensive
// organization management functionality. It handles hierarchical organization structures,
// organization lifecycle, complex organizational relationships, and business rules
// while ensuring proper data integrity and validation.
type organizationService struct {
	organizationRepo repositories.OrganizationRepository
}

// NewOrganizationService creates a new organization service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - organizationRepo: Repository interface for organization data access operations
//
// Returns:
//   - ports.OrganizationService: The organization service interface implementation
func NewOrganizationService(
	organizationRepo repositories.OrganizationRepository,
) ports.OrganizationService {
	return &organizationService{
		organizationRepo: organizationRepo,
	}
}

// CreateOrganization creates a new organization with comprehensive validation and hierarchy support.
// This method enforces business rules for organization creation and supports hierarchical
// structures with proper code uniqueness and parent validation.
//
// Business Rules:
//   - Organization name is required and validated
//   - Code must be unique if provided
//   - Parent organization must exist if specified
//   - Organization validation ensures proper structure
//   - Hierarchical relationships are validated
//
// Parameters:
//   - ctx: Context for the operation
//   - name: Display name of the organization
//   - description: Optional description of the organization
//   - code: Optional unique code for the organization
//   - email: Contact email for the organization
//   - phone: Contact phone for the organization
//   - address: Physical address of the organization
//   - website: Website URL of the organization
//   - logoURL: Logo image URL for the organization
//   - parentID: Optional parent organization ID for hierarchical structure
//
// Returns:
//   - *entities.Organization: The created organization entity
//   - error: Any error that occurred during the operation
func (s *organizationService) CreateOrganization(ctx context.Context, name, description, code, email, phone, address, website, logoURL string, parentID *uuid.UUID) (*entities.Organization, error) {
	// Validate code uniqueness if provided
	if code != "" {
		exists, err := s.organizationRepo.ExistsByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("organization with this code already exists")
		}
	}

	// Validate parent exists if provided
	if parentID != nil {
		exists, err := s.organizationRepo.ExistsByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("parent organization not found")
		}
	}

	// Create organization entity with the provided parameters
	organization, err := entities.NewOrganization(name, description, code, email, phone, address, website, logoURL, parentID)
	if err != nil {
		return nil, err
	}

	// Persist the organization to the repository
	if err := s.organizationRepo.Create(ctx, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

// GetOrganizationByID retrieves an organization by its unique identifier.
// This method includes soft delete checking to ensure deleted organizations
// are not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the organization to retrieve
//
// Returns:
//   - *entities.Organization: The organization entity if found
//   - error: Error if organization not found or other issues occur
func (s *organizationService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Check if the organization has been soft deleted
	if organization.IsDeleted() {
		return nil, errors.New("organization not found")
	}
	return organization, nil
}

// GetOrganizationByCode retrieves an organization by its unique code identifier.
// This method is useful for code-based organization lookups and routing.
//
// Parameters:
//   - ctx: Context for the operation
//   - code: Code identifier of the organization to retrieve
//
// Returns:
//   - *entities.Organization: The organization entity if found
//   - error: Error if organization not found or other issues occur
func (s *organizationService) GetOrganizationByCode(ctx context.Context, code string) (*entities.Organization, error) {
	organization, err := s.organizationRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	// Check if the organization has been soft deleted
	if organization.IsDeleted() {
		return nil, errors.New("organization not found")
	}
	return organization, nil
}

// GetAllOrganizations retrieves all organizations in the system with pagination.
// This method is useful for administrative purposes and organization management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of organizations to return
//   - offset: Number of organizations to skip for pagination
//
// Returns:
//   - []*entities.Organization: List of all organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetAllOrganizations(ctx context.Context, limit, offset int) ([]*entities.Organization, error) {
	return s.organizationRepo.GetAll(ctx, limit, offset)
}

// UpdateOrganization updates an existing organization's information and metadata.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Organization must exist and not be deleted
//   - Updated code must be unique if changed
//   - Organization validation ensures proper structure
//   - Hierarchical relationships are maintained
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the organization to update
//   - name: Updated display name of the organization
//   - description: Updated description of the organization
//   - code: Updated unique code for the organization
//   - email: Updated contact email for the organization
//   - phone: Updated contact phone for the organization
//   - address: Updated physical address of the organization
//   - website: Updated website URL of the organization
//   - logoURL: Updated logo image URL for the organization
//
// Returns:
//   - *entities.Organization: The updated organization entity
//   - error: Any error that occurred during the operation
func (s *organizationService) UpdateOrganization(ctx context.Context, id uuid.UUID, name, description, code, email, phone, address, website, logoURL string) (*entities.Organization, error) {
	// Retrieve existing organization to ensure it exists and is not deleted
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if organization.IsDeleted() {
		return nil, errors.New("organization not found")
	}

	// Validate code uniqueness if changed
	if code != "" && (organization.Code == nil || *organization.Code != code) {
		exists, err := s.organizationRepo.ExistsByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("organization with this code already exists")
		}
	}

	// Update the organization entity with new information
	if err := organization.UpdateOrganization(name, description, code, email, phone, address, website, logoURL); err != nil {
		return nil, err
	}

	// Persist the updated organization to the repository
	if err := s.organizationRepo.Update(ctx, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

// DeleteOrganization performs a soft delete of an organization by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Business Rules:
//   - Organization must exist before deletion
//   - Soft delete preserves organization data
//   - Deleted organizations are not returned in queries
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the organization to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if organization.IsDeleted() {
		return errors.New("organization not found")
	}

	return s.organizationRepo.Delete(ctx, id)
}

// GetRootOrganizations retrieves all root-level organizations (no parent).
// This method is useful for building top-level organizational structures.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Organization: List of root-level organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetRootOrganizations(ctx context.Context) ([]*entities.Organization, error) {
	return s.organizationRepo.GetRoots(ctx)
}

// GetChildOrganizations retrieves all direct children of a specific organization.
// This method supports hierarchical navigation and organization tree building.
//
// Business Rules:
//   - Parent organization must exist and be valid
//   - Returns only direct children (not descendants)
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent organization
//
// Returns:
//   - []*entities.Organization: List of child organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetChildOrganizations(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	// Validate parent exists before retrieving children
	exists, err := s.organizationRepo.ExistsByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("parent organization not found")
	}

	return s.organizationRepo.GetChildren(ctx, parentID)
}

// GetDescendantOrganizations retrieves all descendants of a specific organization.
// This method supports deep hierarchical navigation and organization tree traversal.
//
// Business Rules:
//   - Parent organization must exist and be valid
//   - Returns all descendants at any level
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent organization
//
// Returns:
//   - []*entities.Organization: List of descendant organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetDescendantOrganizations(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	// Validate parent exists before retrieving descendants
	exists, err := s.organizationRepo.ExistsByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("parent organization not found")
	}

	return s.organizationRepo.GetDescendants(ctx, parentID)
}

// GetAncestorOrganizations retrieves all ancestors of a specific organization.
// This method supports upward hierarchical navigation and breadcrumb generation.
//
// Business Rules:
//   - Organization must exist and be valid
//   - Returns all ancestors at any level
//
// Parameters:
//   - ctx: Context for the operation
//   - organizationID: UUID of the organization to get ancestors for
//
// Returns:
//   - []*entities.Organization: List of ancestor organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetAncestorOrganizations(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Validate organization exists before retrieving ancestors
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("organization not found")
	}

	return s.organizationRepo.GetAncestors(ctx, organizationID)
}

// GetSiblingOrganizations retrieves all siblings of a specific organization.
// This method supports lateral navigation within the same hierarchical level.
//
// Business Rules:
//   - Organization must exist and be valid
//   - Returns organizations at the same hierarchical level
//
// Parameters:
//   - ctx: Context for the operation
//   - organizationID: UUID of the organization to get siblings for
//
// Returns:
//   - []*entities.Organization: List of sibling organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetSiblingOrganizations(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Validate organization exists before retrieving siblings
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("organization not found")
	}

	return s.organizationRepo.GetSiblings(ctx, organizationID)
}

// GetOrganizationPath retrieves the complete path from root to a specific organization.
// This method is useful for breadcrumb generation and hierarchical navigation.
//
// Business Rules:
//   - Organization must exist and be valid
//   - Returns complete path from root to target organization
//
// Parameters:
//   - ctx: Context for the operation
//   - organizationID: UUID of the organization to get path for
//
// Returns:
//   - []*entities.Organization: Complete path from root to organization
//   - error: Any error that occurred during the operation
func (s *organizationService) GetOrganizationPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Validate organization exists before retrieving path
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("organization not found")
	}

	return s.organizationRepo.GetPath(ctx, organizationID)
}

// GetOrganizationTree retrieves the complete organization hierarchy.
// This method is useful for building complete organizational trees and navigation structures.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Organization: Complete organization hierarchy
//   - error: Any error that occurred during the operation
func (s *organizationService) GetOrganizationTree(ctx context.Context) ([]*entities.Organization, error) {
	return s.organizationRepo.GetTree(ctx)
}

// GetOrganizationSubtree retrieves a subtree starting from a specific organization.
// This method is useful for building partial organizational trees.
//
// Business Rules:
//   - Root organization must exist and be valid
//   - Returns complete subtree from root organization
//
// Parameters:
//   - ctx: Context for the operation
//   - rootID: UUID of the root organization for the subtree
//
// Returns:
//   - []*entities.Organization: Complete subtree from root organization
//   - error: Any error that occurred during the operation
func (s *organizationService) GetOrganizationSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error) {
	// Validate root exists before retrieving subtree
	exists, err := s.organizationRepo.ExistsByID(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("root organization not found")
	}

	return s.organizationRepo.GetSubtree(ctx, rootID)
}

// AddChildOrganization establishes a parent-child relationship between organizations.
// This method enforces hierarchy validation and prevents circular references.
//
// Business Rules:
//   - Both parent and child organizations must exist
//   - Circular references are prevented
//   - Hierarchy validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent organization
//   - childID: UUID of the child organization
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *organizationService) AddChildOrganization(ctx context.Context, parentID, childID uuid.UUID) error {
	// Validate parent organization exists
	parentExists, err := s.organizationRepo.ExistsByID(ctx, parentID)
	if err != nil {
		return err
	}
	if !parentExists {
		return errors.New("parent organization not found")
	}

	// Validate child organization exists
	childExists, err := s.organizationRepo.ExistsByID(ctx, childID)
	if err != nil {
		return err
	}
	if !childExists {
		return errors.New("child organization not found")
	}

	// Validate hierarchy to prevent circular references
	if err := s.ValidateOrganizationHierarchy(ctx, parentID, childID); err != nil {
		return err
	}

	return s.organizationRepo.AddChild(ctx, parentID, childID)
}

// MoveOrganization moves an organization to a new parent in the hierarchy.
// This method enforces hierarchy validation and prevents circular references.
//
// Business Rules:
//   - Both organizations must exist
//   - Circular references are prevented
//   - Hierarchy validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - organizationID: UUID of the organization to move
//   - newParentID: UUID of the new parent organization
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *organizationService) MoveOrganization(ctx context.Context, organizationID, newParentID uuid.UUID) error {
	// Validate organization exists
	orgExists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if !orgExists {
		return errors.New("organization not found")
	}

	// Validate new parent exists
	parentExists, err := s.organizationRepo.ExistsByID(ctx, newParentID)
	if err != nil {
		return err
	}
	if !parentExists {
		return errors.New("new parent organization not found")
	}

	// Validate hierarchy to prevent circular references
	if err := s.ValidateOrganizationHierarchy(ctx, newParentID, organizationID); err != nil {
		return err
	}

	return s.organizationRepo.MoveSubtree(ctx, organizationID, newParentID)
}

// DeleteOrganizationSubtree deletes an organization and all its descendants.
// This method performs a comprehensive deletion of the entire subtree.
//
// Business Rules:
//   - Organization must exist before deletion
//   - All descendants are also deleted
//   - Deletion is comprehensive and irreversible
//
// Parameters:
//   - ctx: Context for the operation
//   - organizationID: UUID of the organization to delete with its subtree
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *organizationService) DeleteOrganizationSubtree(ctx context.Context, organizationID uuid.UUID) error {
	// Validate organization exists before deletion
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("organization not found")
	}

	return s.organizationRepo.DeleteSubtree(ctx, organizationID)
}

// SetOrganizationStatus updates the status of an organization.
// This method is part of the organization lifecycle management.
//
// Business Rules:
//   - Organization must exist and not be deleted
//   - Status must be valid for the organization
//   - Status change is atomic and consistent
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the organization to update status for
//   - status: New status for the organization
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *organizationService) SetOrganizationStatus(ctx context.Context, id uuid.UUID, status entities.OrganizationStatus) error {
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if organization.IsDeleted() {
		return errors.New("organization not found")
	}

	// Update organization status
	if err := organization.SetStatus(status); err != nil {
		return err
	}

	return s.organizationRepo.Update(ctx, organization)
}

// GetOrganizationsByStatus retrieves organizations with a specific status.
// This method is useful for filtering organizations by their lifecycle status.
//
// Parameters:
//   - ctx: Context for the operation
//   - status: Status to filter organizations by
//   - limit: Maximum number of organizations to return
//   - offset: Number of organizations to skip for pagination
//
// Returns:
//   - []*entities.Organization: List of organizations with the specified status
//   - error: Any error that occurred during the operation
func (s *organizationService) GetOrganizationsByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error) {
	return s.organizationRepo.GetByStatus(ctx, status, limit, offset)
}

// SearchOrganizations searches for organizations based on a query string.
// This method supports full-text search capabilities for finding organizations
// by name, description, code, or other attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Organization: List of matching organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) SearchOrganizations(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	return s.organizationRepo.Search(ctx, query, limit, offset)
}

// GetOrganizationsWithPagination retrieves organizations with pagination and filtering.
// This method provides a comprehensive pagination solution with search and status filtering.
//
// Business Rules:
//   - Page and perPage parameters are properly handled
//   - Search and status filtering are mutually exclusive
//   - Total count is calculated for pagination metadata
//   - Offset is calculated based on page and perPage
//
// Parameters:
//   - ctx: Context for the operation
//   - page: Current page number (1-based)
//   - perPage: Number of organizations per page
//   - search: Optional search query for filtering
//   - status: Optional status filter for organizations
//
// Returns:
//   - []*entities.Organization: List of organizations for the current page
//   - int64: Total count of organizations for pagination
//   - error: Any error that occurred during the operation
func (s *organizationService) GetOrganizationsWithPagination(ctx context.Context, page, perPage int, search string, status entities.OrganizationStatus) ([]*entities.Organization, int64, error) {
	limit := perPage
	offset := (page - 1) * perPage

	var organizations []*entities.Organization
	var total int64
	var err error

	// Apply search filter if provided
	if search != "" {
		organizations, err = s.organizationRepo.Search(ctx, search, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.organizationRepo.CountBySearch(ctx, search)
	} else if status != "" {
		// Apply status filter if provided
		organizations, err = s.organizationRepo.GetByStatus(ctx, status, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.organizationRepo.CountByStatus(ctx, status)
	} else {
		// Get all organizations if no filters provided
		organizations, err = s.organizationRepo.GetAll(ctx, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.organizationRepo.Count(ctx)
	}

	if err != nil {
		return nil, 0, err
	}

	return organizations, total, nil
}

// ValidateOrganizationHierarchy validates that a parent-child relationship
// does not create circular references in the organization hierarchy.
//
// Business Rules:
//   - Child cannot be a descendant of parent
//   - Parent cannot be a descendant of child
//   - Circular references are prevented
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent organization
//   - childID: UUID of the child organization
//
// Returns:
//   - error: Any error that occurred during validation
func (s *organizationService) ValidateOrganizationHierarchy(ctx context.Context, parentID, childID uuid.UUID) error {
	// Check if child is already a descendant of parent
	isDescendant, err := s.organizationRepo.IsDescendant(ctx, parentID, childID)
	if err != nil {
		return err
	}
	if isDescendant {
		return errors.New("cannot create circular reference in organization hierarchy")
	}

	// Check if parent is a descendant of child
	isAncestor, err := s.organizationRepo.IsAncestor(ctx, childID, parentID)
	if err != nil {
		return err
	}
	if isAncestor {
		return errors.New("cannot create circular reference in organization hierarchy")
	}

	return nil
}

// CheckOrganizationExists verifies if an organization exists by its ID.
// This method is useful for validation and existence checking.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the organization to check
//
// Returns:
//   - bool: True if organization exists, false otherwise
//   - error: Any error that occurred during the operation
func (s *organizationService) CheckOrganizationExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.organizationRepo.ExistsByID(ctx, id)
}

// CheckOrganizationCodeExists verifies if an organization exists by its code.
// This method is useful for code uniqueness validation.
//
// Parameters:
//   - ctx: Context for the operation
//   - code: Code of the organization to check
//
// Returns:
//   - bool: True if organization with code exists, false otherwise
//   - error: Any error that occurred during the operation
func (s *organizationService) CheckOrganizationCodeExists(ctx context.Context, code string) (bool, error) {
	return s.organizationRepo.ExistsByCode(ctx, code)
}

// GetOrganizationsCount returns total count of organizations for pagination calculations.
// This method supports filtering by search query and status.
//
// Parameters:
//   - ctx: Context for the operation
//   - search: Optional search query for filtered count
//   - status: Optional status filter for count
//
// Returns:
//   - int64: Total count of organizations
//   - error: Any error that occurred during the operation
func (s *organizationService) GetOrganizationsCount(ctx context.Context, search string, status entities.OrganizationStatus) (int64, error) {
	if search != "" {
		return s.organizationRepo.CountBySearch(ctx, search)
	}
	if status != "" {
		return s.organizationRepo.CountByStatus(ctx, status)
	}
	return s.organizationRepo.Count(ctx)
}

// GetChildrenCount returns the number of direct children for a specific organization.
// This method is useful for organizational statistics and reporting.
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent organization
//
// Returns:
//   - int64: Number of direct children
//   - error: Any error that occurred during the operation
func (s *organizationService) GetChildrenCount(ctx context.Context, parentID uuid.UUID) (int64, error) {
	return s.organizationRepo.CountChildren(ctx, parentID)
}

// GetDescendantsCount returns the number of all descendants for a specific organization.
// This method is useful for organizational statistics and reporting.
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent organization
//
// Returns:
//   - int64: Number of all descendants
//   - error: Any error that occurred during the operation
func (s *organizationService) GetDescendantsCount(ctx context.Context, parentID uuid.UUID) (int64, error) {
	return s.organizationRepo.CountDescendants(ctx, parentID)
}
