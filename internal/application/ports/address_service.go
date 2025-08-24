package ports

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

type AddressService interface {
	// Basic CRUD operations
	CreateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error)
	GetAddressByID(ctx context.Context, id uuid.UUID) (*entities.Address, error)
	UpdateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error)
	DeleteAddress(ctx context.Context, id uuid.UUID) error

	// Get addresses by addressable entity
	GetAddressesByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error)
	GetPrimaryAddressByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error)
	GetAddressesByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error)

	// Address management
	SetPrimaryAddress(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error
	SetAddressType(ctx context.Context, id uuid.UUID, addressType entities.AddressType) error

	// Search and filtering
	SearchAddressesByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error)
	SearchAddressesByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error)
	SearchAddressesByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error)
	SearchAddressesByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error)

	// Count operations
	GetAddressCountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error)
	GetAddressCountByType(ctx context.Context, addressType entities.AddressType) (int64, error)
	GetAddressCountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error)

	// Validation
	ValidateAddress(ctx context.Context, addressLine1, city, state, postalCode, country string) error
	CheckAddressExists(ctx context.Context, id uuid.UUID) (bool, error)
	CheckAddressableHasAddresses(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error)
	CheckAddressableHasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error)
}
