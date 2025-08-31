package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/pagination"
)

// Error constants for notification services
var (
	ErrInvalidTemplateName     = errors.New("invalid template name")
	ErrInvalidTemplateTitle    = errors.New("invalid template title")
	ErrInvalidTemplateMessage  = errors.New("invalid template message")
	ErrInvalidTemplateChannels = errors.New("invalid template channels")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrInvalidNotificationType = errors.New("invalid notification type")
)

// NotificationService defines the interface for notification business logic
type NotificationService interface {
	// Basic CRUD operations
	CreateNotification(ctx context.Context, notification *entities.Notification) error
	GetNotificationByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.Notification, error)
	UpdateNotification(ctx context.Context, notification *entities.Notification, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// User-specific operations
	GetUserNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetUserUnreadNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetUserReadNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetUserArchivedNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)

	// Status operations
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	MarkAsUnread(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ArchiveNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UnarchiveNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Bulk operations
	MarkMultipleAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error
	MarkMultipleAsUnread(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error
	ArchiveMultipleNotifications(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error

	// Count operations
	GetUserNotificationCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetUserUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// Search and filter operations
	GetUserNotificationsByType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetUserNotificationsByPriority(ctx context.Context, userID uuid.UUID, priority entities.NotificationPriority, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetUserNotificationsByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate string, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)

	// Template-based notifications
	SendNotificationFromTemplate(ctx context.Context, userID uuid.UUID, templateName string, data map[string]interface{}) error
	SendBulkNotificationFromTemplate(ctx context.Context, userIDs []uuid.UUID, templateName string, data map[string]interface{}) error

	// Custom notifications
	SendCustomNotification(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType, title, message string, data map[string]interface{}, priority entities.NotificationPriority, channels []entities.NotificationChannel) error
	SendBulkCustomNotification(ctx context.Context, userIDs []uuid.UUID, notificationType entities.NotificationType, title, message string, data map[string]interface{}, priority entities.NotificationPriority, channels []entities.NotificationChannel) error

	// System notifications
	SendSystemNotification(ctx context.Context, userIDs []uuid.UUID, title, message string, priority entities.NotificationPriority) error
	SendMaintenanceNotification(ctx context.Context, userIDs []uuid.UUID, title, message string, scheduledTime string) error
	SendSecurityAlert(ctx context.Context, userIDs []uuid.UUID, title, message string) error

	// Cleanup operations
	CleanupExpiredNotifications(ctx context.Context) error
	CleanupOldArchivedNotifications(ctx context.Context, daysOld int) error
}

// NotificationTemplateService defines the interface for notification template business logic
type NotificationTemplateService interface {
	CreateTemplate(ctx context.Context, template *entities.NotificationTemplate) error
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*entities.NotificationTemplate, error)
	GetTemplateByName(ctx context.Context, name string) (*entities.NotificationTemplate, error)
	GetTemplateByType(ctx context.Context, notificationType entities.NotificationType) (*entities.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, template *entities.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	GetAllTemplates(ctx context.Context, pagination *pagination.PaginationRequest) ([]*entities.NotificationTemplate, error)

	// Template validation
	ValidateTemplate(ctx context.Context, template *entities.NotificationTemplate) error
	TestTemplate(ctx context.Context, templateID uuid.UUID, testData map[string]interface{}) (string, error)
}

// NotificationPreferenceService defines the interface for notification preference business logic
type NotificationPreferenceService interface {
	CreatePreference(ctx context.Context, preference *entities.NotificationPreference) error
	GetPreferenceByID(ctx context.Context, id uuid.UUID) (*entities.NotificationPreference, error)
	GetUserPreferenceByType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType) (*entities.NotificationPreference, error)
	GetAllUserPreferences(ctx context.Context, userID uuid.UUID) ([]*entities.NotificationPreference, error)
	UpdatePreference(ctx context.Context, preference *entities.NotificationPreference) error
	DeletePreference(ctx context.Context, id uuid.UUID) error
	DeleteUserPreferences(ctx context.Context, userID uuid.UUID) error

	// Bulk operations
	CreateDefaultUserPreferences(ctx context.Context, userID uuid.UUID) error
	UpdateMultipleUserPreferences(ctx context.Context, preferences []*entities.NotificationPreference) error

	// Preference validation
	ValidatePreferences(ctx context.Context, preferences []*entities.NotificationPreference) error
}

// NotificationDeliveryService defines the interface for notification delivery business logic
type NotificationDeliveryService interface {
	CreateDelivery(ctx context.Context, delivery *entities.NotificationDelivery) error
	GetDeliveryByID(ctx context.Context, id uuid.UUID) (*entities.NotificationDelivery, error)
	GetDeliveriesByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]*entities.NotificationDelivery, error)
	UpdateDelivery(ctx context.Context, delivery *entities.NotificationDelivery) error
	DeleteDelivery(ctx context.Context, id uuid.UUID) error

	// Delivery status operations
	MarkAsDelivered(ctx context.Context, id uuid.UUID, deliveredAt string) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, error string) error
	IncrementDeliveryAttempts(ctx context.Context, id uuid.UUID) error

	// Delivery processing
	ProcessPendingDeliveries(ctx context.Context) error
	RetryFailedDeliveries(ctx context.Context, maxAttempts int) error

	// Cleanup operations
	CleanupOldDeliveries(ctx context.Context, daysOld int) error
}
