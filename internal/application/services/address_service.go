package services

import (
	"context"
	"errors"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

type addressService struct {
	addressRepo repositories.AddressRepository
}

func NewAddressService(addressRepo repositories.AddressRepository) ports.AddressService {
	return &addressService{
		addressRepo: addressRepo,
	}
}

func (s *addressService) CreateAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressLine1, city, state, postalCode, country string, addressLine2 *string, latitude, longitude *float64, isPrimary bool, addressType entities.AddressType) (*entities.Address, error) {
	// Validate input
	if err := s.ValidateAddress(ctx, addressLine1, city, state, postalCode, country); err != nil {
		return nil, err
	}

	// If this is a primary address, unset other primary addresses for this addressable
	if isPrimary {
		err := s.addressRepo.UnsetOtherPrimaries(ctx, addressableID, addressableType, uuid.Nil)
		if err != nil {
			return nil, err
		}
	}

	// Create new address
	address := entities.NewAddress(
		addressableID,
		addressableType,
		addressLine1,
		city,
		state,
		postalCode,
		country,
		addressLine2,
		latitude,
		longitude,
		isPrimary,
		addressType,
	)

	// Save to repository
	err := s.addressRepo.Create(ctx, address)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *addressService) GetAddressByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if address.IsDeleted() {
		return nil, errors.New("address not found")
	}

	return address, nil
}

func (s *addressService) UpdateAddress(ctx context.Context, id uuid.UUID, addressLine1, city, state, postalCode, country string, addressLine2 *string, latitude, longitude *float64, isPrimary bool, addressType entities.AddressType) (*entities.Address, error) {
	// Get existing address
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if address.IsDeleted() {
		return nil, errors.New("address not found")
	}

	// Validate input
	if err := s.ValidateAddress(ctx, addressLine1, city, state, postalCode, country); err != nil {
		return nil, err
	}

	// If setting as primary, unset other primary addresses
	if isPrimary && !address.IsPrimary {
		err := s.addressRepo.UnsetOtherPrimaries(ctx, address.AddressableID, address.AddressableType, id)
		if err != nil {
			return nil, err
		}
	}

	// Update address
	address.UpdateAddress(addressLine1, city, state, postalCode, country, addressLine2, latitude, longitude, isPrimary, addressType)

	// Save to repository
	err = s.addressRepo.Update(ctx, address)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *addressService) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	// Get existing address
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if address.IsDeleted() {
		return errors.New("address not found")
	}

	// Soft delete
	return s.addressRepo.Delete(ctx, id)
}

func (s *addressService) GetAddressesByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error) {
	return s.addressRepo.GetByAddressable(ctx, addressableID, addressableType)
}

func (s *addressService) GetPrimaryAddressByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error) {
	return s.addressRepo.GetPrimaryByAddressable(ctx, addressableID, addressableType)
}

func (s *addressService) GetAddressesByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error) {
	return s.addressRepo.GetByAddressableAndType(ctx, addressableID, addressableType, addressType)
}

func (s *addressService) SetPrimaryAddress(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error {
	// Get existing address
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if address.IsDeleted() {
		return errors.New("address not found")
	}

	// Verify the address belongs to the specified addressable
	if address.AddressableID != addressableID || address.AddressableType != addressableType {
		return errors.New("address does not belong to the specified entity")
	}

	// Unset other primary addresses
	err = s.addressRepo.UnsetOtherPrimaries(ctx, addressableID, addressableType, id)
	if err != nil {
		return err
	}

	// Set this address as primary
	return s.addressRepo.SetPrimary(ctx, id, addressableID, addressableType)
}

func (s *addressService) SetAddressType(ctx context.Context, id uuid.UUID, addressType entities.AddressType) error {
	// Get existing address
	address, err := s.addressRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if address.IsDeleted() {
		return errors.New("address not found")
	}

	// Update address type
	address.SetAddressType(addressType)

	// Save to repository
	return s.addressRepo.Update(ctx, address)
}

func (s *addressService) SearchAddressesByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByCity(ctx, city, limit, offset)
}

func (s *addressService) SearchAddressesByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByState(ctx, state, limit, offset)
}

func (s *addressService) SearchAddressesByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByCountry(ctx, country, limit, offset)
}

func (s *addressService) SearchAddressesByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error) {
	return s.addressRepo.SearchByPostalCode(ctx, postalCode, limit, offset)
}

func (s *addressService) GetAddressCountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error) {
	return s.addressRepo.CountByAddressable(ctx, addressableID, addressableType)
}

func (s *addressService) GetAddressCountByType(ctx context.Context, addressType entities.AddressType) (int64, error) {
	return s.addressRepo.CountByType(ctx, addressType)
}

func (s *addressService) GetAddressCountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error) {
	return s.addressRepo.CountByAddressableAndType(ctx, addressableID, addressableType, addressType)
}

func (s *addressService) ValidateAddress(ctx context.Context, addressLine1, city, state, postalCode, country string) error {
	if strings.TrimSpace(addressLine1) == "" {
		return errors.New("address line 1 is required")
	}

	if strings.TrimSpace(city) == "" {
		return errors.New("city is required")
	}

	if strings.TrimSpace(state) == "" {
		return errors.New("state is required")
	}

	if strings.TrimSpace(postalCode) == "" {
		return errors.New("postal code is required")
	}

	if strings.TrimSpace(country) == "" {
		return errors.New("country is required")
	}

	return nil
}

func (s *addressService) CheckAddressExists(ctx context.Context, id uuid.UUID) (bool, error) {
	exists, err := s.addressRepo.ExistsByAddressable(ctx, id, entities.AddressableTypeUser) // Using user as placeholder, actual implementation should check by ID
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *addressService) CheckAddressableHasAddresses(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	return s.addressRepo.ExistsByAddressable(ctx, addressableID, addressableType)
}

func (s *addressService) CheckAddressableHasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	return s.addressRepo.HasPrimaryAddress(ctx, addressableID, addressableType)
}
