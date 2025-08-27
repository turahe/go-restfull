// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateAddressRequest represents the request for creating a new address entity.
// This struct defines the required and optional fields for address creation,
// including validation tags for field constraints and business rules.
// The request supports polymorphic relationships with users and organizations.
type CreateAddressRequest struct {
	// AddressableID is the UUID of the entity (user or organization) that owns this address
	AddressableID string `json:"addressable_id" validate:"required,uuid4"`
	// AddressableType specifies whether this address belongs to a user or organization
	AddressableType string `json:"addressable_type" validate:"required,oneof=user organization"`
	// AddressLine1 is the primary street address (required, 1-255 characters)
	AddressLine1 string `json:"address_line1" validate:"required,min=1,max=255"`
	// AddressLine2 is an optional secondary address line (max 255 characters)
	AddressLine2 *string `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	// City is the city or municipality name (required, 1-255 characters)
	City string `json:"city" validate:"required,min=1,max=255"`
	// State is the state, province, or region (required, 1-255 characters)
	State string `json:"state" validate:"required,min=1,max=255"`
	// PostalCode is the postal or ZIP code (required, 1-20 characters)
	PostalCode string `json:"postal_code" validate:"required,min=1,max=20"`
	// Country is the country name (required, 1-255 characters)
	Country string `json:"country" validate:"required,min=1,max=255"`
	// Latitude is the optional geographic latitude coordinate (-90 to 90 degrees)
	Latitude *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	// Longitude is the optional geographic longitude coordinate (-180 to 180 degrees)
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	// IsPrimary indicates whether this is the primary address for the addressable entity
	IsPrimary bool `json:"is_primary"`
	// AddressType specifies the purpose of this address (home, work, billing, shipping, other)
	AddressType string `json:"address_type" validate:"required,oneof=home work billing shipping other"`
}

// Validate performs validation on the CreateAddressRequest using the validator package.
// This method checks all field constraints including required fields, length limits,
// and enumerated value restrictions.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateAddressRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the CreateAddressRequest to an Address domain entity.
// This method parses the addressable_id string to UUID and converts the request
// fields to the appropriate domain entity structure for persistence.
//
// Returns:
//   - *entities.Address: The created address entity
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *CreateAddressRequest) ToEntity() (*entities.Address, error) {
	// Parse the addressable_id string to UUID
	addressableID, err := uuid.Parse(r.AddressableID)
	if err != nil {
		return nil, err
	}

	// Create and populate the address entity
	address := &entities.Address{
		AddressableID:   addressableID,
		AddressableType: entities.AddressableType(r.AddressableType),
		AddressLine1:    r.AddressLine1,
		AddressLine2:    r.AddressLine2,
		City:            r.City,
		State:           r.State,
		PostalCode:      r.PostalCode,
		Country:         r.Country,
		Latitude:        r.Latitude,
		Longitude:       r.Longitude,
		IsPrimary:       r.IsPrimary,
		AddressType:     entities.AddressType(r.AddressType),
	}

	return address, nil
}

// UpdateAddressRequest represents the request for updating an existing address entity.
// This struct uses omitempty tags to make all fields optional, allowing partial updates.
// Only provided fields will be updated in the existing address entity.
type UpdateAddressRequest struct {
	// AddressLine1 is the primary street address (optional, 1-255 characters if provided)
	AddressLine1 string `json:"address_line1,omitempty" validate:"omitempty,min=1,max=255"`
	// AddressLine2 is an optional secondary address line (max 255 characters if provided)
	AddressLine2 *string `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	// City is the city or municipality name (optional, 1-255 characters if provided)
	City string `json:"city,omitempty" validate:"omitempty,min=1,max=255"`
	// State is the state, province, or region (optional, 1-255 characters if provided)
	State string `json:"state,omitempty" validate:"omitempty,min=1,max=255"`
	// PostalCode is the postal or ZIP code (optional, 1-20 characters if provided)
	PostalCode string `json:"postal_code,omitempty" validate:"omitempty,min=1,max=20"`
	// Country is the country name (optional, 1-255 characters if provided)
	Country string `json:"country,omitempty" validate:"omitempty,min=1,max=255"`
	// Latitude is the optional geographic latitude coordinate (-90 to 90 degrees if provided)
	Latitude *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	// Longitude is the optional geographic longitude coordinate (-180 to 180 degrees if provided)
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	// IsPrimary indicates whether this is the primary address (optional boolean)
	IsPrimary *bool `json:"is_primary,omitempty"`
	// AddressType specifies the purpose of this address (optional, must be valid enum if provided)
	AddressType string `json:"address_type,omitempty" validate:"omitempty,oneof=home work billing shipping other"`
}

