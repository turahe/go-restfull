package services

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/rabbitmq"
)

// EmailService handles email operations using RabbitMQ
type EmailService struct {
	rabbitMQService *rabbitmq.Service
}

// NewEmailService creates a new email service
func NewEmailService(rabbitMQService *rabbitmq.Service) *EmailService {
	return &EmailService{
		rabbitMQService: rabbitMQService,
	}
}

// SendWelcomeEmail sends a welcome email via RabbitMQ
func (s *EmailService) SendWelcomeEmail(email, username string) error {
	ctx := context.Background()

	// Create welcome email event
	event := entities.NewWelcomeEmailEvent(email, username)

	// Publish to email queue
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendPasswordResetEmail sends a password reset email via RabbitMQ
func (s *EmailService) SendPasswordResetEmail(email, otp string) error {
	ctx := context.Background()

	// Create password reset email event
	event := entities.NewPasswordResetEmailEvent(email, otp)

	// Publish to email queue
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendUserRegistrationEmail sends a user registration email via RabbitMQ
func (s *EmailService) SendUserRegistrationEmail(userID, username, email string) error {
	ctx := context.Background()

	// Create user registration email event
	event := entities.NewUserRegistrationEvent(userID, username, email)

	// Publish to email queue
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendCustomEmail sends a custom email via RabbitMQ
func (s *EmailService) SendCustomEmail(to, subject, body string, headers map[string]string) error {
	ctx := context.Background()

	event := &entities.EmailEvent{
		Type:      "custom",
		To:        to,
		Subject:   subject,
		Body:      body,
		Headers:   headers,
		Timestamp: time.Now(),
		Priority:  "normal",
	}

	// Publish to email queue
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendEmailWithTemplate sends an email with a template via RabbitMQ
func (s *EmailService) SendEmailWithTemplate(to, subject, template string, data map[string]interface{}) error {
	ctx := context.Background()

	event := &entities.EmailEvent{
		Type:      "template",
		To:        to,
		Subject:   subject,
		Template:  template,
		Data:      data,
		Timestamp: time.Now(),
		Priority:  "normal",
	}

	// Publish to email queue
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendBulkEmail sends multiple emails via RabbitMQ
func (s *EmailService) SendBulkEmail(emails []entities.EmailEvent) error {
	ctx := context.Background()

	for _, email := range emails {
		if err := s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", email); err != nil {
			return fmt.Errorf("failed to publish email to %s: %w", email.To, err)
		}
	}

	return nil
}

// GetEmailQueueInfo gets information about the email queue
func (s *EmailService) GetEmailQueueInfo(ctx context.Context) (map[string]interface{}, error) {
	return s.rabbitMQService.GetQueueInfo(ctx, "email.sending")
}

// HealthCheck performs a health check on the email service
func (s *EmailService) HealthCheck(ctx context.Context) error {
	return s.rabbitMQService.HealthCheck(ctx)
}
