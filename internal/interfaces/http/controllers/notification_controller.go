package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/helper/pagination"
	"github.com/turahe/go-restfull/pkg/logger"
	"go.uber.org/zap")

// NotificationController handles HTTP requests for notifications
type NotificationController struct {
	notificationService services.NotificationService
	templateService     services.NotificationTemplateService
	preferenceService   services.NotificationPreferenceService
	deliveryService     services.NotificationDeliveryService
}

// NewNotificationController creates a new notification controller
func NewNotificationController(
	notificationService services.NotificationService,
	templateService services.NotificationTemplateService,
	preferenceService services.NotificationPreferenceService,
	deliveryService services.NotificationDeliveryService,
) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		templateService:     templateService,
		preferenceService:   preferenceService,
		deliveryService:     deliveryService,
	}
}

// getUserIDFromContext extracts user ID from Fiber context
func (c *NotificationController) getUserIDFromContext(ctx *fiber.Ctx) (uuid.UUID, error) {
	userIDInterface := ctx.Locals("user_id")
	if userIDInterface == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusInternalServerError, "Invalid user ID format")
	}

	return userID, nil
}

// GetUserNotifications gets all notifications for the authenticated user
func (c *NotificationController) GetUserNotifications(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Query("per_page", "10"))
	sortBy := ctx.Query("sort_by", "created_at")
	sortDesc := ctx.Query("sort_desc", "true") == "true"
	search := ctx.Query("search", "")

	paginationReq := &pagination.PaginationRequest{
		Page:     page,
		PerPage:  perPage,
		SortBy:   sortBy,
		SortDesc: sortDesc,
		Search:   search,
	}

	notifications, err := c.notificationService.GetUserNotifications(ctx.Context(), userID, paginationReq)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get notifications",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   notifications,
	})
}

// GetUserUnreadNotifications gets unread notifications for the authenticated user
func (c *NotificationController) GetUserUnreadNotifications(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Query("per_page", "10"))

	paginationReq := &pagination.PaginationRequest{
		Page:     page,
		PerPage:  perPage,
		SortBy:   "created_at",
		SortDesc: true,
	}

	notifications, err := c.notificationService.GetUserUnreadNotifications(ctx.Context(), userID, paginationReq)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get unread notifications",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   notifications,
	})
}

// GetNotificationByID gets a specific notification by ID
func (c *NotificationController) GetNotificationByID(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	notificationID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	notification, err := c.notificationService.GetNotificationByID(ctx.Context(), notificationID, userID)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Notification not found",
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   notification,
	})
}

// MarkAsRead marks a notification as read
func (c *NotificationController) MarkAsRead(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	notificationID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	err = c.notificationService.MarkAsRead(ctx.Context(), notificationID, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to mark notification as read",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notification marked as read",
	})
}

// MarkAsUnread marks a notification as unread
func (c *NotificationController) MarkAsUnread(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	notificationID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	err = c.notificationService.MarkAsUnread(ctx.Context(), notificationID, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to mark notification as unread",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notification marked as unread",
	})
}

// ArchiveNotification archives a notification
func (c *NotificationController) ArchiveNotification(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	notificationID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	err = c.notificationService.ArchiveNotification(ctx.Context(), notificationID, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to archive notification",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notification archived",
	})
}

// DeleteNotification deletes a notification
func (c *NotificationController) DeleteNotification(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	notificationID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	err = c.notificationService.DeleteNotification(ctx.Context(), notificationID, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete notification",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notification deleted",
	})
}

// GetNotificationCount gets the count of notifications for the authenticated user
func (c *NotificationController) GetNotificationCount(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	totalCount, err := c.notificationService.GetUserNotificationCount(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get notification count",
			"error":   err.Error(),
		})
	}

	unreadCount, err := c.notificationService.GetUserUnreadNotificationCount(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get unread notification count",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_count":  totalCount,
			"unread_count": unreadCount,
		},
	})
}

// BulkMarkAsRead marks multiple notifications as read
func (c *NotificationController) BulkMarkAsRead(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	var request struct {
		IDs []string `json:"ids"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Basic validation
	if len(request.IDs) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "At least one notification ID is required",
		})
	}

	// Convert string IDs to UUIDs
	var ids []uuid.UUID
	for _, idStr := range request.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid notification ID format",
			})
	}
		ids = append(ids, id)
	}

	err = c.notificationService.MarkMultipleAsRead(ctx.Context(), ids, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to mark notifications as read",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notifications marked as read",
	})
}

// GetUserPreferences gets notification preferences for the authenticated user
func (c *NotificationController) GetUserPreferences(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	preferences, err := c.preferenceService.GetAllUserPreferences(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get notification preferences",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   preferences,
	})
}

// UpdateUserPreferences updates notification preferences for the authenticated user
func (c *NotificationController) UpdateUserPreferences(ctx *fiber.Ctx) error {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	var preferences []entities.NotificationPreference
	if err := ctx.BodyParser(&preferences); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Basic validation
	if len(preferences) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "At least one preference is required",
		})
	}

	// Set user ID for all preferences
	for i := range preferences {
		preferences[i].UserID = userID
	}

	// err = c.preferenceService.UpdateMultiplePreferences(ctx.Context(), preferences)
	// if err != nil {
	// 	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": "Failed to update notification preferences",
	// 		"error":   err.Error(),
	// 	})
	// }

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notification preferences updated",
	})
}
