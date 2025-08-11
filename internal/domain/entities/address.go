// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Address entity and related types
// for managing address information with polymorphic relationships.
package entities

import (
	"time"

	"github.com/google/uuid"
)

// AddressType represents the category or purpose of an address.
// This enum provides predefined values for common address types used
// throughout the application for categorization and filtering purposes.
type AddressType string

// Address type constants defining the standard categories for addresses.
// These values are used to classify addresses based on their intended use.
const (
	AddressTypeHome     AddressType = "home"     // Residential/home address
	AddressTypeWork     AddressType = "work"     // Business/workplace address
	AddressTypeBilling  AddressType = "billing"  // Billing/invoice address
	AddressTypeShipping AddressType = "shipping" // Shipping/delivery address
	AddressTypeOther    AddressType = "other"    // Miscellaneous/other address types
)

// AddressableType represents the type of entity that can have addresses.
// This enum enables polymorphic relationships where addresses can be
// associated with different types of entities (users, organizations, etc.).
type AddressableType string

// Addressable type constants defining which entities can have addresses.
// This supports the polymorphic relationship pattern in the database.
const (
	AddressableTypeUser         AddressableType = "user"         // User entity can have addresses
	AddressableTypeOrganization AddressableType = "organization" // Organization entity can have addresses
)

// Address represents a physical location with comprehensive location details.
// It supports polymorphic relationships through addressable_id and addressable_type
// fields, allowing addresses to be associated with different entity types.
//
// The entity includes:
// - Standard address components (street, city, state, postal code, country)
// - Optional geographic coordinates for mapping and distance calculations
// - Address categorization and primary status
// - Audit trail with creation, update, and deletion tracking
// - Soft delete support for data retention
type Address struct {
	ID              uuid.UUID       `json:"id"`                      // Unique identifier for the address
	AddressableID   uuid.UUID       `json:"addressable_id"`          // ID of the entity this address belongs to
	AddressableType AddressableType `json:"addressable_type"`        // Type of entity (user, organization, etc.)
	AddressLine1    string          `json:"address_line1"`           // Primary street address line
	AddressLine2    *string         `json:"address_line2,omitempty"` // Optional secondary address line (apt, suite, etc.)
	City            string          `json:"city"`                    // City or municipality name
	State           string          `json:"state"`                   // State, province, or region
	PostalCode      string          `json:"postal_code"`             // Postal/ZIP code
	Country         string          `json:"country"`                 // Country name
	Latitude        *float64        `json:"latitude,omitempty"`      // Optional latitude coordinate for mapping
	Longitude       *float64        `json:"longitude,omitempty"`     // Optional longitude coordinate for mapping
	IsPrimary       bool            `json:"is_primary"`              // Whether this is the primary address for the entity
	AddressType     AddressType     `json:"address_type"`            // Category/purpose of this address
	CreatedBy       uuid.UUID       `json:"created_by"`              // ID of user who created this address
	UpdatedBy       uuid.UUID       `json:"updated_by"`              // ID of user who last updated this address
	DeletedBy       *uuid.UUID      `json:"deleted_by,omitempty"`    // ID of user who deleted this address (soft delete)
	CreatedAt       time.Time       `json:"created_at"`              // Timestamp when address was created
	UpdatedAt       time.Time       `json:"updated_at"`              // Timestamp when address was last updated
	DeletedAt       *time.Time      `json:"deleted_at,omitempty"`    // Timestamp when address was soft deleted
}

// NewAddress creates a new Address instance with the provided details.
// This constructor initializes required fields and sets default values
// for timestamps and generates a new UUID for the address.
//
// Parameters:
//   - addressableID: UUID of the entity this address belongs to
//   - addressableType: Type of entity (user, organization, etc.)
//   - addressLine1: Primary street address line
//   - city: City or municipality name
//   - state: State, province, or region
//   - postalCode: Postal/ZIP code
//   - country: Country name
//   - addressLine2: Optional secondary address line (can be nil)
//   - latitude: Optional latitude coordinate (can be nil)
//   - longitude: Optional longitude coordinate (can be nil)
//   - isPrimary: Whether this is the primary address
//   - addressType: Category/purpose of this address
//
// Returns:
//   - *Address: Pointer to the newly created address entity
func NewAddress(
	addressableID uuid.UUID,
	addressableType AddressableType,
	addressLine1, city, state, postalCode, country string,
	addressLine2 *string,
	latitude, longitude *float64,
	isPrimary bool,
	addressType AddressType,
) *Address {
	return &Address{
		ID:              uuid.New(),      // Generate new unique identifier
		AddressableID:   addressableID,   // Set the owning entity ID
		AddressableType: addressableType, // Set the entity type
		AddressLine1:    addressLine1,    // Set primary address line
		AddressLine2:    addressLine2,    // Set optional secondary line
		City:            city,            // Set city
		State:           state,           // Set state/province
		PostalCode:      postalCode,      // Set postal code
		Country:         country,         // Set country
		Latitude:        latitude,        // Set optional latitude
		Longitude:       longitude,       // Set optional longitude
		IsPrimary:       isPrimary,       // Set primary status
		AddressType:     addressType,     // Set address type
		CreatedAt:       time.Now(),      // Set creation timestamp
		UpdatedAt:       time.Now(),      // Set initial update timestamp
	}
}