// Validate performs validation on the UpdateAddressRequest using the validator package.
// This method checks field constraints for any provided fields while allowing
// all fields to be optional for partial updates.
//
// Returns:
//   - error: Validation error if any provided field fails validation, nil if valid
func (r *UpdateAddressRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the UpdateAddressRequest to update an existing Address entity.
// This method applies only the provided fields to the existing address, preserving
// unchanged values. It's designed for partial updates where not all fields are provided.
//
// Parameters:
//   - existingAddress: The existing address entity to update
//
// Returns:
//   - *entities.Address: The updated address entity
//   - error: Any error that occurred during transformation
func (r *UpdateAddressRequest) ToEntity(existingAddress *entities.Address) (*entities.Address, error) {
	// Update fields only if provided in the request
	if r.AddressLine1 != "" {
		existingAddress.AddressLine1 = r.AddressLine1
	}
	if r.AddressLine2 != nil {
		existingAddress.AddressLine2 = r.AddressLine2
	}
	if r.City != "" {
		existingAddress.City = r.City
	}
	if r.State != "" {
		existingAddress.State = r.State
	}
	if r.PostalCode != "" {
		existingAddress.PostalCode = r.PostalCode
	}
	if r.Country != "" {
		existingAddress.Country = r.Country
	}
	if r.Latitude != nil {
		existingAddress.Latitude = r.Latitude
	}
	if r.Longitude != nil {
		existingAddress.Longitude = r.Longitude
	}
	if r.IsPrimary != nil {
		existingAddress.IsPrimary = *r.IsPrimary
	}
	if r.AddressType != "" {
		existingAddress.AddressType = entities.AddressType(r.AddressType)
	}

	return existingAddress, nil
}

// SetPrimaryAddressRequest represents the request for setting an address as the primary address.
// This request is used to designate a specific address as the main address for a user
// or organization, which may affect billing, shipping, or display preferences.
type SetPrimaryAddressRequest struct {
	// AddressableID is the UUID of the entity (user or organization) that owns the address
	AddressableID string `json:"addressable_id" validate:"required,uuid4"`
	// AddressableType specifies whether this address belongs to a user or organization
	AddressableType string `json:"addressable_type" validate:"required,oneof=user organization"`
}

// Validate performs validation on the SetPrimaryAddressRequest using the validator package.
// This method ensures the addressable_id is a valid UUID and the addressable_type
// is one of the allowed values.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *SetPrimaryAddressRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// SetAddressTypeRequest represents the request for changing the type/purpose of an address.
// This request allows updating the address classification (e.g., from "home" to "work")
// without modifying other address details.
type SetAddressTypeRequest struct {
	// AddressType specifies the new purpose of the address (must be valid enum value)
	AddressType string `json:"address_type" validate:"required,oneof=home work billing shipping other"`
}

// Validate performs validation on the SetAddressTypeRequest using the validator package.
// This method ensures the address_type is one of the allowed enumerated values.
//
// Returns:
//   - error: Validation error if the address_type is invalid, nil if valid
func (r *SetAddressTypeRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// SearchAddressesRequest represents the request for searching addresses by various criteria.
// This struct supports filtering addresses by geographic and administrative boundaries,
// with optional pagination parameters for result management.
type SearchAddressesRequest struct {
	// City filters addresses by city or municipality name (optional, 1-255 characters if provided)
	City string `json:"city,omitempty" validate:"omitempty,min=1,max=255"`
	// State filters addresses by state, province, or region (optional, 1-255 characters if provided)
	State string `json:"state,omitempty" validate:"omitempty,min=1,max=255"`
	// Country filters addresses by country name (optional, 1-255 characters if provided)
	Country string `json:"country,omitempty" validate:"omitempty,min=1,max=255"`
	// PostalCode filters addresses by postal or ZIP code (optional, 1-20 characters if provided)
	PostalCode string `json:"postal_code,omitempty" validate:"omitempty,min=1,max=20"`
	// Limit controls the maximum number of results returned (optional, 1-100, defaults to 10)
	Limit int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	// Offset controls the number of results to skip for pagination (optional, minimum 0)
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Validate performs validation on the SearchAddressesRequest using the validator package.
// This method checks field constraints and sets sensible defaults for pagination
// parameters to ensure consistent search behavior.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *SearchAddressesRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Set default pagination values for consistent behavior
	if r.Limit == 0 {
		r.Limit = 10
	}

	return nil
}
