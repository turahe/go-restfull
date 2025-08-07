package ports

import (
	"context"
)

// MessagingService defines the interface for messaging operations
type MessagingService interface {
	SendEmail(ctx context.Context, to, subject, body, html string) error
	SendBackup(ctx context.Context, database, path string, compress bool) error
	SendCleanup(ctx context.Context, cleanupType string, olderThanDays int) error
	SendNotification(ctx context.Context, userID, notificationType, title, message string, data map[string]interface{}) error
	StartConsumers(ctx context.Context) error
	Close() error
}
