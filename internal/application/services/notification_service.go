package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/helper/pagination"
)

// NotificationService implements the notification service interface
type NotificationService struct {
	notificationRepo repositories.NotificationRepository
	templateRepo     repositories.NotificationTemplateRepository
	preferenceRepo   repositories.NotificationPreferenceRepository
	deliveryRepo     repositories.NotificationDeliveryRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	templateRepo repositories.NotificationTemplateRepository,
	preferenceRepo repositories.NotificationPreferenceRepository,
	deliveryRepo repositories.NotificationDeliveryRepository,
) services.NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		preferenceRepo:   preferenceRepo,
		deliveryRepo:     deliveryRepo,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *entities.Notification) error {
	notification.ID = uuid.New()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	notification.Status = entities.NotificationStatusUnread

	return s.notificationRepo.Create(ctx, notification)
}

// GetNotificationByID gets a notification by ID
func (s *NotificationService) GetNotificationByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.Notification, error) {
	return s.notificationRepo.GetByID(ctx, id)
}

// UpdateNotification updates a notification
func (s *NotificationService) UpdateNotification(ctx context.Context, notification *entities.Notification, userID uuid.UUID) error {
	notification.UpdatedAt = time.Now()
	return s.notificationRepo.Update(ctx, notification)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.Delete(ctx, id)
}

// GetUserNotifications gets all notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetByUserID(ctx, userID, pagination)
}

// GetUserUnreadNotifications gets unread notifications for a user
func (s *NotificationService) GetUserUnreadNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetUnreadByUserID(ctx, userID, pagination)
}

// GetUserReadNotifications gets read notifications for a user
func (s *NotificationService) GetUserReadNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetReadByUserID(ctx, userID, pagination)
}

// GetUserArchivedNotifications gets archived notifications for a user
func (s *NotificationService) GetUserArchivedNotifications(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetArchivedByUserID(ctx, userID, pagination)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.MarkAsRead(ctx, id, userID)
}

// MarkAsUnread marks a notification as unread
func (s *NotificationService) MarkAsUnread(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.MarkAsUnread(ctx, id, userID)
}

// ArchiveNotification archives a notification
func (s *NotificationService) ArchiveNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.Archive(ctx, id, userID)
}

// UnarchiveNotification unarchives a notification
func (s *NotificationService) UnarchiveNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.Unarchive(ctx, id, userID)
}

// MarkMultipleAsRead marks multiple notifications as read
func (s *NotificationService) MarkMultipleAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.MarkMultipleAsRead(ctx, ids, userID)
}

// MarkMultipleAsUnread marks multiple notifications as unread
func (s *NotificationService) MarkMultipleAsUnread(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.MarkMultipleAsUnread(ctx, ids, userID)
}

// ArchiveMultipleNotifications archives multiple notifications
func (s *NotificationService) ArchiveMultipleNotifications(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	return s.notificationRepo.ArchiveMultiple(ctx, ids, userID)
}

// GetUserNotificationCount gets the total count of notifications for a user
func (s *NotificationService) GetUserNotificationCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.CountByUserID(ctx, userID)
}

// GetUserUnreadNotificationCount gets the count of unread notifications for a user
func (s *NotificationService) GetUserUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.CountUnreadByUserID(ctx, userID)
}

// GetUserNotificationsByType gets notifications by type for a user
func (s *NotificationService) GetUserNotificationsByType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetByType(ctx, userID, notificationType, pagination)
}

// GetUserNotificationsByPriority gets notifications by priority for a user
func (s *NotificationService) GetUserNotificationsByPriority(ctx context.Context, userID uuid.UUID, priority entities.NotificationPriority, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetByPriority(ctx, userID, priority, pagination)
}

// GetUserNotificationsByDateRange gets notifications by date range for a user
func (s *NotificationService) GetUserNotificationsByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate string, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	return s.notificationRepo.GetByDateRange(ctx, userID, startDate, endDate, pagination)
}

