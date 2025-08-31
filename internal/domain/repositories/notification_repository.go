package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/pagination"
)

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, notification *entities.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error)
	Update(ctx context.Context, notification *entities.Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// User-specific operations
	GetByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetUnreadByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetReadByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetArchivedByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)

	// Status operations
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	MarkAsUnread(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	Archive(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	Unarchive(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Bulk operations
	MarkMultipleAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error
	MarkMultipleAsUnread(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error
	ArchiveMultiple(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error

	// Count operations
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// Search and filter operations
	GetByType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetByPriority(ctx context.Context, userID uuid.UUID, priority entities.NotificationPriority, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)
	GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate string, pagination *pagination.PaginationRequest) ([]*entities.Notification, error)

	// Cleanup operations
	DeleteExpired(ctx context.Context) error
	DeleteOldArchived(ctx context.Context, daysOld int) error
}

// NotificationTemplateRepository defines the interface for notification template operations
type NotificationTemplateRepository interface {
	Create(ctx context.Context, template *entities.NotificationTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.NotificationTemplate, error)
	GetByName(ctx context.Context, name string) (*entities.NotificationTemplate, error)
	GetByType(ctx context.Context, notificationType entities.NotificationType) (*entities.NotificationTemplate, error)
	Update(ctx context.Context, template *entities.NotificationTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, pagination *pagination.PaginationRequest) ([]*entities.NotificationTemplate, error)
}

// NotificationPreferenceRepository defines the interface for notification preference operations
type NotificationPreferenceRepository interface {
	Create(ctx context.Context, preference *entities.NotificationPreference) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.NotificationPreference, error)
	GetByUserIDAndType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType) (*entities.NotificationPreference, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.NotificationPreference, error)
	Update(ctx context.Context, preference *entities.NotificationPreference) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// Bulk operations
	CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error
	UpdateMultiplePreferences(ctx context.Context, preferences []*entities.NotificationPreference) error
}

// NotificationDeliveryRepository defines the interface for notification delivery operations
type NotificationDeliveryRepository interface {
	Create(ctx context.Context, delivery *entities.NotificationDelivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.NotificationDelivery, error)
	GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]*entities.NotificationDelivery, error)
	Update(ctx context.Context, delivery *entities.NotificationDelivery) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Status operations
	MarkAsDelivered(ctx context.Context, id uuid.UUID, deliveredAt string) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, error string) error
	IncrementAttempts(ctx context.Context, id uuid.UUID) error

	// Cleanup operations
	DeleteOldDeliveries(ctx context.Context, daysOld int) error
}