// UpdateAddress updates the address details with new values.
// This method modifies the address fields and automatically updates
// the UpdatedAt timestamp to reflect the change.
//
// Parameters:
//   - addressLine1: New primary street address line
//   - city: New city or municipality name
//   - state: New state, province, or region
//   - postalCode: New postal/ZIP code
//   - country: New country name
//   - addressLine2: New optional secondary address line (can be nil)
//   - latitude: New optional latitude coordinate (can be nil)
//   - longitude: New optional longitude coordinate (can be nil)
//   - isPrimary: New primary address status
//   - addressType: New address type category
//
// Note: This method automatically updates the UpdatedAt timestamp
func (a *Address) UpdateAddress(
	addressLine1, city, state, postalCode, country string,
	addressLine2 *string,
	latitude, longitude *float64,
	isPrimary bool,
	addressType AddressType,
) {
	// Update all address fields with new values
	a.AddressLine1 = addressLine1
	a.AddressLine2 = addressLine2
	a.City = city
	a.State = state
	a.PostalCode = postalCode
	a.Country = country
	a.Latitude = latitude
	a.Longitude = longitude
	a.IsPrimary = isPrimary
	a.AddressType = addressType

	// Update the modification timestamp
	a.UpdatedAt = time.Now()
}

// SetPrimary sets whether this address is the primary address for the entity.
// Only one address per entity should typically be marked as primary.
// This method automatically updates the UpdatedAt timestamp.
//
// Parameters:
//   - isPrimary: Boolean indicating if this should be the primary address
//
// Note: This method automatically updates the UpdatedAt timestamp
func (a *Address) SetPrimary(isPrimary bool) {
	a.IsPrimary = isPrimary
	a.UpdatedAt = time.Now()
}

// SetAddressType changes the category/purpose of this address.
// This method automatically updates the UpdatedAt timestamp.
//
// Parameters:
//   - addressType: New address type category
//
// Note: This method automatically updates the UpdatedAt timestamp
func (a *Address) SetAddressType(addressType AddressType) {
	a.AddressType = addressType
	a.UpdatedAt = time.Now()
}

// SoftDelete marks the address as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The address will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (a *Address) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now // Set deletion timestamp
	a.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the address has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
//
// Returns:
//   - bool: true if address is deleted, false if active
func (a *Address) IsDeleted() bool {
	return a.DeletedAt != nil
}

// IsPrimaryAddress checks if this address is marked as the primary address
// for the associated entity.
//
// Returns:
//   - bool: true if this is the primary address, false otherwise
func (a *Address) IsPrimaryAddress() bool {
	return a.IsPrimary
}

// GetFullAddress returns a formatted string representation of the complete address.
// This method concatenates all address components into a human-readable format,
// handling optional fields gracefully.
//
// Returns:
//   - string: Formatted full address string
//
// Format: "AddressLine1, AddressLine2, City, State PostalCode, Country"
// Note: AddressLine2 is only included if it exists and is not empty
func (a *Address) GetFullAddress() string {
	// Start with the primary address line
	address := a.AddressLine1

	// Add secondary line if it exists and is not empty
	if a.AddressLine2 != nil && *a.AddressLine2 != "" {
		address += ", " + *a.AddressLine2
	}

	// Add city, state, postal code, and country
	address += ", " + a.City + ", " + a.State + " " + a.PostalCode + ", " + a.Country

	return address
}

// HasCoordinates checks if the address has valid geographic coordinates.
// Returns true if both latitude and longitude are set, false otherwise.
// This is useful for determining if the address can be used for mapping
// or distance calculations.
//
// Returns:
//   - bool: true if both coordinates are available, false otherwise
func (a *Address) HasCoordinates() bool {
	return a.Latitude != nil && a.Longitude != nil
}
