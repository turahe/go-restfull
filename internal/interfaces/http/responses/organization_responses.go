package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// OrganizationResource represents a single organization in API responses
// Following Laravel API Resource pattern for consistent formatting
type OrganizationResource struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	Code           *string                `json:"code,omitempty"`
	Type           *string                `json:"type,omitempty"`
	Status         string                 `json:"status"`
	ParentID       *string                `json:"parent_id,omitempty"`
	RecordLeft     *uint64                `json:"record_left,omitempty"`
	RecordRight    *uint64                `json:"record_right,omitempty"`
	RecordDepth    *uint64                `json:"record_depth,omitempty"`
	RecordOrdering *uint64                `json:"record_ordering,omitempty"`
	IsRoot         bool                   `json:"is_root"`
	HasChildren    bool                   `json:"has_children"`
	HasParent      bool                   `json:"has_parent"`
	Level          int                    `json:"level"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
	Parent         *OrganizationResource  `json:"parent,omitempty"`
	Children       []OrganizationResource `json:"children,omitempty"`
}

// OrganizationCollection represents a collection of organizations
// Following Laravel API Resource Collection pattern
type OrganizationCollection struct {
	Data  []OrganizationResource `json:"data"`
	Meta  *CollectionMeta        `json:"meta,omitempty"`
	Links *CollectionLinks       `json:"links,omitempty"`
}

// OrganizationResourceResponse represents a single organization response with Laravel-style formatting
type OrganizationResourceResponse struct {
	Status string               `json:"status"`
	Data   OrganizationResource `json:"data"`
}

// OrganizationCollectionResponse represents a collection response with Laravel-style formatting
type OrganizationCollectionResponse struct {
	Status string                 `json:"status"`
	Data   OrganizationCollection `json:"data"`
}

// NewOrganizationResource creates a new OrganizationResource from an Organization entity
// This transforms the domain entity into a consistent API response format
func NewOrganizationResource(org *entities.Organization) *OrganizationResource {
	var parentID *string
	if org.ParentID != nil {
		parentIDStr := org.ParentID.String()
		parentID = &parentIDStr
	}

	var orgType *string
	if org.Type != nil {
		typeStr := string(*org.Type)
		orgType = &typeStr
	}

	// Transform parent organization if exists
	var parentResource *OrganizationResource
	if org.Parent != nil {
		parentResource = NewOrganizationResource(org.Parent)
	}

	// Transform children organizations if exist
	var childrenResources []OrganizationResource
	if org.Children != nil && len(org.Children) > 0 {
		childrenResources = make([]OrganizationResource, len(org.Children))
		for i, child := range org.Children {
			childrenResources[i] = *NewOrganizationResource(child)
		}
	}

	// Calculate computed fields
	level := 0
	if org.RecordDepth != nil {
		level = int(*org.RecordDepth)
	}

	return &OrganizationResource{
		ID:             org.ID.String(),
		Name:           org.Name,
		Description:    org.Description,
		Code:           org.Code,
		Type:           orgType,
		Status:         string(org.Status),
		ParentID:       parentID,
		RecordLeft:     org.RecordLeft,
		RecordRight:    org.RecordRight,
		RecordDepth:    org.RecordDepth,
		RecordOrdering: org.RecordOrdering,
		IsRoot:         org.ParentID == nil,
		HasChildren:    org.Children != nil && len(org.Children) > 0,
		HasParent:      org.ParentID != nil,
		Level:          level,
		CreatedAt:      org.CreatedAt,
		UpdatedAt:      org.UpdatedAt,
		DeletedAt:      org.DeletedAt,
		Parent:         parentResource,
		Children:       childrenResources,
	}
}

// NewOrganizationCollection creates a new OrganizationCollection from a slice of Organization entities
// This transforms multiple domain entities into a consistent API response format
func NewOrganizationCollection(organizations []*entities.Organization) *OrganizationCollection {
	orgResources := make([]OrganizationResource, len(organizations))
	for i, org := range organizations {
		orgResources[i] = *NewOrganizationResource(org)
	}

	return &OrganizationCollection{
		Data: orgResources,
	}
}

// NewPaginatedOrganizationCollection creates a new OrganizationCollection with pagination metadata
// This follows Laravel's paginated resource collection pattern
func NewPaginatedOrganizationCollection(
	organizations []*entities.Organization,
	page, perPage int,
	total int64,
	baseURL string,
) *OrganizationCollection {
	collection := NewOrganizationCollection(organizations)

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

// NewOrganizationResourceResponse creates a new single organization response
func NewOrganizationResourceResponse(org *entities.Organization) *OrganizationResourceResponse {
	return &OrganizationResourceResponse{
		Status: "success",
		Data:   *NewOrganizationResource(org),
	}
}

// NewOrganizationCollectionResponse creates a new organization collection response
func NewOrganizationCollectionResponse(organizations []*entities.Organization) *OrganizationCollectionResponse {
	return &OrganizationCollectionResponse{
		Status: "success",
		Data:   *NewOrganizationCollection(organizations),
	}
}

// NewPaginatedOrganizationCollectionResponse creates a new paginated organization collection response
func NewPaginatedOrganizationCollectionResponse(
	organizations []*entities.Organization,
	page, perPage int,
	total int64,
	baseURL string,
) *OrganizationCollectionResponse {
	return &OrganizationCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedOrganizationCollection(organizations, page, perPage, total, baseURL),
	}
}
