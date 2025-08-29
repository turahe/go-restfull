// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// MediaResource represents a media item in API responses.
// This struct provides a comprehensive view of media files including file metadata,
// hierarchical organization using nested set model fields, and computed properties
// for file type detection and size formatting.
type MediaResource struct {
	// ID is the unique identifier for the media item
	ID string `json:"id"`
	// Name is the display name of the media item
	Name string `json:"name"`
	// FileName is the original filename of the uploaded file
	FileName string `json:"file_name"`
	// Hash is the unique hash value for file integrity verification
	Hash string `json:"hash"`
	// Disk specifies the storage disk where the file is located
	Disk string `json:"disk"`
	// MimeType indicates the MIME type of the file (e.g., "image/jpeg", "video/mp4")
	MimeType string `json:"mime_type"`
	// Size is the file size in bytes
	Size int64 `json:"size"`
	// RecordLeft is the left boundary value for nested set model hierarchy
	RecordLeft *uint64 `json:"record_left,omitempty"`
	// RecordRight is the right boundary value for nested set model hierarchy
	RecordRight *uint64 `json:"record_right,omitempty"`
	// RecordOrdering is the ordering value within the hierarchy level
	RecordOrdering *uint64 `json:"record_ordering,omitempty"`
	// RecordDepth is the depth level in the hierarchy tree
	RecordDepth *uint64 `json:"record_depth,omitempty"`
	// CreatedBy is the ID of the user who created the media item
	CreatedBy string `json:"created_by"`
	// UpdatedBy is the ID of the user who last updated the media item
	UpdatedBy string `json:"updated_by"`
	// DeletedBy is the optional ID of the user who deleted the media item
	DeletedBy *string `json:"deleted_by,omitempty"`
	// CreatedAt is the timestamp when the media item was created
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the media item was last updated
	UpdatedAt string `json:"updated_at"`
	// DeletedAt is the optional timestamp when the media item was soft-deleted
	DeletedAt *string `json:"deleted_at,omitempty"`

	// Computed fields
	// IsDeleted indicates whether the media item has been soft-deleted
	IsDeleted bool `json:"is_deleted"`
	// IsImage indicates whether the file is an image (based on MIME type)
	IsImage bool `json:"is_image"`
	// IsVideo indicates whether the file is a video (based on MIME type)
	IsVideo bool `json:"is_video"`
	// IsAudio indicates whether the file is an audio file (based on MIME type)
	IsAudio bool `json:"is_audio"`
	// FileExtension is the file extension extracted from the filename
	FileExtension string `json:"file_extension"`
	// URL is the public URL where the media file can be accessed
	URL string `json:"url"`
	// FileSizeInMB is the file size formatted in megabytes
	FileSizeInMB float64 `json:"file_size_in_mb"`
	// FileSizeInKB is the file size formatted in kilobytes
	FileSizeInKB float64 `json:"file_size_in_kb"`
}

// MediaCollection represents a collection of media items.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type MediaCollection struct {
	// Data contains the array of media item resources
	Data []MediaResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// MediaResourceResponse represents a single media item response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type MediaResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the media item resource
	Data MediaResource `json:"data"`
}

// MediaCollectionResponse represents a collection of media items response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type MediaCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the media collection
	Data MediaCollection `json:"data"`
}

// NewMediaResource creates a new MediaResource from a media entity.
// This function transforms a domain media entity into an API response resource,
// including computed fields for file type detection and size formatting.
//
// Parameters:
//   - media: The domain media entity to convert
//
// Returns:
//   - A new MediaResource with all fields populated from the entity
func NewMediaResource(media *entities.Media) MediaResource {
	resource := MediaResource{
		ID:        media.ID.String(),
		Name:      media.Name,
		FileName:  media.FileName,
		Hash:      media.Hash,
		Disk:      media.Disk,
		MimeType:  media.MimeType,
		Size:      media.Size,
		CreatedBy: media.CreatedBy.String(),
		UpdatedBy: media.UpdatedBy.String(),
		CreatedAt: media.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: media.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Computed fields
		IsDeleted:     media.IsDeleted(),
		IsImage:       media.IsImage(),
		IsVideo:       media.IsVideo(),
		IsAudio:       media.IsAudio(),
		FileExtension: media.GetFileExtension(),
		URL:           media.GetURL(),
		FileSizeInMB:  media.GetFileSizeInMB(),
		FileSizeInKB:  media.GetFileSizeInKB(),
	}

	// Set optional nested set model fields if they exist
	if media.RecordLeft != nil {
		resource.RecordLeft = media.RecordLeft
	}

	if media.RecordRight != nil {
		resource.RecordRight = media.RecordRight
	}

	if media.RecordOrdering != nil {
		resource.RecordOrdering = media.RecordOrdering
	}

	if media.RecordDepth != nil {
		resource.RecordDepth = media.RecordDepth
	}

	// Set optional deletion tracking fields if they exist
	if media.DeletedBy != nil {
		deletedBy := media.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if media.DeletedAt != nil {
		deletedAt := media.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	return resource
}

// NewMediaResourceResponse creates a new MediaResourceResponse.
// This function wraps a MediaResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - media: The domain media entity to convert and wrap
//
// Returns:
//   - A new MediaResourceResponse with success status and media data
func NewMediaResourceResponse(media *entities.Media) MediaResourceResponse {
	return MediaResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Media retrieved successfully",
		Data:            NewMediaResource(media),
	}
}

// NewMediaCollection creates a new MediaCollection.
// This function creates a collection from a slice of media entities,
// converting each entity to a MediaResource.
//
// Parameters:
//   - media: Slice of domain media entities to convert
//
// Returns:
//   - A new MediaCollection with all media resources converted
func NewMediaCollection(media []*entities.Media) MediaCollection {
	mediaResources := make([]MediaResource, len(media))
	for i, m := range media {
		mediaResources[i] = NewMediaResource(m)
	}

	return MediaCollection{
		Data: mediaResources,
	}
}

// NewPaginatedMediaCollection creates a new MediaCollection with pagination.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - media: Slice of media entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated MediaCollection with metadata and navigation links
func NewPaginatedMediaCollection(media []*entities.Media, page, perPage, total int, baseURL string) MediaCollection {
	collection := NewMediaCollection(media)

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

// NewMediaCollectionResponse creates a new MediaCollectionResponse.
// This function wraps a MediaCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - media: Slice of domain media entities to convert and wrap
//
// Returns:
//   - A new MediaCollectionResponse with success status and media collection data
func NewMediaCollectionResponse(media []*entities.Media) MediaCollectionResponse {
	return MediaCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Media retrieved successfully",
		Data:            NewMediaCollection(media),
	}
}

// NewPaginatedMediaCollectionResponse creates a new MediaCollectionResponse with pagination.
// This function wraps a paginated MediaCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - media: Slice of media entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated MediaCollectionResponse with success status and pagination data
func NewPaginatedMediaCollectionResponse(media []*entities.Media, page, perPage, total int, baseURL string) MediaCollectionResponse {
	return MediaCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Media retrieved successfully",
		Data:            NewPaginatedMediaCollection(media, page, perPage, total, baseURL),
	}
}
