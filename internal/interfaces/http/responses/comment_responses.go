package responses

import (
	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CommentResource represents a comment in API responses
type CommentResource struct {
	ID             string           `json:"id"`
	ModelID        string           `json:"model_id"`
	ModelType      string           `json:"model_type"`
	ParentID       *string          `json:"parent_id,omitempty"`
	Status         string           `json:"status"`
	Content        *ContentResource `json:"content,omitempty"`
	Author         *UserResource    `json:"author,omitempty"`
	RecordLeft     *uint64          `json:"record_left,omitempty"`
	RecordRight    *uint64          `json:"record_right,omitempty"`
	RecordOrdering *uint64          `json:"record_ordering,omitempty"`
	RecordDepth    *uint64          `json:"record_depth,omitempty"`
	CreatedBy      string           `json:"created_by"`
	UpdatedBy      string           `json:"updated_by"`
	DeletedBy      *string          `json:"deleted_by,omitempty"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
	DeletedAt      *string          `json:"deleted_at,omitempty"`

	// Computed fields
	IsReply    bool `json:"is_reply"`
	IsApproved bool `json:"is_approved"`
	IsPending  bool `json:"is_pending"`
	IsRejected bool `json:"is_rejected"`
	IsDeleted  bool `json:"is_deleted"`
}

// ContentResource represents content in comment responses
type ContentResource struct {
	ID          string `json:"id"`
	ContentRaw  string `json:"content_raw"`
	ContentHTML string `json:"content_html"`
}

// CommentCollection represents a collection of comments
type CommentCollection struct {
	Data  []CommentResource `json:"data"`
	Meta  CollectionMeta    `json:"meta"`
	Links CollectionLinks   `json:"links"`
}

// CommentResourceResponse represents a single comment response
type CommentResourceResponse struct {
	ResponseCode    int             `json:"response_code"`
	ResponseMessage string          `json:"response_message"`
	Data            CommentResource `json:"data"`
}

// CommentCollectionResponse represents a collection of comments response
type CommentCollectionResponse struct {
	ResponseCode    int               `json:"response_code"`
	ResponseMessage string            `json:"response_message"`
	Data            CommentCollection `json:"data"`
}

// NewCommentResource creates a new CommentResource from comment entity
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

		// Computed fields
		IsReply:    comment.IsReply(),
		IsApproved: comment.IsApproved(),
		IsPending:  comment.IsPending(),
		IsRejected: comment.IsRejected(),
		IsDeleted:  comment.IsDeleted(),
	}

	// Set optional fields
	if comment.ParentID != nil {
		parentID := comment.ParentID.String()
		resource.ParentID = &parentID
	}

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

	if comment.DeletedBy != nil {
		deletedBy := comment.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if comment.DeletedAt != nil {
		deletedAt := comment.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	// Add content if provided
	if content != nil {
		resource.Content = &ContentResource{
			ID:          content.ID.String(),
			ContentRaw:  content.ContentRaw,
			ContentHTML: content.ContentHTML,
		}
	}

	// Add author if provided
	if author != nil {
		resource.Author = NewUserResource(author)
	}

	return resource
}

// NewCommentResourceResponse creates a new CommentResourceResponse
func NewCommentResourceResponse(comment *entities.Comment, content *entities.Content, author *entities.User) CommentResourceResponse {
	return CommentResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Comment retrieved successfully",
		Data:            NewCommentResource(comment, content, author),
	}
}

// NewCommentCollection creates a new CommentCollection
func NewCommentCollection(comments []*entities.Comment, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollection {
	commentResources := make([]CommentResource, len(comments))
	for i, comment := range comments {
		var content *entities.Content
		var author *entities.User

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

// NewPaginatedCommentCollection creates a new CommentCollection with pagination
func NewPaginatedCommentCollection(comments []*entities.Comment, page, perPage, total int, baseURL string, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollection {
	collection := NewCommentCollection(comments, contents, authors)

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

// NewCommentCollectionResponse creates a new CommentCollectionResponse
func NewCommentCollectionResponse(comments []*entities.Comment, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollectionResponse {
	return CommentCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Comments retrieved successfully",
		Data:            NewCommentCollection(comments, contents, authors),
	}
}

// NewPaginatedCommentCollectionResponse creates a new CommentCollectionResponse with pagination
func NewPaginatedCommentCollectionResponse(comments []*entities.Comment, page, perPage, total int, baseURL string, contents map[uuid.UUID]*entities.Content, authors map[uuid.UUID]*entities.User) CommentCollectionResponse {
	return CommentCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Comments retrieved successfully",
		Data:            NewPaginatedCommentCollection(comments, page, perPage, total, baseURL, contents, authors),
	}
}

// buildPaginationLink is defined in common_responses.go to avoid duplication
