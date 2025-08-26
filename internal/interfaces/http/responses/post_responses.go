package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// PostResource represents a single post in API responses
// Following Laravel API Resource pattern for consistent formatting
type PostResource struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Subtitle    string     `json:"subtitle"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	IsSticky    bool       `json:"is_sticky"`
	Language    string     `json:"language"`
	Layout      string     `json:"layout"`
	Content     string     `json:"content"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	IsPublished bool       `json:"is_published"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// PostCollection represents a collection of posts
// Following Laravel API Resource Collection pattern
type PostCollection struct {
	Data  []PostResource   `json:"data"`
	Meta  *CollectionMeta  `json:"meta,omitempty"`
	Links *CollectionLinks `json:"links,omitempty"`
}

// PostResourceResponse represents a single post response with Laravel-style formatting
type PostResourceResponse struct {
	Status string       `json:"status"`
	Data   PostResource `json:"data"`
}

// PostCollectionResponse represents a collection response with Laravel-style formatting
type PostCollectionResponse struct {
	Status string         `json:"status"`
	Data   PostCollection `json:"data"`
}

// NewPostResource creates a new PostResource from a Post entity
// This transforms the domain entity into a consistent API response format
func NewPostResource(post *entities.Post) *PostResource {
	// Determine status based on published_at
	status := "draft"
	if post.PublishedAt != nil {
		status = "published"
	}

	return &PostResource{
		ID:          post.ID.String(),
		Title:       post.Title,
		Slug:        post.Slug,
		Subtitle:    post.Subtitle,
		Description: post.Description,
		Type:        post.Type,
		IsSticky:    post.IsSticky,
		Language:    post.Language,
		Layout:      post.Layout,
		Content:     post.Content,
		PublishedAt: post.PublishedAt,
		IsPublished: post.PublishedAt != nil,
		Status:      status,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		DeletedAt:   post.DeletedAt,
	}
}

// NewPostCollection creates a new PostCollection from a slice of Post entities
// This transforms multiple domain entities into a consistent API response format
func NewPostCollection(posts []*entities.Post) *PostCollection {
	postResources := make([]PostResource, len(posts))
	for i, post := range posts {
		postResources[i] = *NewPostResource(post)
	}

	return &PostCollection{
		Data: postResources,
	}
}

// NewPaginatedPostCollection creates a new PostCollection with pagination metadata
// This follows Laravel's paginated resource collection pattern
func NewPaginatedPostCollection(
	posts []*entities.Post,
	page, perPage int,
	total int64,
	baseURL string,
) *PostCollection {
	collection := NewPostCollection(posts)

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

// NewPostResourceResponse creates a new single post response
func NewPostResourceResponse(post *entities.Post) *PostResourceResponse {
	return &PostResourceResponse{
		Status: "success",
		Data:   *NewPostResource(post),
	}
}

// NewPostCollectionResponse creates a new post collection response
func NewPostCollectionResponse(posts []*entities.Post) *PostCollectionResponse {
	return &PostCollectionResponse{
		Status: "success",
		Data:   *NewPostCollection(posts),
	}
}

// NewPaginatedPostCollectionResponse creates a new paginated post collection response
func NewPaginatedPostCollectionResponse(
	posts []*entities.Post,
	page, perPage int,
	total int64,
	baseURL string,
) *PostCollectionResponse {
	return &PostCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedPostCollection(posts, page, perPage, total, baseURL),
	}
}
