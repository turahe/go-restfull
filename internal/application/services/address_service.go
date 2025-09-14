// Package services provides application-level business logic for address management.
// This package contains the address service implementation that handles address-related
// operations including CRUD operations, validation, and business rules.
// Package services provides application-level business logic for address management.
// This package contains the address service implementation that handles address-related
// operations including CRUD operations, validation, and business rules.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// addressService implements the AddressService interface and provides business logic
// for address management operations. It handles address creation, updates, deletion,
// and various search operations while enforcing business rules and validation.
// addressService implements the AddressService interface and provides business logic
// for address management operations. It handles address creation, updates, deletion,
// and various search operations while enforcing business rules and validation.
type addressService struct {
	addressRepo  repositories.AddressRepository
	mediaService ports.MediaService
}

// NewAddressService creates a new instance of the address service with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - addressRepo: The repository interface for data access operations
//   - mediaService: Media service for handling address images and media
//
// Returns:
//   - ports.AddressService: The service interface implementation
func NewAddressService(addressRepo repositories.AddressRepository, mediaService ports.MediaService) ports.AddressService {
	return &addressService{
		addressRepo:  addressRepo,
		mediaService: mediaService,
	}
}

// CreateAddress creates a new address for a specific addressable entity (user or organization).
// This method enforces business rules such as validation, primary address management,
// and ensures data integrity.
//
// Business Rules:
//   - All required fields must be provided and validated
//   - If the new address is marked as primary, all other primary addresses for the same
//     addressable entity are automatically unset
//   - Address coordinates (latitude/longitude) are optional
//   - Address line 2 is optional
//
// Parameters:
//   - ctx: Context for the operation
//   - address: The address entity to create
//
// Returns:
//   - *entities.Address: The created address entity
//   - error: Any error that occurred during the operation
func (s *addressService) CreateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error) {
	// Validate input parameters to ensure data integrity
	if err := s.ValidateAddress(ctx, address.AddressLine1, address.City, address.State, address.PostalCode, address.Country); err != nil {
		return nil, err
	}

	// Business rule: If this is a primary address, unset other primary addresses
	// for the same addressable entity to maintain data consistency
	if address.IsPrimary {
		err := s.addressRepo.UnsetOtherPrimaries(ctx, address.AddressableID, address.AddressableType, uuid.Nil)
		if err != nil {
			return nil, err
		}
	}

	// Persist the address to the repository
	err := s.addressRepo.Create(ctx, address)
	if err != nil {
		return nil, err
	}

	return address, nil
}

// GetAddressByID retrieves an address by its unique identifier.
// This method includes soft delete checking to ensure deleted addresses
// are not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the address to retrieve
//
// Returns:
//   - *entities.Address: The address entity if found
//   - error: Error if address not found or other issues occur
func (s *addressService) GetAddressByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the address has been soft deleted
	if address.IsDeleted() {
		return nil, errors.New("address not found")
	}

	return address, nil
}

// UpdateAddress updates an existing address with new information.
// This method enforces business rules and maintains data integrity
// during the update process.
//
// Business Rules:
//   - All required fields must be validated
//   - If setting as primary, other primary addresses are automatically unset
//   - Soft deleted addresses cannot be updated
//   - Address coordinates and secondary line are optional
//
// Parameters:
//   - ctx: Context for the operation
//   - address: The address entity to update
//
// Returns:
//   - *entities.Address: The updated address entity
//   - error: Any error that occurred during the operation
func (s *addressService) UpdateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error) {
	// Retrieve existing address to ensure it exists and is not deleted
	existingAddress, err := s.addressRepo.GetByID(ctx, address.ID)
	if err != nil {
		return nil, err
	}

	// Check if the address has been soft deleted
	if existingAddress.IsDeleted() {
		return nil, errors.New("address not found")
	}

	// Validate input parameters to ensure data integrity
	if err := s.ValidateAddress(ctx, address.AddressLine1, address.City, address.State, address.PostalCode, address.Country); err != nil {
		return nil, err
	}

	// Business rule: If this is a primary address, unset other primary addresses
	// for the same addressable entity to maintain data consistency
	if address.IsPrimary && !existingAddress.IsPrimary {
		err := s.addressRepo.UnsetOtherPrimaries(ctx, address.AddressableID, address.AddressableType, address.ID)
		if err != nil {
			return nil, err
		}
	}

	// Update the address in the repository
	err = s.addressRepo.Update(ctx, address)
	if err != nil {
		return nil, err
	}

	return address, nil
}

