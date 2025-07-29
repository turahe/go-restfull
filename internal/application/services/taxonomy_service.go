package services

import (
	"context"
	"errors"
	"strings"
	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/dto"

	"github.com/google/uuid"
)

// TaxonomyService implements the TaxonomyService interface
type TaxonomyService struct {
	taxonomyRepository repositories.TaxonomyRepository
}

// NewTaxonomyService creates a new taxonomy service
func NewTaxonomyService(taxonomyRepository repositories.TaxonomyRepository) ports.TaxonomyService {
	return &TaxonomyService{
		taxonomyRepository: taxonomyRepository,
	}
}

// CreateTaxonomy creates a new taxonomy
func (s *TaxonomyService) CreateTaxonomy(ctx context.Context, name, slug, code, description string, parentID *uuid.UUID) (*entities.Taxonomy, error) {
	// Validate input
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("taxonomy name is required")
	}

	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("taxonomy slug is required")
	}

	// Check if slug already exists
	exists, err := s.taxonomyRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("taxonomy slug already exists")
	}

	// Create taxonomy entity
	taxonomy := entities.NewTaxonomy(name, slug, code, description, parentID)

	// Validate taxonomy
	if err := taxonomy.Validate(); err != nil {
		return nil, err
	}

	// Save to repository
	err = s.taxonomyRepository.Create(ctx, taxonomy)
	if err != nil {
		return nil, err
	}

	return taxonomy, nil
}

// GetTaxonomyByID retrieves a taxonomy by ID
func (s *TaxonomyService) GetTaxonomyByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetByID(ctx, id)
}

// GetTaxonomyBySlug retrieves a taxonomy by slug
func (s *TaxonomyService) GetTaxonomyBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetBySlug(ctx, slug)
}

// GetAllTaxonomies retrieves all taxonomies with pagination
func (s *TaxonomyService) GetAllTaxonomies(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetAll(ctx, limit, offset)
}

// GetAllTaxonomiesWithSearch retrieves all taxonomies with optional search and pagination
func (s *TaxonomyService) GetAllTaxonomiesWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetAllWithSearch(ctx, query, limit, offset)
}

// GetRootTaxonomies retrieves root taxonomies (no parent)
func (s *TaxonomyService) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetRootTaxonomies(ctx)
}

// GetTaxonomyChildren retrieves children of a taxonomy
func (s *TaxonomyService) GetTaxonomyChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetChildren(ctx, parentID)
}

// GetTaxonomyHierarchy retrieves the complete taxonomy hierarchy
func (s *TaxonomyService) GetTaxonomyHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetHierarchy(ctx)
}

// GetTaxonomyDescendants retrieves all descendants of a taxonomy
func (s *TaxonomyService) GetTaxonomyDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetDescendants(ctx, id)
}

// GetTaxonomyAncestors retrieves all ancestors of a taxonomy
func (s *TaxonomyService) GetTaxonomyAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetAncestors(ctx, id)
}

// GetTaxonomySiblings retrieves siblings of a taxonomy
func (s *TaxonomyService) GetTaxonomySiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.GetSiblings(ctx, id)
}

// SearchTaxonomies searches taxonomies by query
func (s *TaxonomyService) SearchTaxonomies(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	return s.taxonomyRepository.Search(ctx, query, limit, offset)
}

// UpdateTaxonomy updates a taxonomy
func (s *TaxonomyService) UpdateTaxonomy(ctx context.Context, id uuid.UUID, name, slug, code, description string, parentID *uuid.UUID) (*entities.Taxonomy, error) {
	// Get existing taxonomy
	taxonomy, err := s.taxonomyRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate input
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("taxonomy name is required")
	}

	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("taxonomy slug is required")
	}

	// Check if slug already exists (excluding current taxonomy)
	if slug != taxonomy.Slug {
		exists, err := s.taxonomyRepository.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("taxonomy slug already exists")
		}
	}

	// Update taxonomy
	taxonomy.UpdateTaxonomy(name, slug, code, description, parentID)

	// Save to repository
	err = s.taxonomyRepository.Update(ctx, taxonomy)
	if err != nil {
		return nil, err
	}

	return taxonomy, nil
}

// DeleteTaxonomy deletes a taxonomy
func (s *TaxonomyService) DeleteTaxonomy(ctx context.Context, id uuid.UUID) error {
	return s.taxonomyRepository.Delete(ctx, id)
}

// GetTaxonomyCount retrieves the total count of taxonomies
func (s *TaxonomyService) GetTaxonomyCount(ctx context.Context) (int64, error) {
	return s.taxonomyRepository.Count(ctx)
}

// GetTaxonomyCountWithSearch retrieves the total count of taxonomies with optional search
func (s *TaxonomyService) GetTaxonomyCountWithSearch(ctx context.Context, query string) (int64, error) {
	return s.taxonomyRepository.CountWithSearch(ctx, query)
}

// SearchTaxonomiesWithPagination performs a unified search with pagination and returns a structured response
func (s *TaxonomyService) SearchTaxonomiesWithPagination(ctx context.Context, request *dto.TaxonomySearchRequest) (*dto.TaxonomySearchResponse, error) {
	// Validate request
	if err := dto.ValidateTaxonomySearchRequest(request); err != nil {
		return nil, err
	}

	// Get taxonomies with search
	taxonomies, err := s.GetAllTaxonomiesWithSearch(ctx, request.Query, request.GetLimit(), request.GetOffset())
	if err != nil {
		return nil, err
	}

	// Get total count with search
	totalCount, err := s.GetTaxonomyCountWithSearch(ctx, request.Query)
	if err != nil {
		return nil, err
	}

	// Create unified response
	response := dto.CreateTaxonomySearchResponse(taxonomies, request.Page, request.PerPage, totalCount)

	return response, nil
}
