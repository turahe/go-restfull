package v1

import (
	"fmt"

	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(protected fiber.Router, container *container.Container) {
	userController := container.GetUserController()
	menuController := container.GetMenuController()

	if userController == nil {
		fmt.Printf("Skipping user routes: userController is nil\n")
		return
	}

	// User routes (protected)
	users := protected.Group("/users")
	users.Post("/", userController.CreateUser)
	users.Get("/", userController.GetUsers)

	// Profile routes (protected) - must come before /:id routes
	users.Get("/profile", userController.GetProfile)
	users.Put("/profile", userController.UpdateProfile)

	// User-Menu routes (protected) - must come before /:id routes
	users.Get("/:user_id/menus", menuController.GetUserMenus)

	// User CRUD routes (protected) - parameterized routes come last
	users.Get("/:id", userController.GetUserByID)
	users.Put("/:id", userController.UpdateUser)
	users.Delete("/:id", userController.DeleteUser)
	users.Put("/:id/password", userController.ChangePassword)
}
