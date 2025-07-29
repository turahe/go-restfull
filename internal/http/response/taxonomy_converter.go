package response

import (
	"math"
	"webapi/internal/domain/entities"
	"webapi/internal/helper/pagination"
)

// ConvertTaxonomyEntityToDTO converts a taxonomy entity to DTO
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
		RecordLeft:  entity.RecordLeft,
		RecordRight: entity.RecordRight,
		RecordDepth: entity.RecordDepth,
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

// ConvertTaxonomyEntitiesToDTOs converts a slice of taxonomy entities to DTOs
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

// CreatePaginationMeta creates pagination metadata
func CreatePaginationMeta(page, perPage int, totalItems int64) pagination.PaginationResponse {
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))

	// Ensure current page doesn't exceed total pages
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	from := (page-1)*perPage + 1
	to := page * perPage
	if to > int(totalItems) {
		to = int(totalItems)
	}
	if totalItems == 0 {
		from = 0
		to = 0
	}

	hasNextPage := page < totalPages
	hasPrevPage := page > 1

	var nextPage, prevPage int
	if hasNextPage {
		nextPage = page + 1
	}
	if hasPrevPage {
		prevPage = page - 1
	}

	return pagination.PaginationResponse{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		HasNextPage:  hasNextPage,
		HasPrevPage:  hasPrevPage,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		From:         from,
		To:           to,
	}
}

// CreateTaxonomySearchResponse creates a unified search response
func CreateTaxonomySearchResponse(taxonomies []*entities.Taxonomy, page, perPage int, totalItems int64) *TaxonomySearchResponse {
	return &TaxonomySearchResponse{
		Data:       ConvertTaxonomyEntitiesToDTOs(taxonomies),
		Pagination: CreatePaginationMeta(page, perPage, totalItems),
	}
}

// CreateTaxonomyHierarchyResponse creates a unified hierarchy response
func CreateTaxonomyHierarchyResponse(taxonomies []*entities.Taxonomy, page, perPage int, totalItems int64) *TaxonomyHierarchyResponse {
	return &TaxonomyHierarchyResponse{
		Data:       ConvertTaxonomyEntitiesToDTOs(taxonomies),
		Pagination: CreatePaginationMeta(page, perPage, totalItems),
	}
}

// CreateTaxonomyDetailResponse creates a unified detail response
func CreateTaxonomyDetailResponse(taxonomy *entities.Taxonomy) *TaxonomyDetailResponse {
	return &TaxonomyDetailResponse{
		Data: ConvertTaxonomyEntityToDTO(taxonomy),
	}
}

// ValidateTaxonomySearchRequest validates the search request parameters
func ValidateTaxonomySearchRequest(req *TaxonomySearchRequest) error {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 10
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	if req.SortBy == "" {
		req.SortBy = "record_left"
	}
	return nil
}

// GetOffset calculates the offset for database queries
func (req *TaxonomySearchRequest) GetOffset() int {
	return (req.Page - 1) * req.PerPage
}

// GetLimit returns the limit for database queries
func (req *TaxonomySearchRequest) GetLimit() int {
	return req.PerPage
}

// GetOrderBy generates the ORDER BY clause for SQL queries
func (req *TaxonomySearchRequest) GetOrderBy() string {
	if req.SortBy == "" {
		return "ORDER BY record_left ASC"
	}

	order := "ORDER BY " + req.SortBy
	if req.SortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	return order
}
