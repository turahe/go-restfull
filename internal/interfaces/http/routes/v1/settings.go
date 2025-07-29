package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterSettingRoutes registers all setting-related routes
// TODO: Implement when SettingController is created
func RegisterSettingRoutes(protected fiber.Router, container *container.Container) {
	// settingController := container.GetSettingController()

	// Setting routes (protected)
	// settings := protected.Group("/settings")
	// settings.Get("/", settingController.GetSettings)
	// settings.Get("/:key", settingController.GetSettingByKey)
	// settings.Post("/", settingController.CreateSetting)
	// settings.Put("/:key", settingController.UpdateSetting)
	// settings.Delete("/:key", settingController.DeleteSetting)
}
