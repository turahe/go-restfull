package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterBackupRoutes(router fiber.Router, container *container.Container) {
	backupController := controllers.NewBackupController()

	// Backup routes
	backup := router.Group("/backup")
	{
		// Create a new backup
		backup.Post("/create", backupController.CreateBackup)

		// Create backup with custom parameters
		backup.Post("/create-custom", backupController.CreateBackupWithPayload)

		// Get backup statistics
		backup.Get("/stats", backupController.GetBackupStats)

		// Cleanup old backups
		backup.Post("/cleanup", backupController.CleanupOldBackups)
	}
}
