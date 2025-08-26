package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// AddressResource represents a single address in API responses
// Following Laravel API Resource pattern for consistent formatting
type AddressResource struct {
	ID              string     `json:"id"`
	AddressableID   string     `json:"addressable_id"`
	AddressableType string     `json:"addressable_type"`
	AddressLine1    string     `json:"address_line1"`
	AddressLine2    *string    `json:"address_line2,omitempty"`
	City            string     `json:"city"`
	State           string     `json:"state"`
	PostalCode      string     `json:"postal_code"`
	Country         string     `json:"country"`
	Latitude        *float64   `json:"latitude,omitempty"`
	Longitude       *float64   `json:"longitude,omitempty"`
	IsPrimary       bool       `json:"is_primary"`
	AddressType     string     `json:"address_type"`
	FullAddress     string     `json:"full_address"`
	HasCoordinates  bool       `json:"has_coordinates"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// AddressCollection represents a collection of addresses
// Following Laravel API Resource Collection pattern
type AddressCollection struct {
	Data  []AddressResource `json:"data"`
	Meta  *CollectionMeta   `json:"meta,omitempty"`
	Links *CollectionLinks  `json:"links,omitempty"`
}

// CollectionMeta and CollectionLinks are defined in common_responses.go

// NewAddressResource creates a new AddressResource from an Address entity
// This transforms the domain entity into a consistent API response format
func NewAddressResource(address *entities.Address) *AddressResource {
	return &AddressResource{
		ID:              address.ID.String(),
		AddressableID:   address.AddressableID.String(),
		AddressableType: string(address.AddressableType),
		AddressLine1:    address.AddressLine1,
		AddressLine2:    address.AddressLine2,
		City:            address.City,
		State:           address.State,
		PostalCode:      address.PostalCode,
		Country:         address.Country,
		Latitude:        address.Latitude,
		Longitude:       address.Longitude,
		IsPrimary:       address.IsPrimary,
		AddressType:     string(address.AddressType),
		FullAddress:     address.GetFullAddress(),
		HasCoordinates:  address.HasCoordinates(),
		CreatedAt:       address.CreatedAt,
		UpdatedAt:       address.UpdatedAt,
		DeletedAt:       address.DeletedAt,
	}
}

// NewAddressCollection creates a new AddressCollection from a slice of Address entities
// This transforms multiple domain entities into a consistent API response format
func NewAddressCollection(addresses []*entities.Address) *AddressCollection {
	addressResources := make([]AddressResource, len(addresses))
	for i, address := range addresses {
		addressResources[i] = *NewAddressResource(address)
	}

	return &AddressCollection{
		Data: addressResources,
	}
}

// NewPaginatedAddressCollection creates a new AddressCollection with pagination metadata
// This follows Laravel's paginated resource collection pattern
func NewPaginatedAddressCollection(
	addresses []*entities.Address,
	page, perPage int,
	total int64,
	baseURL string,
) *AddressCollection {
	collection := NewAddressCollection(addresses)

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

// generatePageURL is defined in common_responses.go to avoid duplication

// AddressResourceResponse represents a single address response with Laravel-style formatting
type AddressResourceResponse struct {
	Status string          `json:"status"`
	Data   AddressResource `json:"data"`
}

// AddressCollectionResponse represents a collection response with Laravel-style formatting
type AddressCollectionResponse struct {
	Status string            `json:"status"`
	Data   AddressCollection `json:"data"`
}

// NewAddressResourceResponse creates a new single address response
func NewAddressResourceResponse(address *entities.Address) *AddressResourceResponse {
	return &AddressResourceResponse{
		Status: "success",
		Data:   *NewAddressResource(address),
	}
}

// NewAddressCollectionResponse creates a new address collection response
func NewAddressCollectionResponse(addresses []*entities.Address) *AddressCollectionResponse {
	return &AddressCollectionResponse{
		Status: "success",
		Data:   *NewAddressCollection(addresses),
	}
}

// NewPaginatedAddressCollectionResponse creates a new paginated address collection response
func NewPaginatedAddressCollectionResponse(
	addresses []*entities.Address,
	page, perPage int,
	total int64,
	baseURL string,
) *AddressCollectionResponse {
	return &AddressCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedAddressCollection(addresses, page, perPage, total, baseURL),
	}
}
