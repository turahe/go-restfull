// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CommentResource represents a comment in API responses.
// This struct provides a comprehensive view of a comment including its content,
// author information, hierarchical structure (for nested comments), and status.
// It follows the Laravel API Resource pattern for consistent formatting.
type CommentResource struct {
	// ID is the unique identifier for the comment
	ID string `json:"id"`
	// ModelID is the ID of the entity this comment belongs to (e.g., post, article)
	ModelID string `json:"model_id"`
	// ModelType is the type of entity this comment belongs to (e.g., "Post", "Article")
	ModelType string `json:"model_type"`
	// ParentID is the optional ID of the parent comment for nested replies
	ParentID *string `json:"parent_id,omitempty"`
	// Status indicates the current status of the comment (e.g., "approved", "pending", "rejected")
	Status string `json:"status"`
	// Content contains the comment's text content in both raw and HTML formats
	Content *ContentResource `json:"content,omitempty"`
	// Author contains information about the user who wrote the comment
	Author *UserResource `json:"author,omitempty"`
	// RecordLeft is used for nested set model operations (tree structure)
	RecordLeft *int64 `json:"record_left,omitempty"`
	// RecordRight is used for nested set model operations (tree structure)
	RecordRight *int64 `json:"record_right,omitempty"`
	// RecordOrdering determines the display order of comments
	RecordOrdering *int64 `json:"record_ordering,omitempty"`
	// RecordDepth indicates the nesting level of the comment in the tree
	RecordDepth *int64 `json:"record_depth,omitempty"`
	// CreatedBy is the ID of the user who created the comment
	CreatedBy string `json:"created_by"`
	// UpdatedBy is the ID of the user who last updated the comment
	UpdatedBy string `json:"updated_by"`
	// DeletedBy is the optional ID of the user who deleted the comment
	DeletedBy *string `json:"deleted_by,omitempty"`
	// CreatedAt is the timestamp when the comment was created
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the comment was last updated
	UpdatedAt string `json:"updated_at"`
	// DeletedAt is the optional timestamp when the comment was soft-deleted
	DeletedAt *string `json:"deleted_at,omitempty"`

	// Computed fields for easy status checking
	// IsReply indicates whether this comment is a reply to another comment
	IsReply bool `json:"is_reply"`
	// IsApproved indicates whether the comment has been approved for display
	IsApproved bool `json:"is_approved"`
	// IsPending indicates whether the comment is awaiting approval
	IsPending bool `json:"is_pending"`
	// IsRejected indicates whether the comment has been rejected
	IsRejected bool `json:"is_rejected"`
	// IsDeleted indicates whether the comment has been soft-deleted
	IsDeleted bool `json:"is_deleted"`
}

// ContentResource represents content in comment responses.
// This struct contains the comment's text content in multiple formats
// for different display purposes.
type ContentResource struct {
	// ID is the unique identifier for the content
	ID string `json:"id"`
	// ContentRaw is the original, unprocessed text content
	ContentRaw string `json:"content_raw"`
	// ContentHTML is the HTML-formatted version of the content for safe display
	ContentHTML string `json:"content_html"`
}

// CommentCollection represents a collection of comments.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type CommentCollection struct {
	// Data contains the array of comment resources
	Data []CommentResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// CommentResourceResponse represents a single comment response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type CommentResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the comment resource
	Data CommentResource `json:"data"`
}

// CommentCollectionResponse represents a collection of comments response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type CommentCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the comment collection
	Data CommentCollection `json:"data"`
}

