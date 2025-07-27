package responses

import (
	"time"

	"webapi/internal/domain/entities"
)

// TagResponse represents a tag in API responses
type TagResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TagListResponse represents a list of tags with pagination
type TagListResponse struct {
	Tags  []TagResponse `json:"tags"`
	Total int64         `json:"total"`
	Limit int           `json:"limit"`
	Page  int           `json:"page"`
}

// NewTagResponse creates a new TagResponse from tag entity
func NewTagResponse(tag *entities.Tag) *TagResponse {
	return &TagResponse{
		ID:          tag.ID.String(),
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

// NewTagListResponse creates a new TagListResponse from tag entities
func NewTagListResponse(tags []*entities.Tag, total int64, limit, page int) *TagListResponse {
	tagResponses := make([]TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = *NewTagResponse(tag)
	}

	return &TagListResponse{
		Tags:  tagResponses,
		Total: total,
		Limit: limit,
		Page:  page,
	}
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	PostID    string     `json:"post_id"`
	UserID    string     `json:"user_id"`
	ParentID  *string    `json:"parent_id,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CommentListResponse represents a list of comments with pagination
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Page     int               `json:"page"`
}

// NewCommentResponse creates a new CommentResponse from comment entity
func NewCommentResponse(comment *entities.Comment) *CommentResponse {
	response := &CommentResponse{
		ID:        comment.ID.String(),
		Content:   comment.Content,
		PostID:    comment.PostID.String(),
		UserID:    comment.UserID.String(),
		Status:    comment.Status,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		DeletedAt: comment.DeletedAt,
	}

	if comment.ParentID != nil {
		parentID := comment.ParentID.String()
		response.ParentID = &parentID
	}

	return response
}

// NewCommentListResponse creates a new CommentListResponse from comment entities
func NewCommentListResponse(comments []*entities.Comment, total int64, limit, page int) *CommentListResponse {
	commentResponses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		commentResponses[i] = *NewCommentResponse(comment)
	}

	return &CommentListResponse{
		Comments: commentResponses,
		Total:    total,
		Limit:    limit,
		Page:     page,
	}
}

// MediaResponse represents a media in API responses
type MediaResponse struct {
	ID           string     `json:"id"`
	FileName     string     `json:"file_name"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	Size         int64      `json:"size"`
	Path         string     `json:"path"`
	URL          string     `json:"url"`
	UserID       string     `json:"user_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// MediaListResponse represents a list of media with pagination
type MediaListResponse struct {
	Media []MediaResponse `json:"media"`
	Total int64           `json:"total"`
	Limit int             `json:"limit"`
	Page  int             `json:"page"`
}

// NewMediaResponse creates a new MediaResponse from media entity
func NewMediaResponse(media *entities.Media) *MediaResponse {
	return &MediaResponse{
		ID:           media.ID.String(),
		FileName:     media.FileName,
		OriginalName: media.OriginalName,
		MimeType:     media.MimeType,
		Size:         media.Size,
		Path:         media.Path,
		URL:          media.URL,
		UserID:       media.UserID.String(),
		CreatedAt:    media.CreatedAt,
		UpdatedAt:    media.UpdatedAt,
		DeletedAt:    media.DeletedAt,
	}
}

// NewMediaListResponse creates a new MediaListResponse from media entities
func NewMediaListResponse(media []*entities.Media, total int64, limit, page int) *MediaListResponse {
	mediaResponses := make([]MediaResponse, len(media))
	for i, m := range media {
		mediaResponses[i] = *NewMediaResponse(m)
	}

	return &MediaListResponse{
		Media: mediaResponses,
		Total: total,
		Limit: limit,
		Page:  page,
	}
}

// TaxonomyResponse represents a taxonomy in API responses
type TaxonomyResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Code        string             `json:"code,omitempty"`
	Description string             `json:"description,omitempty"`
	ParentID    *string            `json:"parent_id,omitempty"`
	RecordLeft  int64              `json:"record_left"`
	RecordRight int64              `json:"record_right"`
	RecordDepth int64              `json:"record_depth"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty"`
	Children    []TaxonomyResponse `json:"children,omitempty"`
}

// TaxonomyListResponse represents a list of taxonomies with pagination
type TaxonomyListResponse struct {
	Taxonomies []TaxonomyResponse `json:"taxonomies"`
	Total      int64              `json:"total"`
	Limit      int                `json:"limit"`
	Page       int                `json:"page"`
}

// NewTaxonomyResponse creates a new TaxonomyResponse from taxonomy entity
func NewTaxonomyResponse(taxonomy *entities.Taxonomy) *TaxonomyResponse {
	response := &TaxonomyResponse{
		ID:          taxonomy.ID.String(),
		Name:        taxonomy.Name,
		Slug:        taxonomy.Slug,
		Code:        taxonomy.Code,
		Description: taxonomy.Description,
		RecordLeft:  taxonomy.RecordLeft,
		RecordRight: taxonomy.RecordRight,
		RecordDepth: taxonomy.RecordDepth,
		CreatedAt:   taxonomy.CreatedAt,
		UpdatedAt:   taxonomy.UpdatedAt,
		DeletedAt:   taxonomy.DeletedAt,
	}

	if taxonomy.ParentID != nil {
		parentID := taxonomy.ParentID.String()
		response.ParentID = &parentID
	}

	return response
}

// NewTaxonomyListResponse creates a new TaxonomyListResponse from taxonomy entities
func NewTaxonomyListResponse(taxonomies []*entities.Taxonomy, total int64, limit, page int) *TaxonomyListResponse {
	taxonomyResponses := make([]TaxonomyResponse, len(taxonomies))
	for i, taxonomy := range taxonomies {
		taxonomyResponses[i] = *NewTaxonomyResponse(taxonomy)
	}

	return &TaxonomyListResponse{
		Taxonomies: taxonomyResponses,
		Total:      total,
		Limit:      limit,
		Page:       page,
	}
}

// BuildTaxonomyTree builds a hierarchical taxonomy tree from flat taxonomy list
func BuildTaxonomyTree(taxonomies []*entities.Taxonomy) []TaxonomyResponse {
	taxonomyMap := make(map[string]*TaxonomyResponse)
	var rootTaxonomies []TaxonomyResponse

	// First pass: create all taxonomy responses
	for _, taxonomy := range taxonomies {
		taxonomyResponse := NewTaxonomyResponse(taxonomy)
		taxonomyMap[taxonomy.ID.String()] = taxonomyResponse
	}

	// Second pass: build hierarchy
	for _, taxonomy := range taxonomies {
		taxonomyResponse := taxonomyMap[taxonomy.ID.String()]

		if taxonomy.ParentID == nil {
			// Root taxonomy
			rootTaxonomies = append(rootTaxonomies, *taxonomyResponse)
		} else {
			// Child taxonomy
			if parent, exists := taxonomyMap[taxonomy.ParentID.String()]; exists {
				parent.Children = append(parent.Children, *taxonomyResponse)
			}
		}
	}

	return rootTaxonomies
}
