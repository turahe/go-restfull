package entities

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// User-related notifications
	NotificationTypeUserRegistration  NotificationType = "user_registration"
	NotificationTypePasswordChange    NotificationType = "password_change"
	NotificationTypeEmailVerification NotificationType = "email_verification"
	NotificationTypeProfileUpdate     NotificationType = "profile_update"

	// System notifications
	NotificationTypeSystemAlert   NotificationType = "system_alert"
	NotificationTypeMaintenance   NotificationType = "maintenance"
	NotificationTypeSecurityAlert NotificationType = "security_alert"

	// Content notifications
	NotificationTypeNewPost      NotificationType = "new_post"
	NotificationTypeCommentReply NotificationType = "comment_reply"
	NotificationTypeMention      NotificationType = "mention"
	NotificationTypeLike         NotificationType = "like"

	// Organization notifications
	NotificationTypeInvitation       NotificationType = "invitation"
	NotificationTypeRoleChange       NotificationType = "role_change"
	NotificationTypePermissionUpdate NotificationType = "permission_update"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// NotificationStatus represents the current status of a notification
type NotificationStatus string

const (
	NotificationStatusUnread   NotificationStatus = "unread"
	NotificationStatusRead     NotificationStatus = "read"
	NotificationStatusArchived NotificationStatus = "archived"
	NotificationStatusDeleted  NotificationStatus = "deleted"
)

// NotificationChannel represents the delivery channel for a notification
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelPush    NotificationChannel = "push"
	NotificationChannelInApp   NotificationChannel = "in_app"
	NotificationChannelWebhook NotificationChannel = "webhook"
)

// Notification represents a notification entity
type Notification struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	UserID     uuid.UUID              `json:"user_id" db:"user_id"`
	Type       NotificationType       `json:"type" db:"type"`
	Title      string                 `json:"title" db:"title"`
	Message    string                 `json:"message" db:"message"`
	Data       map[string]interface{} `json:"data" db:"data"`
	Priority   NotificationPriority   `json:"priority" db:"priority"`
	Status     NotificationStatus     `json:"status" db:"status"`
	Channels   []NotificationChannel  `json:"channels" db:"channels"`
	ReadAt     *time.Time             `json:"read_at" db:"read_at"`
	ArchivedAt *time.Time             `json:"archived_at" db:"archived_at"`
	ExpiresAt  *time.Time             `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time             `json:"deleted_at" db:"deleted_at"`

	// Relationships
	User *User `json:"user,omitempty"`
}

// NotificationTemplate represents a template for generating notifications
type NotificationTemplate struct {
	ID        uuid.UUID             `json:"id" db:"id"`
	Name      string                `json:"name" db:"name"`
	Type      NotificationType      `json:"type" db:"type"`
	Title     string                `json:"title" db:"title"`
	Message   string                `json:"message" db:"message"`
	Subject   string                `json:"subject" db:"subject"` // For email notifications
	Channels  []NotificationChannel `json:"channels" db:"channels"`
	Priority  NotificationPriority  `json:"priority" db:"priority"`
	IsActive  bool                  `json:"is_active" db:"is_active"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt time.Time             `json:"updated_at" db:"updated_at"`
}

// NotificationPreference represents user preferences for notifications
type NotificationPreference struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"user_id" db:"user_id"`
	Type      NotificationType `json:"type" db:"type"`
	Email     bool             `json:"email" db:"email"`
	SMS       bool             `json:"sms" db:"sms"`
	Push      bool             `json:"push" db:"push"`
	InApp     bool             `json:"in_app" db:"in_app"`
	Webhook   bool             `json:"webhook" db:"webhook"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`

	// Relationships
	User *User `json:"user,omitempty"`
}

// NotificationDelivery represents the delivery status of a notification
type NotificationDelivery struct {
	ID             uuid.UUID           `json:"id" db:"id"`
	NotificationID uuid.UUID           `json:"notification_id" db:"notification_id"`
	Channel        NotificationChannel `json:"channel" db:"channel"`
	Status         string              `json:"status" db:"status"` // sent, failed, pending
	Attempts       int                 `json:"attempts" db:"attempts"`
	LastAttemptAt  *time.Time          `json:"last_attempt_at" db:"last_attempt_at"`
	DeliveredAt    *time.Time          `json:"delivered_at" db:"delivered_at"`
	Error          string              `json:"error" db:"error"`
	CreatedAt      time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at" db:"updated_at"`

	// Relationships
	Notification *Notification `json:"notification,omitempty"`
}
