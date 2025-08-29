// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/pagination"
)

// ConvertTaxonomyEntityToDTO converts a taxonomy entity to DTO.
// This function transforms a domain taxonomy entity into a response DTO,
// including recursive conversion of parent and child relationships.
//
// Parameters:
//   - entity: The domain taxonomy entity to convert
//
// Returns:
//   - A new TaxonomyDTO with all fields populated from the entity, or nil if entity is nil
func ConvertTaxonomyEntityToDTO(entity *entities.Taxonomy) *TaxonomyDTO {
	if entity == nil {
		return nil
	}

	taxonomyDTO := &TaxonomyDTO{
		ID:          entity.ID,
		Name:        entity.Name,
		Slug:        entity.Slug,
		Code:        entity.Code,
		Description: entity.Description,
		ParentID:    entity.ParentID,
		RecordLeft:  *entity.RecordLeft,
		RecordRight: *entity.RecordRight,
		RecordDepth: *entity.RecordDepth,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
		DeletedAt:   entity.DeletedAt,
	}

	// Convert parent if exists
	if entity.Parent != nil {
		taxonomyDTO.Parent = ConvertTaxonomyEntityToDTO(entity.Parent)
	}

	// Convert children if exists
	if len(entity.Children) > 0 {
		taxonomyDTO.Children = make([]*TaxonomyDTO, len(entity.Children))
		for i, child := range entity.Children {
			taxonomyDTO.Children[i] = ConvertTaxonomyEntityToDTO(child)
		}
	}

	return taxonomyDTO
}

// ConvertTaxonomyEntitiesToDTOs converts a slice of taxonomy entities to DTOs.
// This function transforms multiple domain taxonomy entities into response DTOs
// by calling ConvertTaxonomyEntityToDTO for each entity.
//
// Parameters:
//   - entities: Slice of domain taxonomy entities to convert
//
// Returns:
//   - A slice of TaxonomyDTOs with all entities converted, or nil if entities is nil
func ConvertTaxonomyEntitiesToDTOs(entities []*entities.Taxonomy) []*TaxonomyDTO {
	if entities == nil {
		return nil
	}

	dtos := make([]*TaxonomyDTO, len(entities))
	for i, entity := range entities {
		dtos[i] = ConvertTaxonomyEntityToDTO(entity)
	}

	return dtos
}

// CreatePaginationMeta creates pagination metadata.
// This function uses the centralized pagination utility to avoid code duplication.
//
// Parameters:
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - totalItems: Total number of items across all pages
//
// Returns:
//   - A PaginationResponse with complete pagination metadata
func CreatePaginationMeta(page, perPage int, totalItems int64) pagination.PaginationResponse {
	// Use the centralized pagination utility
	meta := CreateCollectionMeta(page, perPage, int(totalItems))

	// Convert to pagination.PaginationResponse format
	return pagination.PaginationResponse{
		CurrentPage:  meta.CurrentPage,
		PerPage:      meta.PerPage,
		TotalItems:   meta.TotalItems,
		TotalPages:   meta.TotalPages,
		HasNextPage:  meta.HasNextPage,
		HasPrevPage:  meta.HasPrevPage,
		NextPage:     meta.NextPage,
		PreviousPage: meta.PreviousPage,
		From:         meta.From,
		To:           meta.To,
	}
}

// CreateTaxonomySearchResponse creates a unified search response.
// This function combines taxonomy search results with pagination metadata
// to create a complete search response structure.
//
// Parameters:
//   - taxonomies: Slice of taxonomy entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - totalItems: Total number of items across all pages
//
// Returns:
//   - A complete TaxonomySearchResponse with data and pagination
func CreateTaxonomySearchResponse(taxonomies []*entities.Taxonomy, page, perPage int, totalItems int64) *TaxonomySearchResponse {
	return &TaxonomySearchResponse{
		Data:       ConvertTaxonomyEntitiesToDTOs(taxonomies),
		Pagination: CreatePaginationMeta(page, perPage, totalItems),
	}
}

// CreateTaxonomyHierarchyResponse creates a unified hierarchy response.
// This function combines hierarchical taxonomy data with pagination metadata
// to create a complete hierarchy response structure.
//
// Parameters:
//   - taxonomies: Slice of taxonomy entities organized in hierarchy order
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - totalItems: Total number of items across all pages
//
// Returns:
//   - A complete TaxonomyHierarchyResponse with hierarchical data and pagination
func CreateTaxonomyHierarchyResponse(taxonomies []*entities.Taxonomy, page, perPage int, totalItems int64) *TaxonomyHierarchyResponse {
	return &TaxonomyHierarchyResponse{
		Data:       ConvertTaxonomyEntitiesToDTOs(taxonomies),
		Pagination: CreatePaginationMeta(page, perPage, totalItems),
	}
}

// CreateTaxonomyDetailResponse creates a unified detail response.
// This function creates a response structure for a single taxonomy
// with complete details and relationships.
//
// Parameters:
//   - taxonomy: The taxonomy entity to create a detail response for
//
// Returns:
//   - A complete TaxonomyDetailResponse with the taxonomy details
func CreateTaxonomyDetailResponse(taxonomy *entities.Taxonomy) *TaxonomyDetailResponse {
	return &TaxonomyDetailResponse{
		Data: ConvertTaxonomyEntityToDTO(taxonomy),
	}
}

// ValidateTaxonomySearchRequest validates the search request parameters.
// This function ensures that search request parameters are within valid ranges
// and sets sensible defaults for missing or invalid values.
//
// Parameters:
//   - req: The search request to validate and normalize
//
// Returns:
//   - An error if validation fails, nil otherwise
func ValidateTaxonomySearchRequest(req *TaxonomySearchRequest) error {
	// Ensure page is at least 1
	if req.Page < 1 {
		req.Page = 1
	}
	// Ensure perPage is at least 1
	if req.PerPage < 1 {
		req.PerPage = 10
	}
	// Limit perPage to reasonable maximum
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	// Set default sort field if not specified
	if req.SortBy == "" {
		req.SortBy = "record_left"
	}
	return nil
}

// GetOffset calculates the offset for database queries.
// This method calculates the SQL OFFSET value based on the current page
// and items per page for pagination queries.
//
// Returns:
//   - The calculated offset value for database queries
func (req *TaxonomySearchRequest) GetOffset() int {
	return (req.Page - 1) * req.PerPage
}

// GetLimit returns the limit for database queries.
// This method returns the number of items per page as the SQL LIMIT value.
//
// Returns:
//   - The limit value for database queries
func (req *TaxonomySearchRequest) GetLimit() int {
	return req.PerPage
}

// GetOrderBy generates the ORDER BY clause for SQL queries.
// This method constructs a complete ORDER BY clause based on the sort field
// and sort direction preferences.
//
// Returns:
//   - A complete ORDER BY clause string for SQL queries
func (req *TaxonomySearchRequest) GetOrderBy() string {
	// Use default sort field if none specified
	if req.SortBy == "" {
		return "ORDER BY record_left ASC"
	}

	// Build ORDER BY clause with specified field and direction
	order := "ORDER BY " + req.SortBy
	if req.SortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	return order
}