// DeleteAddress performs a soft delete of an address by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the address to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *addressService) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	// Retrieve existing address to ensure it exists and is not already deleted
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the address has already been soft deleted
	if address.IsDeleted() {
		return errors.New("address not found")
	}

	// Perform soft delete by marking the address as deleted
	return s.addressRepo.Delete(ctx, id)
}

// GetAddressesByAddressable retrieves all addresses for a specific addressable entity
// (user or organization). This method returns all addresses regardless of their
// primary status or type.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//
// Returns:
//   - []*entities.Address: List of addresses for the entity
//   - error: Any error that occurred during the operation
func (s *addressService) GetAddressesByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error) {
	return s.addressRepo.GetByAddressable(ctx, addressableID, addressableType)
}

// GetPrimaryAddressByAddressable retrieves the primary address for a specific
// addressable entity. Each entity can have only one primary address.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//
// Returns:
//   - *entities.Address: The primary address for the entity
//   - error: Any error that occurred during the operation
func (s *addressService) GetPrimaryAddressByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error) {
	return s.addressRepo.GetPrimaryByAddressable(ctx, addressableID, addressableType)
}

// GetAddressesByAddressableAndType retrieves addresses for a specific addressable
// entity filtered by address type (e.g., home, work, billing, shipping).
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//   - addressType: Type of address to filter by
//
// Returns:
//   - []*entities.Address: List of addresses matching the criteria
//   - error: Any error that occurred during the operation
func (s *addressService) GetAddressesByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error) {
	return s.addressRepo.GetByAddressableAndType(ctx, addressableID, addressableType, addressType)
}

// SetPrimaryAddress sets a specific address as the primary address for an
// addressable entity. This operation automatically unsets other primary
// addresses for the same entity to maintain data consistency.
//
// Business Rules:
//   - The address must belong to the specified addressable entity
//   - Other primary addresses for the same entity are automatically unset
//   - Only one address can be primary per entity
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the address to set as primary
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *addressService) SetPrimaryAddress(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error {
	// Retrieve the address to verify it exists and belongs to the specified entity
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the address has been soft deleted
	if address.IsDeleted() {
		return errors.New("address not found")
	}

	// Verify the address belongs to the specified addressable entity
	if address.AddressableID != addressableID || address.AddressableType != addressableType {
		return errors.New("address does not belong to the specified entity")
	}

	// Unset other primary addresses for the same entity
	err = s.addressRepo.UnsetOtherPrimaries(ctx, addressableID, addressableType, id)
	if err != nil {
		return err
	}

	// Set this address as the primary address
	return s.addressRepo.SetPrimary(ctx, id, addressableID, addressableType)
}

// SetAddressType updates the type of an existing address (e.g., home, work, billing).
// This method allows changing the classification of an address without affecting
// other address properties.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the address to update
//   - addressType: New address type to assign
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *addressService) SetAddressType(ctx context.Context, id uuid.UUID, addressType entities.AddressType) error {
	// Retrieve the address to ensure it exists and is not deleted
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the address has been soft deleted
	if address.IsDeleted() {
		return errors.New("address not found")
	}

	// Update the address type
	address.SetAddressType(addressType)

	// Persist the updated address to the repository
	return s.addressRepo.Update(ctx, address)
}

