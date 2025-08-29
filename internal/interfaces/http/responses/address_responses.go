// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// AddressResource represents a single address in API responses.
// This struct follows the Laravel API Resource pattern for consistent formatting
// and provides a standardized way to represent address data in HTTP responses.
type AddressResource struct {
	// ID is the unique identifier for the address
	ID string `json:"id"`
	// AddressableID is the ID of the entity that owns this address (e.g., user, organization)
	AddressableID string `json:"addressable_id"`
	// AddressableType is the type of entity that owns this address (e.g., "user", "organization")
	AddressableType string `json:"addressable_type"`
	// AddressLine1 is the primary address line (e.g., street number and name)
	AddressLine1 string `json:"address_line1"`
	// AddressLine2 is an optional secondary address line (e.g., apartment number)
	AddressLine2 *string `json:"address_line2,omitempty"`
	// City is the city or municipality name
	City string `json:"city"`
	// State is the state or province name
	State string `json:"state"`
	// PostalCode is the postal or ZIP code
	PostalCode string `json:"postal_code"`
	// Country is the country name
	Country string `json:"country"`
	// Latitude is the optional latitude coordinate for geolocation
	Latitude *float64 `json:"latitude,omitempty"`
	// Longitude is the optional longitude coordinate for geolocation
	Longitude *float64 `json:"longitude,omitempty"`
	// IsPrimary indicates whether this is the primary address for the addressable entity
	IsPrimary bool `json:"is_primary"`
	// AddressType specifies the type of address (e.g., "home", "work", "billing")
	AddressType string `json:"address_type"`
	// HasCoordinates indicates whether this address has valid latitude/longitude data
	HasCoordinates bool `json:"has_coordinates"`
	// CreatedAt is the timestamp when the address was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the address was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the address was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// AddressCollection represents a collection of addresses.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type AddressCollection struct {
	// Data contains the array of address resources
	Data []AddressResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// AddressResourceResponse represents a single address response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type AddressResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the address resource
	Data AddressResource `json:"data"`
}

// AddressCollectionResponse represents a collection of addresses response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type AddressCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the address collection
	Data AddressCollection `json:"data"`
}

// NewAddressResource creates a new AddressResource from an address entity.
// This function transforms a domain address entity into an API response resource,
// including computed fields for coordinate validation and address formatting.
//
// Parameters:
//   - address: The domain address entity to convert
//
// Returns:
//   - A new AddressResource with all fields populated from the entity
func NewAddressResource(address *entities.Address) AddressResource {
	resource := AddressResource{
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
		HasCoordinates:  address.HasCoordinates(),
		CreatedAt:       address.CreatedAt,
		UpdatedAt:       address.UpdatedAt,
		DeletedAt:       address.DeletedAt,
	}

	return resource
}

// NewAddressCollection creates a new AddressCollection.
// This function creates a collection from a slice of address entities,
// converting each entity to an AddressResource.
//
// Parameters:
//   - addresses: Slice of domain address entities to convert
//
// Returns:
//   - A new AddressCollection with all addresses converted
func NewAddressCollection(addresses []*entities.Address) AddressCollection {
	addressResources := make([]AddressResource, len(addresses))
	for i, address := range addresses {
		addressResources[i] = NewAddressResource(address)
	}

	return AddressCollection{
		Data: addressResources,
	}
}

// NewAddressResourceResponse creates a new AddressResourceResponse.
// This function wraps an AddressResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - address: The domain address entity to convert and wrap
//
// Returns:
//   - A new AddressResourceResponse with success status and address data
func NewAddressResourceResponse(address *entities.Address) *AddressResourceResponse {
	return &AddressResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Address retrieved successfully",
		Data:            NewAddressResource(address),
	}
}

// NewAddressCollectionResponse creates a new AddressCollectionResponse.
// This function wraps an AddressCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - addresses: Slice of domain address entities to convert and wrap
//
// Returns:
//   - A new AddressCollectionResponse with success status and address collection data
func NewAddressCollectionResponse(addresses []*entities.Address) *AddressCollectionResponse {
	return &AddressCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Addresses retrieved successfully",
		Data:            NewAddressCollection(addresses),
	}
}

// NewPaginatedAddressCollectionResponse creates a new AddressCollectionResponse with pagination.
// This function wraps a paginated AddressCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - addresses: Slice of address entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated AddressCollectionResponse with success status and pagination data
func NewPaginatedAddressCollectionResponse(addresses []*entities.Address, page, perPage, total int, baseURL string) *AddressCollectionResponse {
	collection := NewAddressCollection(addresses)

	// Use the pagination utility to create metadata and links
	meta, links := CreatePaginatedCollection(int(page), int(perPage), int(total), baseURL)
	collection.Meta = meta
	collection.Links = links

	return &AddressCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Addresses retrieved successfully",
		Data:            collection,
	}
}