// SendNotificationFromTemplate sends a notification using a template
func (s *NotificationService) SendNotificationFromTemplate(ctx context.Context, userID uuid.UUID, templateName string, data map[string]interface{}) error {
	// Get template
	template, err := s.templateRepo.GetByName(ctx, templateName)
	if err != nil {
		return err
	}

	// Create notification from template
	notification := &entities.Notification{
		UserID:   userID,
		Type:     template.Type,
		Title:    template.Title,
		Message:  template.Message,
		Data:     data,
		Priority: template.Priority,
		Channels: template.Channels,
	}

	return s.CreateNotification(ctx, notification)
}

// SendBulkNotificationFromTemplate sends notifications to multiple users using a template
func (s *NotificationService) SendBulkNotificationFromTemplate(ctx context.Context, userIDs []uuid.UUID, templateName string, data map[string]interface{}) error {
	for _, userID := range userIDs {
		if err := s.SendNotificationFromTemplate(ctx, userID, templateName, data); err != nil {
			// Log error but continue with other users
			continue
		}
	}
	return nil
}

// SendCustomNotification sends a custom notification
func (s *NotificationService) SendCustomNotification(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType, title, message string, data map[string]interface{}, priority entities.NotificationPriority, channels []entities.NotificationChannel) error {
	notification := &entities.Notification{
		UserID:   userID,
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Data:     data,
		Priority: priority,
		Channels: channels,
	}

	return s.CreateNotification(ctx, notification)
}

// SendBulkCustomNotification sends custom notifications to multiple users
func (s *NotificationService) SendBulkCustomNotification(ctx context.Context, userIDs []uuid.UUID, notificationType entities.NotificationType, title, message string, data map[string]interface{}, priority entities.NotificationPriority, channels []entities.NotificationChannel) error {
	for _, userID := range userIDs {
		if err := s.SendCustomNotification(ctx, userID, notificationType, title, message, data, priority, channels); err != nil {
			// Log error but continue with other users
			continue
		}
	}
	return nil
}

// SendSystemNotification sends a system notification to multiple users
func (s *NotificationService) SendSystemNotification(ctx context.Context, userIDs []uuid.UUID, title, message string, priority entities.NotificationPriority) error {
	return s.SendBulkCustomNotification(ctx, userIDs, entities.NotificationTypeSystemAlert, title, message, nil, priority, []entities.NotificationChannel{entities.NotificationChannelInApp})
}

// SendMaintenanceNotification sends a maintenance notification
func (s *NotificationService) SendMaintenanceNotification(ctx context.Context, userIDs []uuid.UUID, title, message string, scheduledTime string) error {
	data := map[string]interface{}{
		"scheduled_time": scheduledTime,
	}
	return s.SendBulkCustomNotification(ctx, userIDs, entities.NotificationTypeMaintenance, title, message, data, entities.NotificationPriorityHigh, []entities.NotificationChannel{entities.NotificationChannelInApp, entities.NotificationChannelEmail})
}

// SendSecurityAlert sends a security alert
func (s *NotificationService) SendSecurityAlert(ctx context.Context, userIDs []uuid.UUID, title, message string) error {
	return s.SendBulkCustomNotification(ctx, userIDs, entities.NotificationTypeSecurityAlert, title, message, nil, entities.NotificationPriorityUrgent, []entities.NotificationChannel{entities.NotificationChannelInApp, entities.NotificationChannelEmail})
}

// CleanupExpiredNotifications cleans up expired notifications
func (s *NotificationService) CleanupExpiredNotifications(ctx context.Context) error {
	return s.notificationRepo.DeleteExpired(ctx)
}

// CleanupOldArchivedNotifications cleans up old archived notifications
func (s *NotificationService) CleanupOldArchivedNotifications(ctx context.Context, daysOld int) error {
	return s.notificationRepo.DeleteOldArchived(ctx, daysOld)
}