// SearchAddressesByCity searches for addresses in a specific city with pagination.
// This method is useful for location-based queries and reporting.
//
// Parameters:
//   - ctx: Context for the operation
//   - city: City name to search for
//   - limit: Maximum number of results to return
//   - offset: Number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: List of addresses in the specified city
//   - error: Any error that occurred during the operation
func (s *addressService) SearchAddressesByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByCity(ctx, city, limit, offset)
}

// SearchAddressesByState searches for addresses in a specific state/province
// with pagination. This method supports regional reporting and analysis.
//
// Parameters:
//   - ctx: Context for the operation
//   - state: State/province name to search for
//   - limit: Maximum number of results to return
//   - offset: Number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: List of addresses in the specified state
//   - error: Any error that occurred during the operation
func (s *addressService) SearchAddressesByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByState(ctx, state, limit, offset)
}

// SearchAddressesByCountry searches for addresses in a specific country
// with pagination. This method supports international reporting and analysis.
//
// Parameters:
//   - ctx: Context for the operation
//   - country: Country name to search for
//   - limit: Maximum number of results to return
//   - offset: Number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: List of addresses in the specified country
//   - error: Any error that occurred during the operation
func (s *addressService) SearchAddressesByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByCountry(ctx, country, limit, offset)
}

// SearchAddressesByPostalCode searches for addresses with a specific postal/ZIP code
// with pagination. This method is useful for precise location-based queries.
//
// Parameters:
//   - ctx: Context for the operation
//   - postalCode: Postal/ZIP code to search for
//   - limit: Maximum number of results to return
//   - offset: Number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: List of addresses with the specified postal code
//   - error: Any error that occurred during the operation
func (s *addressService) SearchAddressesByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByPostalCode(ctx, postalCode, limit, offset)
}

// GetAddressCountByAddressable returns the total number of addresses for a specific
// addressable entity. This method is useful for statistics and reporting.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//
// Returns:
//   - int64: Total count of addresses for the entity
//   - error: Any error that occurred during the operation
func (s *addressService) GetAddressCountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error) {
	return s.addressRepo.CountByAddressable(ctx, addressableID, addressableType)
}

// GetAddressCountByType returns the total number of addresses of a specific type
// across all entities. This method is useful for system-wide statistics.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressType: Type of address to count
//
// Returns:
//   - int64: Total count of addresses of the specified type
//   - error: Any error that occurred during the operation
func (s *addressService) GetAddressCountByType(ctx context.Context, addressType entities.AddressType) (int64, error) {
	return s.addressRepo.CountByType(ctx, addressType)
}

// GetAddressCountByAddressableAndType returns the total number of addresses of a
// specific type for a specific addressable entity. This method provides detailed
// statistics for reporting and analysis.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//   - addressType: Type of address to count
//
// Returns:
//   - int64: Total count of addresses matching the criteria
//   - error: Any error that occurred during the operation
func (s *addressService) GetAddressCountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error) {
	return s.addressRepo.CountByAddressableAndType(ctx, addressableID, addressableType, addressType)
}

// ValidateAddress performs comprehensive validation of address data to ensure
// data integrity and business rule compliance. This method is called internally
// by other service methods before persisting address data.
//
// Validation Rules:
//   - Address line 1 is required and cannot be empty
//   - City is required and cannot be empty
//   - State/province is required and cannot be empty
//   - Postal/ZIP code is required and cannot be empty
//   - Country is required and cannot be empty
//   - All fields are trimmed of leading/trailing whitespace
//
// Parameters:
//   - ctx: Context for the operation
//   - addressLine1: Primary address line to validate
//   - city: City name to validate
//   - state: State/province name to validate
//   - postalCode: Postal/ZIP code to validate
//   - country: Country name to validate
//
// Returns:
//   - error: Validation error if any rule is violated, nil if validation passes
func (s *addressService) ValidateAddress(ctx context.Context, addressLine1, city, state, postalCode, country string) error {
	// Validate address line 1 (required field)
	if strings.TrimSpace(addressLine1) == "" {
		return errors.New("address line 1 is required")
	}

	// Validate city (required field)
	if strings.TrimSpace(city) == "" {
		return errors.New("city is required")
	}

	// Validate state/province (required field)
	if strings.TrimSpace(state) == "" {
		return errors.New("state is required")
	}

	// Validate postal/ZIP code (required field)
	if strings.TrimSpace(postalCode) == "" {
		return errors.New("postal code is required")
	}

	// Validate country (required field)
	if strings.TrimSpace(country) == "" {
		return errors.New("country is required")
	}

	return nil
}