// NewCommentResource creates a new CommentResource from comment entity.
// This function transforms the domain entity into a consistent API response format,
// handling optional fields and computed properties appropriately.
//
// Parameters:
//   - comment: The comment domain entity to convert
//   - content: Optional content entity associated with the comment
//   - author: Optional user entity who authored the comment
//
// Returns:
//   - A new CommentResource with all fields properly formatted
func NewCommentResource(comment *entities.Comment, content *entities.Content, author *entities.User) CommentResource {
	resource := CommentResource{
		ID:        comment.ID.String(),
		ModelID:   comment.ModelID.String(),
		ModelType: comment.ModelType,
		Status:    string(comment.Status),
		CreatedBy: comment.CreatedBy.String(),
		UpdatedBy: comment.UpdatedBy.String(),
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Set computed fields based on comment state
		IsReply:    comment.IsReply(),
		IsApproved: comment.IsApproved(),
		IsPending:  comment.IsPending(),
		IsRejected: comment.IsRejected(),
		IsDeleted:  comment.IsDeleted(),
	}

	// Handle optional parent ID for nested comments
	if comment.ParentID != nil {
		parentID := comment.ParentID.String()
		resource.ParentID = &parentID
	}

	// Set nested set model fields if available
	if comment.RecordLeft != nil {
		resource.RecordLeft = comment.RecordLeft
	}

	if comment.RecordRight != nil {
		resource.RecordRight = comment.RecordRight
	}

	if comment.RecordOrdering != nil {
		resource.RecordOrdering = comment.RecordOrdering
	}

	if comment.RecordDepth != nil {
		resource.RecordDepth = comment.RecordDepth
	}

	// Handle soft deletion information
	if comment.DeletedBy != nil {
		deletedBy := comment.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if comment.DeletedAt != nil {
		deletedAt := comment.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	// Add content information if provided
	if content != nil {
		resource.Content = &ContentResource{
			ID:          content.ID.String(),
			ContentRaw:  content.ContentRaw,
			ContentHTML: content.ContentHTML,
		}
	}

	// Add author information if provided
	if author != nil {
		resource.Author = NewUserResource(author)
	}

	return resource
}

// NewCommentResourceResponse creates a new CommentResourceResponse.
// This function wraps a CommentResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - comment: The comment domain entity to convert
//   - content: Optional content entity associated with the comment
//   - author: Optional user entity who authored the comment
//
// Returns:
//   - A new CommentResourceResponse with success status and comment data
func NewCommentResourceResponse(comment *entities.Comment, content *entities.Content, author *entities.User) CommentResourceResponse {
	return CommentResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Comment retrieved successfully",
		Data:            NewCommentResource(comment, content, author),
	}
}

// NewCommentCollection creates a new CommentCollection.
// This function transforms multiple comment domain entities into a consistent
// API response format, efficiently handling associated content and author data.
//
// Parameters:
//   - comments: Slice of comment domain entities to convert
//   - contents: Map of content entities keyed by comment ID
//   - authors: Map of user entities keyed by creator ID
//
// Returns:
//   - A new CommentCollection with all comments properly formatted
func NewCommentCollection(comments []*entities.Comment, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollection {
	commentResources := make([]CommentResource, len(comments))
	for i, comment := range comments {
		var content *entities.Content
		var author *entities.User

		// Look up associated content and author data
		if contents != nil {
			content = contents[comment.ID]
		}
		if authors != nil {
			author = authors[comment.CreatedBy]
		}

		commentResources[i] = NewCommentResource(comment, content, author)
	}

	return CommentCollection{
		Data: commentResources,
	}
}

// NewPaginatedCommentCollection creates a new CommentCollection with pagination.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - comments: Slice of comment domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//   - contents: Map of content entities keyed by comment ID
//   - authors: Map of user entities keyed by creator ID
//
// Returns:
//   - A new paginated CommentCollection with metadata and navigation links
func NewPaginatedCommentCollection(comments []*entities.Comment, page, perPage, total int, baseURL string, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollection {
	collection := NewCommentCollection(comments, contents, authors)

	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	// Set pagination metadata
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

	// Build pagination navigation links
	collection.Links = CollectionLinks{
		First: buildPaginationLink(baseURL, 1, perPage),
		Last:  buildPaginationLink(baseURL, totalPages, perPage),
	}

	// Add previous page link if not on first page
	if page > 1 {
		collection.Links.Prev = buildPaginationLink(baseURL, page-1, perPage)
	}

	// Add next page link if not on last page
	if page < totalPages {
		collection.Links.Next = buildPaginationLink(baseURL, page+1, perPage)
	}

	return collection
}

// NewCommentCollectionResponse creates a new CommentCollectionResponse.
// This function wraps a CommentCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - comments: Slice of comment domain entities to convert
//   - contents: Map of content entities keyed by comment ID
//   - authors: Map of user entities keyed by creator ID
//
// Returns:
//   - A new CommentCollectionResponse with success status and comment collection data
func NewCommentCollectionResponse(comments []*entities.Comment, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollectionResponse {
	return CommentCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Comments retrieved successfully",
		Data:            NewCommentCollection(comments, contents, authors),
	}
}

// NewPaginatedCommentCollectionResponse creates a new CommentCollectionResponse with pagination.
// This function wraps a paginated CommentCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - comments: Slice of comment domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//   - contents: Map of content entities keyed by comment ID
//   - authors: Map of user entities keyed by creator ID
//
// Returns:
//   - A new paginated CommentCollectionResponse with success status and pagination data
func NewPaginatedCommentCollectionResponse(comments []*entities.Comment, page, perPage, total int, baseURL string, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollectionResponse {
	return CommentCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Comments retrieved successfully",
		Data:            NewPaginatedCommentCollection(comments, page, perPage, total, baseURL, contents, authors),
	}
}

// buildPaginationLink is defined in common_responses.go to avoid duplication
