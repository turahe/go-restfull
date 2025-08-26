package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// TagResource represents a tag in API responses
type TagResource struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	CreatedBy   string  `json:"created_by"`
	UpdatedBy   string  `json:"updated_by"`
	DeletedBy   *string `json:"deleted_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`

	// Computed fields
	IsDeleted bool `json:"is_deleted"`
}

// TagCollection represents a collection of tags
type TagCollection struct {
	Data  []TagResource   `json:"data"`
	Meta  CollectionMeta  `json:"meta"`
	Links CollectionLinks `json:"links"`
}

// TagResourceResponse represents a single tag response
type TagResourceResponse struct {
	ResponseCode    int         `json:"response_code"`
	ResponseMessage string      `json:"response_message"`
	Data            TagResource `json:"data"`
}

// TagCollectionResponse represents a collection of tags response
type TagCollectionResponse struct {
	ResponseCode    int           `json:"response_code"`
	ResponseMessage string        `json:"response_message"`
	Data            TagCollection `json:"data"`
}

// NewTagResource creates a new TagResource from tag entity
func NewTagResource(tag *entities.Tag) TagResource {
	resource := TagResource{
		ID:          tag.ID.String(),
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		CreatedBy:   tag.CreatedBy.String(),
		UpdatedBy:   tag.UpdatedBy.String(),
		CreatedAt:   tag.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   tag.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Computed fields
		IsDeleted: tag.IsDeleted(),
	}

	// Set optional fields
	if tag.DeletedBy != nil {
		deletedBy := tag.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if tag.DeletedAt != nil {
		deletedAt := tag.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	return resource
}

// NewTagResourceResponse creates a new TagResourceResponse
func NewTagResourceResponse(tag *entities.Tag) TagResourceResponse {
	return TagResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Tag retrieved successfully",
		Data:            NewTagResource(tag),
	}
}

// NewTagCollection creates a new TagCollection
func NewTagCollection(tags []*entities.Tag) TagCollection {
	tagResources := make([]TagResource, len(tags))
	for i, tag := range tags {
		tagResources[i] = NewTagResource(tag)
	}

	return TagCollection{
		Data: tagResources,
	}
}

// NewPaginatedTagCollection creates a new TagCollection with pagination
func NewPaginatedTagCollection(tags []*entities.Tag, page, perPage, total int, baseURL string) TagCollection {
	collection := NewTagCollection(tags)

	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	collection.Meta = CollectionMeta{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   int64(total),
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		From:         from,
		To:           to,
	}

	// Build pagination links
	collection.Links = CollectionLinks{
		First: buildPaginationLink(baseURL, 1, perPage),
		Last:  buildPaginationLink(baseURL, totalPages, perPage),
	}

	if page > 1 {
		collection.Links.Prev = buildPaginationLink(baseURL, page-1, perPage)
	}

	if page < totalPages {
		collection.Links.Next = buildPaginationLink(baseURL, page+1, perPage)
	}

	return collection
}

// NewTagCollectionResponse creates a new TagCollectionResponse
func NewTagCollectionResponse(tags []*entities.Tag) TagCollectionResponse {
	return TagCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Tags retrieved successfully",
		Data:            NewTagCollection(tags),
	}
}

// NewPaginatedTagCollectionResponse creates a new TagCollectionResponse with pagination
func NewPaginatedTagCollectionResponse(tags []*entities.Tag, page, perPage, total int, baseURL string) TagCollectionResponse {
	return TagCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Tags retrieved successfully",
		Data:            NewPaginatedTagCollection(tags, page, perPage, total, baseURL),
	}
}
