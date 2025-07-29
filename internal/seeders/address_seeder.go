package seeders

import (
	"context"
	"log"

	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
)

type AddressSeeder struct {
	addressService ports.AddressService
	userService    ports.UserService
}

func NewAddressSeeder(addressService ports.AddressService, userService ports.UserService) *AddressSeeder {
	return &AddressSeeder{
		addressService: addressService,
		userService:    userService,
	}
}

func (s *AddressSeeder) Seed(ctx context.Context) error {
	log.Println("Seeding addresses...")

	// Get some existing users to add addresses to
	users, err := s.userService.GetAllUsers(ctx, 10, 0)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		log.Println("No users found, skipping address seeding")
		return nil
	}

	// Sample address data
	sampleAddresses := []struct {
		addressLine1 string
		addressLine2 *string
		city         string
		state        string
		postalCode   string
		country      string
		latitude     *float64
		longitude    *float64
		isPrimary    bool
		addressType  entities.AddressType
	}{
		{
			addressLine1: "123 Main Street",
			addressLine2: nil,
			city:         "New York",
			state:        "NY",
			postalCode:   "10001",
			country:      "USA",
			latitude:     float64Ptr(40.7505),
			longitude:    float64Ptr(-73.9934),
			isPrimary:    true,
			addressType:  entities.AddressTypeHome,
		},
		{
			addressLine1: "456 Business Avenue",
			addressLine2: stringPtr("Suite 100"),
			city:         "New York",
			state:        "NY",
			postalCode:   "10002",
			country:      "USA",
			latitude:     float64Ptr(40.7168),
			longitude:    float64Ptr(-73.9861),
			isPrimary:    false,
			addressType:  entities.AddressTypeWork,
		},
		{
			addressLine1: "789 Oak Drive",
			addressLine2: nil,
			city:         "Los Angeles",
			state:        "CA",
			postalCode:   "90210",
			country:      "USA",
			latitude:     float64Ptr(34.1030),
			longitude:    float64Ptr(-118.4105),
			isPrimary:    true,
			addressType:  entities.AddressTypeHome,
		},
		{
			addressLine1: "321 Corporate Plaza",
			addressLine2: stringPtr("Floor 15"),
			city:         "Los Angeles",
			state:        "CA",
			postalCode:   "90012",
			country:      "USA",
			latitude:     float64Ptr(34.0522),
			longitude:    float64Ptr(-118.2437),
			isPrimary:    false,
			addressType:  entities.AddressTypeWork,
		},
		{
			addressLine1: "654 Pine Street",
			addressLine2: nil,
			city:         "Chicago",
			state:        "IL",
			postalCode:   "60601",
			country:      "USA",
			latitude:     float64Ptr(41.8781),
			longitude:    float64Ptr(-87.6298),
			isPrimary:    true,
			addressType:  entities.AddressTypeHome,
		},
		{
			addressLine1: "987 Tech Center",
			addressLine2: stringPtr("Building A"),
			city:         "Chicago",
			state:        "IL",
			postalCode:   "60602",
			country:      "USA",
			latitude:     float64Ptr(41.8857),
			longitude:    float64Ptr(-87.6228),
			isPrimary:    false,
			addressType:  entities.AddressTypeWork,
		},
		{
			addressLine1: "147 Maple Avenue",
			addressLine2: nil,
			city:         "Houston",
			state:        "TX",
			postalCode:   "77001",
			country:      "USA",
			latitude:     float64Ptr(29.7604),
			longitude:    float64Ptr(-95.3698),
			isPrimary:    true,
			addressType:  entities.AddressTypeHome,
		},
		{
			addressLine1: "258 Energy Tower",
			addressLine2: stringPtr("Suite 500"),
			city:         "Houston",
			state:        "TX",
			postalCode:   "77002",
			country:      "USA",
			latitude:     float64Ptr(29.7604),
			longitude:    float64Ptr(-95.3698),
			isPrimary:    false,
			addressType:  entities.AddressTypeWork,
		},
		{
			addressLine1: "369 Sunset Boulevard",
			addressLine2: nil,
			city:         "Phoenix",
			state:        "AZ",
			postalCode:   "85001",
			country:      "USA",
			latitude:     float64Ptr(33.4484),
			longitude:    float64Ptr(-112.0740),
			isPrimary:    true,
			addressType:  entities.AddressTypeHome,
		},
		{
			addressLine1: "741 Innovation Hub",
			addressLine2: stringPtr("Floor 8"),
			city:         "Phoenix",
			state:        "AZ",
			postalCode:   "85002",
			country:      "USA",
			latitude:     float64Ptr(33.4484),
			longitude:    float64Ptr(-112.0740),
			isPrimary:    false,
			addressType:  entities.AddressTypeWork,
		},
	}

	// Add addresses for each user
	for i, user := range users {
		if i >= len(sampleAddresses) {
			break
		}

		addr := sampleAddresses[i]
		_, err := s.addressService.CreateAddress(
			ctx,
			user.ID,
			entities.AddressableTypeUser,
			addr.addressLine1,
			addr.city,
			addr.state,
			addr.postalCode,
			addr.country,
			addr.addressLine2,
			addr.latitude,
			addr.longitude,
			addr.isPrimary,
			addr.addressType,
		)
		if err != nil {
			log.Printf("Error creating address for user %s: %v", user.ID, err)
			continue
		}

		// Add a second address (work address) for some users
		if i%2 == 0 && i+1 < len(sampleAddresses) {
			workAddr := sampleAddresses[i+1]
			_, err := s.addressService.CreateAddress(
				ctx,
				user.ID,
				entities.AddressableTypeUser,
				workAddr.addressLine1,
				workAddr.city,
				workAddr.state,
				workAddr.postalCode,
				workAddr.country,
				workAddr.addressLine2,
				workAddr.latitude,
				workAddr.longitude,
				workAddr.isPrimary,
				workAddr.addressType,
			)
			if err != nil {
				log.Printf("Error creating work address for user %s: %v", user.ID, err)
			}
		}
	}

	log.Printf("Successfully seeded %d addresses", len(users))
	return nil
}

func (s *AddressSeeder) Clean(ctx context.Context) error {
	log.Println("Cleaning addresses...")
	// Note: In a real implementation, you might want to add a method to clean addresses
	// For now, we'll just log that cleaning is not implemented
	log.Println("Address cleaning not implemented")
	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
