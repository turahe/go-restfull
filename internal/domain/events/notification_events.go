package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// NotificationCreatedEvent represents an event when a notification is created
type NotificationCreatedEvent struct {
	ID        uuid.UUID                      `json:"id"`
	UserID    uuid.UUID                      `json:"user_id"`
	Type      entities.NotificationType      `json:"type"`
	Title     string                         `json:"title"`
	Message   string                         `json:"message"`
	Data      map[string]interface{}         `json:"data"`
	Priority  entities.NotificationPriority  `json:"priority"`
	Channels  []entities.NotificationChannel `json:"channels"`
	CreatedAt time.Time                      `json:"created_at"`
}

// NotificationReadEvent represents an event when a notification is marked as read
type NotificationReadEvent struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	ReadAt time.Time `json:"read_at"`
}

// NotificationArchivedEvent represents an event when a notification is archived
type NotificationArchivedEvent struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ArchivedAt time.Time `json:"archived_at"`
}

// NotificationDeletedEvent represents an event when a notification is deleted
type NotificationDeletedEvent struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// NotificationDeliveryEvent represents an event when a notification is delivered
type NotificationDeliveryEvent struct {
	ID             uuid.UUID                    `json:"id"`
	NotificationID uuid.UUID                    `json:"notification_id"`
	Channel        entities.NotificationChannel `json:"channel"`
	Status         string                       `json:"status"`
	DeliveredAt    time.Time                    `json:"delivered_at"`
}

// NotificationPreferenceUpdatedEvent represents an event when notification preferences are updated
type NotificationPreferenceUpdatedEvent struct {
	ID        uuid.UUID                 `json:"id"`
	UserID    uuid.UUID                 `json:"user_id"`
	Type      entities.NotificationType `json:"type"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

// BulkNotificationEvent represents an event for sending multiple notifications
type BulkNotificationEvent struct {
	UserIDs   []uuid.UUID                    `json:"user_ids"`
	Type      entities.NotificationType      `json:"type"`
	Title     string                         `json:"title"`
	Message   string                         `json:"message"`
	Data      map[string]interface{}         `json:"data"`
	Priority  entities.NotificationPriority  `json:"priority"`
	Channels  []entities.NotificationChannel `json:"channels"`
	CreatedAt time.Time                      `json:"created_at"`
}
