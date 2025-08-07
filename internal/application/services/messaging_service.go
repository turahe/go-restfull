package services

import (
	"context"
	"fmt"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/infrastructure/messaging"

	"go.uber.org/zap"
)

// MessagingServiceImpl implements MessagingService interface
type MessagingServiceImpl struct {
	rabbitMQClient *messaging.RabbitMQClient
	handlerFactory *messaging.HandlerFactory
}

// NewMessagingService creates a new messaging service
func NewMessagingService(rabbitMQClient *messaging.RabbitMQClient) ports.MessagingService {
	return &MessagingServiceImpl{
		rabbitMQClient: rabbitMQClient,
		handlerFactory: messaging.NewHandlerFactory(),
	}
}

// SendEmail sends an email message to the email queue
func (s *MessagingServiceImpl) SendEmail(ctx context.Context, to, subject, body, html string) error {
	emailMsg := messaging.EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
		HTML:    html,
	}

	message, err := messaging.NewMessage("email", emailMsg, 3, 0, 1)
	if err != nil {
		return fmt.Errorf("failed to create email message: %w", err)
	}

	err = s.rabbitMQClient.PublishMessage(ctx, "email_queue", message)
	if err != nil {
		return fmt.Errorf("failed to publish email message: %w", err)
	}

	return nil
}

// SendBackup sends a backup message to the backup queue
func (s *MessagingServiceImpl) SendBackup(ctx context.Context, database, path string, compress bool) error {
	backupMsg := messaging.BackupMessage{
		Database: database,
		Path:     path,
		Compress: compress,
	}

	message, err := messaging.NewMessage("backup", backupMsg, 3, 0, 2)
	if err != nil {
		return fmt.Errorf("failed to create backup message: %w", err)
	}

	err = s.rabbitMQClient.PublishMessage(ctx, "backup_queue", message)
	if err != nil {
		return fmt.Errorf("failed to publish backup message: %w", err)
	}

	return nil
}

// SendCleanup sends a cleanup message to the cleanup queue
func (s *MessagingServiceImpl) SendCleanup(ctx context.Context, cleanupType string, olderThanDays int) error {
	cleanupMsg := messaging.CleanupMessage{
		Type:      cleanupType,
		OlderThan: olderThanDays,
	}

	message, err := messaging.NewMessage("cleanup", cleanupMsg, 3, 0, 1)
	if err != nil {
		return fmt.Errorf("failed to create cleanup message: %w", err)
	}

	err = s.rabbitMQClient.PublishMessage(ctx, "cleanup_queue", message)
	if err != nil {
		return fmt.Errorf("failed to publish cleanup message: %w", err)
	}

	return nil
}

// SendNotification sends a notification message to the notification queue
func (s *MessagingServiceImpl) SendNotification(ctx context.Context, userID, notificationType, title, message string, data map[string]interface{}) error {
	notificationMsg := messaging.NotificationMessage{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Data:    data,
	}

	msg, err := messaging.NewMessage("notification", notificationMsg, 3, 0, 1)
	if err != nil {
		return fmt.Errorf("failed to create notification message: %w", err)
	}

	err = s.rabbitMQClient.PublishMessage(ctx, "notification_queue", msg)
	if err != nil {
		return fmt.Errorf("failed to publish notification message: %w", err)
	}

	return nil
}

// StartConsumers starts all message consumers
func (s *MessagingServiceImpl) StartConsumers(ctx context.Context) error {
	// Start email consumer
	go func() {
		err := s.rabbitMQClient.ConsumeMessages(ctx, "email_queue", s.createHandler("email"))
		if err != nil {
			zap.L().Error("Failed to start email consumer", zap.Error(err))
		}
	}()

	// Start backup consumer
	go func() {
		err := s.rabbitMQClient.ConsumeMessages(ctx, "backup_queue", s.createHandler("backup"))
		if err != nil {
			zap.L().Error("Failed to start backup consumer", zap.Error(err))
		}
	}()

	// Start cleanup consumer
	go func() {
		err := s.rabbitMQClient.ConsumeMessages(ctx, "cleanup_queue", s.createHandler("cleanup"))
		if err != nil {
			zap.L().Error("Failed to start cleanup consumer", zap.Error(err))
		}
	}()

	// Start notification consumer
	go func() {
		err := s.rabbitMQClient.ConsumeMessages(ctx, "notification_queue", s.createHandler("notification"))
		if err != nil {
			zap.L().Error("Failed to start notification consumer", zap.Error(err))
		}
	}()

	return nil
}

// createHandler creates a handler for a specific message type
func (s *MessagingServiceImpl) createHandler(messageType string) messaging.MessageHandler {
	handler, err := s.handlerFactory.CreateHandler(messageType)
	if err != nil {
		zap.L().Error("Failed to create handler",
			zap.String("message_type", messageType),
			zap.Error(err),
		)
		return messaging.NewNoOpHandler()
	}
	return handler
}

// Close closes the messaging service
func (s *MessagingServiceImpl) Close() error {
	return s.rabbitMQClient.Close()
}
