package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"
	"github.com/turahe/go-restfull/internal/router/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterNotificationRoutes registers notification-related routes
func RegisterNotificationRoutes(v1Group fiber.Router, container *container.Container) {
	notificationController := controllers.NewNotificationController(
		container.NotificationService,
		container.NotificationTemplateService,
		container.NotificationPreferenceService,
		container.NotificationDeliveryService,
	)

	// Notification routes (require authentication)
	notifications := v1Group.Group("/notifications", middleware.JWTAuth(), middleware.RBACMiddleware(container.RBACService))

	// Get notifications
	notifications.Get("/", notificationController.GetUserNotifications)
	notifications.Get("/unread", notificationController.GetUserUnreadNotifications)
	notifications.Get("/count", notificationController.GetNotificationCount)

	// Individual notification operations
	notifications.Get("/:id", notificationController.GetNotificationByID)
	notifications.Put("/:id/read", notificationController.MarkAsRead)
	notifications.Put("/:id/unread", notificationController.MarkAsUnread)
	notifications.Put("/:id/archive", notificationController.ArchiveNotification)
	notifications.Delete("/:id", notificationController.DeleteNotification)

	// Bulk operations
	notifications.Post("/bulk/read", notificationController.BulkMarkAsRead)

	// Notification preferences
	notifications.Get("/preferences", notificationController.GetUserPreferences)
	notifications.Put("/preferences", notificationController.UpdateUserPreferences)
}
