package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// MediaResource represents a media item in API responses
type MediaResource struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	FileName       string  `json:"file_name"`
	Hash           string  `json:"hash"`
	Disk           string  `json:"disk"`
	MimeType       string  `json:"mime_type"`
	Size           int64   `json:"size"`
	RecordLeft     *uint64 `json:"record_left,omitempty"`
	RecordRight    *uint64 `json:"record_right,omitempty"`
	RecordOrdering *uint64 `json:"record_ordering,omitempty"`
	RecordDepth    *uint64 `json:"record_depth,omitempty"`
	CreatedBy      string  `json:"created_by"`
	UpdatedBy      string  `json:"updated_by"`
	DeletedBy      *string `json:"deleted_by,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	DeletedAt      *string `json:"deleted_at,omitempty"`

	// Computed fields
	IsDeleted     bool    `json:"is_deleted"`
	IsImage       bool    `json:"is_image"`
	IsVideo       bool    `json:"is_video"`
	IsAudio       bool    `json:"is_audio"`
	FileExtension string  `json:"file_extension"`
	URL           string  `json:"url"`
	FileSizeInMB  float64 `json:"file_size_in_mb"`
	FileSizeInKB  float64 `json:"file_size_in_kb"`
}

// MediaCollection represents a collection of media items
type MediaCollection struct {
	Data  []MediaResource `json:"data"`
	Meta  CollectionMeta  `json:"meta"`
	Links CollectionLinks `json:"links"`
}

// MediaResourceResponse represents a single media item response
type MediaResourceResponse struct {
	ResponseCode    int           `json:"response_code"`
	ResponseMessage string        `json:"response_message"`
	Data            MediaResource `json:"data"`
}

// MediaCollectionResponse represents a collection of media items response
type MediaCollectionResponse struct {
	ResponseCode    int             `json:"response_code"`
	ResponseMessage string          `json:"response_message"`
	Data            MediaCollection `json:"data"`
}

// NewMediaResource creates a new MediaResource from media entity
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

	// Set optional fields
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

// NewMediaResourceResponse creates a new MediaResourceResponse
func NewMediaResourceResponse(media *entities.Media) MediaResourceResponse {
	return MediaResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Media retrieved successfully",
		Data:            NewMediaResource(media),
	}
}

// NewMediaCollection creates a new MediaCollection
func NewMediaCollection(media []*entities.Media) MediaCollection {
	mediaResources := make([]MediaResource, len(media))
	for i, m := range media {
		mediaResources[i] = NewMediaResource(m)
	}

	return MediaCollection{
		Data: mediaResources,
	}
}

// NewPaginatedMediaCollection creates a new MediaCollection with pagination
func NewPaginatedMediaCollection(media []*entities.Media, page, perPage, total int, baseURL string) MediaCollection {
	collection := NewMediaCollection(media)

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

// NewMediaCollectionResponse creates a new MediaCollectionResponse
func NewMediaCollectionResponse(media []*entities.Media) MediaCollectionResponse {
	return MediaCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Media retrieved successfully",
		Data:            NewMediaCollection(media),
	}
}

// NewPaginatedMediaCollectionResponse creates a new MediaCollectionResponse with pagination
func NewPaginatedMediaCollectionResponse(media []*entities.Media, page, perPage, total int, baseURL string) MediaCollectionResponse {
	return MediaCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Media retrieved successfully",
		Data:            NewPaginatedMediaCollection(media, page, perPage, total, baseURL),
	}
}
