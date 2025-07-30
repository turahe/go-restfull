package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterAddressRoutes registers all address-related routes
func RegisterAddressRoutes(protected fiber.Router, container *container.Container) {
	addressController := container.GetAddressController()

	// Address routes (protected)
	addresses := protected.Group("/addresses")
	addresses.Post("/", addressController.CreateAddress)
	addresses.Get("/:id", addressController.GetAddressByID)
	addresses.Put("/:id", addressController.UpdateAddress)
	addresses.Delete("/:id", addressController.DeleteAddress)
	addresses.Put("/:id/primary", addressController.SetPrimaryAddress)
	addresses.Put("/:id/type", addressController.SetAddressType)
	addresses.Get("/addressable/:addressable_type/:addressable_id", addressController.GetAddressesByAddressable)
	addresses.Get("/addressable/:addressable_type/:addressable_id/primary", addressController.GetPrimaryAddressByAddressable)
	addresses.Get("/addressable/:addressable_type/:addressable_id/type/:address_type", addressController.GetAddressesByAddressableAndType)
	addresses.Get("/search/city", addressController.SearchAddressesByCity)
	addresses.Get("/search/state", addressController.SearchAddressesByState)
	addresses.Get("/search/country", addressController.SearchAddressesByCountry)
	addresses.Get("/search/postal-code", addressController.SearchAddressesByPostalCode)
}
