package pagination

import (
	"time"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// TaxonomySearchRequest represents the request for searching taxonomies
type TaxonomySearchRequest struct {
	Query    string `json:"query" query:"query"`
	Page     int    `json:"page" query:"page"`
	PerPage  int    `json:"per_page" query:"per_page"`
	SortBy   string `json:"sort_by" query:"sort_by"`
	SortDesc bool   `json:"sort_desc" query:"sort_desc"`
}

// TaxonomyDTO represents a taxonomy in API responses
type TaxonomyDTO struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Code        string     `json:"code,omitempty"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	RecordLeft  int64      `json:"record_left"`
	RecordRight int64      `json:"record_right"`
	RecordDepth int64      `json:"record_depth"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	Parent   *TaxonomyDTO   `json:"parent,omitempty"`
	Children []*TaxonomyDTO `json:"children,omitempty"`
}

// TaxonomySearchResponse represents the response for taxonomy search with pagination
type TaxonomySearchResponse struct {
	Data       []*TaxonomyDTO     `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
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

// CreateTaxonomySearchResponse creates a unified search response
func CreateTaxonomySearchResponse(taxonomies []*entities.Taxonomy, page, perPage int, totalItems int64) *TaxonomySearchResponse {
	return &TaxonomySearchResponse{
		Data:       ConvertTaxonomyEntitiesToDTOs(taxonomies),
		Pagination: CreatePaginationResponse(&PaginationRequest{Page: page, PerPage: perPage}, totalItems),
	}
}