// CheckAddressExists verifies whether an address with the specified ID exists
// in the system. This method is useful for validation and business logic
// that requires address existence verification.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the address to check
//
// Returns:
//   - bool: True if the address exists, false otherwise
//   - error: Any error that occurred during the operation
func (s *addressService) CheckAddressExists(ctx context.Context, id uuid.UUID) (bool, error) {
	// Note: This implementation uses a placeholder check. In a real implementation,
	// you would want to check by the actual address ID rather than using
	// the addressable existence check as a proxy.
	exists, err := s.addressRepo.ExistsByAddressable(ctx, id, entities.AddressableTypeUser) // Using user as placeholder, actual implementation should check by ID
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CheckAddressableHasAddresses verifies whether a specific addressable entity
// (user or organization) has any addresses associated with it. This method
// is useful for business logic that depends on address existence.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//
// Returns:
//   - bool: True if the entity has addresses, false otherwise
//   - error: Any error that occurred during the operation
func (s *addressService) CheckAddressableHasAddresses(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	return s.addressRepo.ExistsByAddressable(ctx, addressableID, addressableType)
}

// CheckAddressableHasPrimaryAddress verifies whether a specific addressable entity
// (user or organization) has a primary address. This method is useful for
// business logic that requires primary address verification.
//
// Parameters:
//   - ctx: Context for the operation
//   - addressableID: UUID of the entity (user/organization)
//   - addressableType: Type of the addressable entity
//
// Returns:
//   - bool: True if the entity has a primary address, false otherwise
//   - error: Any error that occurred during the operation
func (s *addressService) CheckAddressableHasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	return s.addressRepo.HasPrimaryAddress(ctx, addressableID, addressableType)
}

// AttachAddressImage attaches an existing media file to an address as an image.
// This method creates a relationship between a media file and an address entity,
// allowing the address to have visual representation.
//
// Business Rules:
//   - Address must exist and be accessible
//   - Media file must exist and be accessible
//   - Media file must be an image (validated by MIME type)
//   - Only one image per address is allowed (replaces existing)
//
// Parameters:
//   - ctx: Context for the operation
//   - addressID: UUID of the address to attach the image to
//   - mediaID: UUID of the media file to attach
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *addressService) AttachAddressImage(ctx context.Context, addressID uuid.UUID, mediaID uuid.UUID) error {
	if s.mediaService == nil {
		return errors.New("media service not available")
	}

	// Validate that the address exists
	address, err := s.GetAddressByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("address not found: %w", err)
	}
	if address == nil {
		return fmt.Errorf("address with ID %s not found", addressID.String())
	}

	// Validate that the media exists and is an image
	media, err := s.mediaService.GetMediaByID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}
	if media == nil {
		return fmt.Errorf("media with ID %s not found", mediaID.String())
	}

	// Validate that the media is an image
	if !media.IsImage() {
		return fmt.Errorf("media file must be an image, got MIME type: %s", media.MimeType)
	}

	// Attach media to address as an image
	err = s.mediaService.AttachMediaToEntity(ctx, mediaID, addressID, "Address", "image")
	if err != nil {
		return fmt.Errorf("failed to attach image to address: %w", err)
	}

	return nil
}
