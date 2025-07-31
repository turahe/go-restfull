// Package services provides application-level business logic for taxonomy management.
// This package contains the taxonomy service implementation that handles hierarchical
// categorization, taxonomy creation, retrieval, and hierarchical relationships while
// ensuring proper data organization and navigation.
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/pagination"

	"github.com/google/uuid"
)

// TaxonomyService implements the TaxonomyService interface and provides comprehensive
// taxonomy management functionality. It handles hierarchical categorization, taxonomy
// creation, retrieval, and complex hierarchical relationships while ensuring proper
// data organization and navigation.
type TaxonomyService struct {
	taxonomyRepository repositories.TaxonomyRepository
}

// NewTaxonomyService creates a new taxonomy service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - taxonomyRepository: Repository interface for taxonomy data access operations
//
// Returns:
//   - ports.TaxonomyService: The taxonomy service interface implementation
func NewTaxonomyService(taxonomyRepository repositories.TaxonomyRepository) ports.TaxonomyService {
	return &TaxonomyService{
		taxonomyRepository: taxonomyRepository,
	}
}

// CreateTaxonomy creates a new taxonomy with comprehensive validation and hierarchy support.
// This method enforces business rules for taxonomy creation and supports hierarchical
// structures with proper slug uniqueness.
//
// Business Rules:
//   - Taxonomy name and slug are required and validated
//   - Slug must be unique across the system
//   - Parent ID is optional for root-level taxonomies
//   - Taxonomy validation ensures proper structure
//   - Code is used for programmatic identification
//
// Parameters:
//   - ctx: Context for the operation
//   - name: Display name of the taxonomy
//   - slug: Unique identifier for the taxonomy
//   - code: Programmatic code for the taxonomy
//   - description: Optional description of the taxonomy
//   - parentID: Optional parent taxonomy ID for hierarchical structure
//
// Returns:
//   - *entities.Taxonomy: The created taxonomy entity
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) CreateTaxonomy(ctx context.Context, name, slug, code, description string, parentID *uuid.UUID) (*entities.Taxonomy, error) {
	// Validate taxonomy name to ensure it's provided and meaningful
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("taxonomy name is required")
	}

	// Validate taxonomy slug to ensure it's provided and meaningful
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("taxonomy slug is required")
	}

	// Check if slug already exists to maintain uniqueness
	exists, err := s.taxonomyRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("taxonomy slug already exists")
	}

	// Create taxonomy entity with the provided parameters
	taxonomy := entities.NewTaxonomy(name, slug, code, description, parentID)

	// Validate the taxonomy entity to ensure proper structure
	if err := taxonomy.Validate(); err != nil {
		return nil, err
	}

	// Persist the taxonomy to the repository
	err = s.taxonomyRepository.Create(ctx, taxonomy)
	if err != nil {
		return nil, err
	}

	return taxonomy, nil
}

// GetTaxonomyByID retrieves a taxonomy by its unique identifier.
// This method provides access to individual taxonomy details and metadata.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the taxonomy to retrieve
//
// Returns:
//   - *entities.Taxonomy: The taxonomy entity if found
//   - error: Error if taxonomy not found or other issues occur
func (s *TaxonomyService) GetTaxonomyByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetByID(ctx, id)
}

// GetTaxonomyBySlug retrieves a taxonomy by its unique slug identifier.
// This method is useful for URL-based taxonomy lookups and routing.
//
// Parameters:
//   - ctx: Context for the operation
//   - slug: Slug identifier of the taxonomy to retrieve
//
// Returns:
//   - *entities.Taxonomy: The taxonomy entity if found
//   - error: Error if taxonomy not found or other issues occur
func (s *TaxonomyService) GetTaxonomyBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetBySlug(ctx, slug)
}

// GetAllTaxonomies retrieves all taxonomies in the system with pagination.
// This method is useful for administrative purposes and taxonomy management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of taxonomies to return
//   - offset: Number of taxonomies to skip for pagination
//
// Returns:
//   - []*entities.Taxonomy: List of all taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetAllTaxonomies(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetAll(ctx, limit, offset)
}

// GetAllTaxonomiesWithSearch retrieves all taxonomies with optional search and pagination.
// This method supports filtering taxonomies by search query while maintaining pagination.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string for filtering
//   - limit: Maximum number of taxonomies to return
//   - offset: Number of taxonomies to skip for pagination
//
// Returns:
//   - []*entities.Taxonomy: List of matching taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetAllTaxonomiesWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetAllWithSearch(ctx, query, limit, offset)
}

// GetRootTaxonomies retrieves all root-level taxonomies (no parent).
// This method is useful for building top-level categorization structures.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Taxonomy: List of root-level taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetRootTaxonomies(ctx)
}

// GetTaxonomyChildren retrieves all direct children of a specific taxonomy.
// This method supports hierarchical navigation and taxonomy tree building.
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent taxonomy
//
// Returns:
//   - []*entities.Taxonomy: List of child taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomyChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetChildren(ctx, parentID)
}

// GetTaxonomyHierarchy retrieves the complete taxonomy hierarchy with parent-child relationships.
// This method is useful for building complete categorization trees and navigation structures.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Taxonomy: Complete taxonomy hierarchy with relationships
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomyHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetHierarchy(ctx)
}

