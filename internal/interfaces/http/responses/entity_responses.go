// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// TagResponse represents a tag in API responses.
// This struct provides a standardized way to represent tag data in HTTP responses,
// including basic tag information and timestamps.
type TagResponse struct {
	// ID is the unique identifier for the tag
	ID string `json:"id"`
	// Name is the display name of the tag
	Name string `json:"name"`
	// Slug is the URL-friendly version of the tag name
	Slug string `json:"slug"`
	// Description is an optional description of the tag
	Description string `json:"description,omitempty"`
	// CreatedAt is the timestamp when the tag was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the tag was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// TagListResponse represents a list of tags with pagination.
// This struct provides a paginated collection of tags following the legacy
// pagination pattern (will be deprecated in favor of the new collection pattern).
type TagListResponse struct {
	// Tags contains the array of tag responses
	Tags []TagResponse `json:"tags"`
	// Total indicates the total number of tags across all pages
	Total int64 `json:"total"`
	// Limit specifies the maximum number of tags per page
	Limit int `json:"limit"`
	// Page indicates the current page number
	Page int `json:"page"`
}

// NewTagResponse creates a new TagResponse from tag entity.
// This function transforms the domain entity into a consistent API response format.
//
// Parameters:
//   - tag: The tag domain entity to convert
//
// Returns:
//   - A pointer to the newly created TagResponse
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

// NewTagListResponse creates a new TagListResponse from tag entities.
// This function transforms multiple tag domain entities into a paginated response format.
//
// Parameters:
//   - tags: Slice of tag domain entities to convert
//   - total: Total number of tags across all pages
//   - limit: Maximum number of tags per page
//   - page: Current page number
//
// Returns:
//   - A pointer to the newly created TagListResponse
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

// CommentResponse represents a comment in API responses.
// This struct provides a simplified view of comment data for basic API responses,
// excluding complex nested content and author information.
type CommentResponse struct {
	// ID is the unique identifier for the comment
	ID string `json:"id"`
	// Content is the comment's text content
	Content string `json:"content"`
	// PostID is the ID of the post this comment belongs to
	PostID string `json:"post_id"`
	// UserID is the ID of the user who wrote the comment
	UserID string `json:"user_id"`
	// ParentID is the optional ID of the parent comment for nested replies
	ParentID *string `json:"parent_id,omitempty"`
	// Status indicates the current status of the comment
	Status string `json:"status"`
	// CreatedAt is the timestamp when the comment was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the comment was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the comment was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CommentListResponse represents a list of comments with pagination.
// This struct provides a paginated collection of comments following the legacy
// pagination pattern (will be deprecated in favor of the new collection pattern).
type CommentListResponse struct {
	// Comments contains the array of comment responses
	Comments []CommentResponse `json:"comments"`
	// Total indicates the total number of comments across all pages
	Total int64 `json:"total"`
	// Limit specifies the maximum number of comments per page
	Limit int `json:"limit"`
	// Page indicates the current page number
	Page int `json:"page"`
}

// NewCommentResponse creates a new CommentResponse from comment entity.
// This function transforms the domain entity into a simplified API response format.
//
// Parameters:
//   - comment: The comment domain entity to convert
//
// Returns:
//   - A pointer to the newly created CommentResponse
func NewCommentResponse(comment *entities.Comment) *CommentResponse {
	response := &CommentResponse{
		ID:        comment.ID.String(),
		Status:    string(comment.Status),
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		DeletedAt: comment.DeletedAt,
	}

	// Handle optional parent ID for nested comments
	if comment.ParentID != nil {
		parentID := comment.ParentID.String()
		response.ParentID = &parentID
	}

	return response
}

// MediaResponse represents a media file in API responses.
// This struct provides information about uploaded media files including
// file metadata, storage details, and timestamps.
type MediaResponse struct {
	// ID is the unique identifier for the media file
	ID string `json:"id"`
	// FileName is the stored filename of the media file
	FileName string `json:"file_name"`
	// OriginalName is the original filename as uploaded by the user
	OriginalName string `json:"original_name"`
	// MimeType is the MIME type of the media file
	MimeType string `json:"mime_type"`
	// Size is the file size in bytes
	Size int64 `json:"size"`
	// Path is the storage disk/path where the file is stored
	Path string `json:"path"`
	// URL is the public URL to access the media file
	URL string `json:"url"`
	// UserID is the ID of the user who uploaded the media file
	UserID string `json:"user_id"`
	// CreatedAt is the timestamp when the media file was uploaded
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the media file was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the media file was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// MediaListResponse represents a list of media files with pagination.
// This struct provides a paginated collection of media files following the legacy
// pagination pattern (will be deprecated in favor of the new collection pattern).
type MediaListResponse struct {
	// Media contains the array of media responses
	Media []MediaResponse `json:"media"`
	// Total indicates the total number of media files across all pages
	Total int64 `json:"total"`
	// Limit specifies the maximum number of media files per page
	Limit int `json:"limit"`
	// Page indicates the current page number
	Page int `json:"page"`
}

// NewMediaResponse creates a new MediaResponse from media entity.
// This function transforms the domain entity into a consistent API response format.
//
// Parameters:
//   - media: The media domain entity to convert
//
// Returns:
//   - A pointer to the newly created MediaResponse
func NewMediaResponse(media *entities.Media) *MediaResponse {
	return &MediaResponse{
		ID:           media.ID.String(),
		FileName:     media.FileName,
		OriginalName: media.Name,
		MimeType:     media.MimeType,
		Size:         media.Size,
		Path:         media.Disk,
		CreatedAt:    media.CreatedAt,
		UpdatedAt:    media.UpdatedAt,
		DeletedAt:    media.DeletedAt,
	}
}

// TaxonomyResponse represents a taxonomy in API responses.
// This struct provides a hierarchical view of taxonomy data including
// nested set model information for tree operations and optional children.
type TaxonomyResponse struct {
	// ID is the unique identifier for the taxonomy
	ID string `json:"id"`
	// Name is the display name of the taxonomy
	Name string `json:"name"`
	// Slug is the URL-friendly version of the taxonomy name
	Slug string `json:"slug"`
	// Code is an optional code identifier for the taxonomy
	Code string `json:"code,omitempty"`
	// Description is an optional description of the taxonomy
	Description string `json:"description,omitempty"`
	// ParentID is the optional ID of the parent taxonomy
	ParentID *string `json:"parent_id,omitempty"`
	// RecordLeft is used for nested set model operations (tree structure)
	RecordLeft *int64 `json:"record_left"`
	// RecordRight is used for nested set model operations (tree structure)
	RecordRight *int64 `json:"record_right"`
	// RecordDepth indicates the nesting level of the taxonomy in the tree
	RecordDepth *int64 `json:"record_depth"`
	// CreatedAt is the timestamp when the taxonomy was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the taxonomy was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the taxonomy was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// Children contains the nested child taxonomies
	Children []TaxonomyResponse `json:"children,omitempty"`
}

// TaxonomyListResponse represents a list of taxonomies with pagination.
// This struct provides a paginated collection of taxonomies following the legacy
// pagination pattern (will be deprecated in favor of the new collection pattern).
type TaxonomyListResponse struct {
	// Taxonomies contains the array of taxonomy responses
	Taxonomies []TaxonomyResponse `json:"taxonomies"`
	// Total indicates the total number of taxonomies across all pages
	Total int64 `json:"total"`
	// Limit specifies the maximum number of taxonomies per page
	Limit int `json:"limit"`
	// Page indicates the current page number
	Page int `json:"page"`
}

// NewTaxonomyResponse creates a new TaxonomyResponse from taxonomy entity.
// This function transforms the domain entity into a consistent API response format,
// handling optional fields and nested set model data.
//
// Parameters:
//   - taxonomy: The taxonomy domain entity to convert
//
// Returns:
//   - A pointer to the newly created TaxonomyResponse
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

	// Handle optional parent ID for hierarchical taxonomies
	if taxonomy.ParentID != nil {
		parentID := taxonomy.ParentID.String()
		response.ParentID = &parentID
	}

	return response
}

// NewTaxonomyListResponse creates a new TaxonomyListResponse from taxonomy entities.
// This function transforms multiple taxonomy domain entities into a paginated response format.
//
// Parameters:
//   - taxonomies: Slice of taxonomy domain entities to convert
//   - total: Total number of taxonomies across all pages
//   - limit: Maximum number of taxonomies per page
//   - page: Current page number
//
// Returns:
//   - A pointer to the newly created TaxonomyListResponse
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

// BuildTaxonomyTree builds a hierarchical taxonomy tree from flat taxonomy list.
// This function transforms a flat list of taxonomies into a hierarchical tree structure
// using the nested set model data (RecordLeft, RecordRight, RecordDepth).
//
// Parameters:
//   - taxonomies: Slice of taxonomy domain entities to organize into a tree
//
// Returns:
//   - A slice of root TaxonomyResponse objects with nested children
func BuildTaxonomyTree(taxonomies []*entities.Taxonomy) []TaxonomyResponse {
	taxonomyMap := make(map[string]*TaxonomyResponse)
	var rootTaxonomies []TaxonomyResponse

	// First pass: create all taxonomy responses and store them in a map
	for _, taxonomy := range taxonomies {
		taxonomyResponse := NewTaxonomyResponse(taxonomy)
		taxonomyMap[taxonomy.ID.String()] = taxonomyResponse
	}

	// Second pass: build the hierarchy by connecting parents and children
	for _, taxonomy := range taxonomies {
		taxonomyResponse := taxonomyMap[taxonomy.ID.String()]

		if taxonomy.ParentID == nil {
			// This is a root taxonomy (no parent)
			rootTaxonomies = append(rootTaxonomies, *taxonomyResponse)
		} else {
			// This is a child taxonomy - add it to its parent's children
			if parent, exists := taxonomyMap[taxonomy.ParentID.String()]; exists {
				parent.Children = append(parent.Children, *taxonomyResponse)
			}
		}
	}

	return rootTaxonomies
}
