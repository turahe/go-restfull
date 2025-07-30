package repositories

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

type AddressRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, address *entities.Address) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Address, error)
	Update(ctx context.Context, address *entities.Address) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Get addresses by addressable entity
	GetByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error)
	GetPrimaryByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error)
	GetByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error)

	// Address management
	SetPrimary(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error
	UnsetOtherPrimaries(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, excludeID uuid.UUID) error

	// Search and filtering
	SearchByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error)
	SearchByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error)
	SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error)
	SearchByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error)

	// Count operations
	CountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error)
	CountByType(ctx context.Context, addressType entities.AddressType) (int64, error)
	CountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error)

	// Validation
	ExistsByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error)
	HasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error)
}