// GetTaxonomyDescendants retrieves all descendants of a specific taxonomy.
// This method supports deep hierarchical navigation and taxonomy tree traversal.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the taxonomy to get descendants for
//
// Returns:
//   - []*entities.Taxonomy: List of descendant taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomyDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetDescendants(ctx, id)
}

// GetTaxonomyAncestors retrieves all ancestors of a specific taxonomy.
// This method supports upward hierarchical navigation and breadcrumb generation.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the taxonomy to get ancestors for
//
// Returns:
//   - []*entities.Taxonomy: List of ancestor taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomyAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetAncestors(ctx, id)
}

// GetTaxonomySiblings retrieves all siblings of a specific taxonomy.
// This method supports lateral navigation within the same hierarchical level.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the taxonomy to get siblings for
//
// Returns:
//   - []*entities.Taxonomy: List of sibling taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomySiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetSiblings(ctx, id)
}

// SearchTaxonomies searches for taxonomies based on a query string.
// This method supports full-text search capabilities for finding taxonomies
// by name, description, code, or other attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Taxonomy: List of matching taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) SearchTaxonomies(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.Search(ctx, query, limit, offset)
}

// UpdateTaxonomy updates an existing taxonomy's information and metadata.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Taxonomy must exist and be accessible
//   - Taxonomy name and slug are required and validated
//   - Slug must be unique (excluding current taxonomy)
//   - Taxonomy validation ensures proper structure
//   - Hierarchical relationships are maintained
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the taxonomy to update
//   - name: Updated display name of the taxonomy
//   - slug: Updated unique identifier for the taxonomy
//   - code: Updated programmatic code for the taxonomy
//   - description: Updated description of the taxonomy
//   - parentID: Updated parent taxonomy ID for hierarchical structure
//
// Returns:
//   - *entities.Taxonomy: The updated taxonomy entity
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) UpdateTaxonomy(ctx context.Context, id uuid.UUID, name, slug, code, description string, parentID *uuid.UUID) (*entities.Taxonomy, error) {
	// Retrieve existing taxonomy to ensure it exists and is accessible
	taxonomy, err := s.taxonomyRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate updated taxonomy name to ensure it's provided and meaningful
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("taxonomy name is required")
	}

	// Validate updated taxonomy slug to ensure it's provided and meaningful
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("taxonomy slug is required")
	}

	// Check if updated slug already exists (excluding current taxonomy)
	if slug != taxonomy.Slug {
		exists, err := s.taxonomyRepository.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("taxonomy slug already exists")
		}
	}

	// Update the taxonomy entity with new information
	taxonomy.UpdateTaxonomy(name, slug, code, description, parentID)

	// Persist the updated taxonomy to the repository
	err = s.taxonomyRepository.Update(ctx, taxonomy)
	if err != nil {
		return nil, err
	}

	return taxonomy, nil
}

// DeleteTaxonomy performs a soft delete of a taxonomy by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the taxonomy to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) DeleteTaxonomy(ctx context.Context, id uuid.UUID) error {
	return s.taxonomyRepository.Delete(ctx, id)
}

// GetTaxonomyCount returns the total number of taxonomies in the system.
// This method is useful for statistics and administrative reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomyCount(ctx context.Context) (int64, error) {
	return s.taxonomyRepository.Count(ctx)
}

// GetTaxonomyCountWithSearch returns the total number of taxonomies matching a search query.
// This method is useful for pagination calculations and search result statistics.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//
// Returns:
//   - int64: Total count of matching taxonomies
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) GetTaxonomyCountWithSearch(ctx context.Context, query string) (int64, error) {
	return s.taxonomyRepository.CountWithSearch(ctx, query)
}

// SearchTaxonomiesWithPagination performs a unified search with pagination and returns
// a structured response with metadata. This method provides a comprehensive search
// experience with proper pagination handling.
//
// Business Rules:
//   - Search request must be validated
//   - Pagination parameters are properly handled
//   - Total count is calculated for pagination metadata
//   - Structured response includes search results and metadata
//
// Parameters:
//   - ctx: Context for the operation
//   - request: Structured search request with pagination parameters
//
// Returns:
//   - *pagination.TaxonomySearchResponse: Structured response with results and metadata
//   - error: Any error that occurred during the operation
func (s *TaxonomyService) SearchTaxonomiesWithPagination(ctx context.Context, request *pagination.TaxonomySearchRequest) (*pagination.TaxonomySearchResponse, error) {
	// Validate search request to ensure proper parameters
	if err := pagination.ValidateTaxonomySearchRequest(request); err != nil {
		return nil, err
	}

	// Get taxonomies with search and pagination
	taxonomies, err := s.GetAllTaxonomiesWithSearch(ctx, request.Query, request.GetLimit(), request.GetOffset())
	if err != nil {
		return nil, err
	}

	// Get total count for pagination metadata
	totalCount, err := s.GetTaxonomyCountWithSearch(ctx, request.Query)
	if err != nil {
		return nil, err
	}

	// Create unified response with results and metadata
	searchResponse := pagination.CreateTaxonomySearchResponse(taxonomies, request.Page, request.PerPage, totalCount)

	return searchResponse, nil
}
