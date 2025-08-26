package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// TaxonomyResource represents a single taxonomy in API responses
// Following Laravel API Resource pattern for consistent formatting
type TaxonomyResource struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Code        string             `json:"code,omitempty"`
	Description string             `json:"description,omitempty"`
	ParentID    *string            `json:"parent_id,omitempty"`
	RecordLeft  *uint64            `json:"record_left,omitempty"`
	RecordRight *uint64            `json:"record_right,omitempty"`
	RecordDepth *uint64            `json:"record_depth,omitempty"`
	IsRoot      bool               `json:"is_root"`
	HasChildren bool               `json:"has_children"`
	HasParent   bool               `json:"has_parent"`
	Level       int                `json:"level"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty"`
	Parent      *TaxonomyResource  `json:"parent,omitempty"`
	Children    []TaxonomyResource `json:"children,omitempty"`
}

// TaxonomyCollection represents a collection of taxonomies
// Following Laravel API Resource Collection pattern
type TaxonomyCollection struct {
	Data  []TaxonomyResource `json:"data"`
	Meta  *CollectionMeta    `json:"meta,omitempty"`
	Links *CollectionLinks   `json:"links,omitempty"`
}

// TaxonomyResourceResponse represents a single taxonomy response with Laravel-style formatting
type TaxonomyResourceResponse struct {
	Status string           `json:"status"`
	Data   TaxonomyResource `json:"data"`
}

// TaxonomyCollectionResponse represents a collection response with Laravel-style formatting
type TaxonomyCollectionResponse struct {
	Status string             `json:"status"`
	Data   TaxonomyCollection `json:"data"`
}

// NewTaxonomyResource creates a new TaxonomyResource from a Taxonomy entity
// This transforms the domain entity into a consistent API response format
func NewTaxonomyResource(taxonomy *entities.Taxonomy) *TaxonomyResource {
	var parentID *string
	if taxonomy.ParentID != nil {
		parentIDStr := taxonomy.ParentID.String()
		parentID = &parentIDStr
	}

	// Transform parent taxonomy if exists
	var parentResource *TaxonomyResource
	if taxonomy.Parent != nil {
		parentResource = NewTaxonomyResource(taxonomy.Parent)
	}

	// Transform children taxonomies if exist
	var childrenResources []TaxonomyResource
	if taxonomy.Children != nil && len(taxonomy.Children) > 0 {
		childrenResources = make([]TaxonomyResource, len(taxonomy.Children))
		for i, child := range taxonomy.Children {
			childrenResources[i] = *NewTaxonomyResource(child)
		}
	}

	// Calculate computed fields
	level := 0
	if taxonomy.RecordDepth != nil {
		level = int(*taxonomy.RecordDepth)
	}

	return &TaxonomyResource{
		ID:          taxonomy.ID.String(),
		Name:        taxonomy.Name,
		Slug:        taxonomy.Slug,
		Code:        taxonomy.Code,
		Description: taxonomy.Description,
		ParentID:    parentID,
		RecordLeft:  taxonomy.RecordLeft,
		RecordRight: taxonomy.RecordRight,
		RecordDepth: taxonomy.RecordDepth,
		IsRoot:      taxonomy.ParentID == nil,
		HasChildren: taxonomy.Children != nil && len(taxonomy.Children) > 0,
		HasParent:   taxonomy.ParentID != nil,
		Level:       level,
		CreatedAt:   taxonomy.CreatedAt,
		UpdatedAt:   taxonomy.UpdatedAt,
		DeletedAt:   taxonomy.DeletedAt,
		Parent:      parentResource,
		Children:    childrenResources,
	}
}

// NewTaxonomyCollection creates a new TaxonomyCollection from a slice of Taxonomy entities
// This transforms multiple domain entities into a consistent API response format
func NewTaxonomyCollection(taxonomies []*entities.Taxonomy) *TaxonomyCollection {
	taxonomyResources := make([]TaxonomyResource, len(taxonomies))
	for i, taxonomy := range taxonomies {
		taxonomyResources[i] = *NewTaxonomyResource(taxonomy)
	}

	return &TaxonomyCollection{
		Data: taxonomyResources,
	}
}

// NewPaginatedTaxonomyCollection creates a new TaxonomyCollection with pagination metadata
// This follows Laravel's paginated resource collection pattern
func NewPaginatedTaxonomyCollection(
	taxonomies []*entities.Taxonomy,
	page, perPage int,
	total int64,
	baseURL string,
) *TaxonomyCollection {
	collection := NewTaxonomyCollection(taxonomies)

	totalPages := (int(total) + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	from := (page-1)*perPage + 1
	to := page * perPage
	if to > int(total) {
		to = int(total)
	}

	collection.Meta = &CollectionMeta{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   total,
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		From:         from,
		To:           to,
	}

	// Generate pagination links
	collection.Links = &CollectionLinks{
		First: generatePageURL(baseURL, 1),
		Last:  generatePageURL(baseURL, totalPages),
		Prev:  generatePageURL(baseURL, page-1),
		Next:  generatePageURL(baseURL, page+1),
	}

	return collection
}

// NewTaxonomyResourceResponse creates a new single taxonomy response
func NewTaxonomyResourceResponse(taxonomy *entities.Taxonomy) *TaxonomyResourceResponse {
	return &TaxonomyResourceResponse{
		Status: "success",
		Data:   *NewTaxonomyResource(taxonomy),
	}
}

// NewTaxonomyCollectionResponse creates a new taxonomy collection response
func NewTaxonomyCollectionResponse(taxonomies []*entities.Taxonomy) *TaxonomyCollectionResponse {
	return &TaxonomyCollectionResponse{
		Status: "success",
		Data:   *NewTaxonomyCollection(taxonomies),
	}
}

// NewPaginatedTaxonomyCollectionResponse creates a new paginated taxonomy collection response
func NewPaginatedTaxonomyCollectionResponse(
	taxonomies []*entities.Taxonomy,
	page, perPage int,
	total int64,
	baseURL string,
) *TaxonomyCollectionResponse {
	return &TaxonomyCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedTaxonomyCollection(taxonomies, page, perPage, total, baseURL),
	}
}
