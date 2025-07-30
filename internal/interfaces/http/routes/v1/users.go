package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(protected fiber.Router, container *container.Container) {
	userController := container.GetUserController()
	menuController := container.GetMenuController()

	// User routes (protected)
	users := protected.Group("/users")
	users.Post("/", userController.CreateUser)
	users.Get("/", userController.GetUsers)
	users.Get("/:id", userController.GetUserByID)
	users.Put("/:id", userController.UpdateUser)
	users.Delete("/:id", userController.DeleteUser)
	users.Put("/:id/password", userController.ChangePassword)
	users.Get("/profile", userController.GetUserByID)
	users.Put("/profile", userController.UpdateUser)

	// User-Menu routes (protected)
	users.Get("/:user_id/menus", menuController.GetUserMenus)
}
