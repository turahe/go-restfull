package services

import (
	"context"
	"fmt"
	"github.com/turahe/go-restfull/pkg/rabbitmq"
	"log"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/services"
)

// EmailConsumerService handles email processing from RabbitMQ
type EmailConsumerService struct {
	rabbitMQService *rabbitmq.Service
	emailService    services.EmailService // Original email service for actual sending
}

// NewEmailConsumerService creates a new email consumer service
func NewEmailConsumerService(
	rabbitMQService *rabbitmq.Service,
	emailService services.EmailService,
) *EmailConsumerService {
	return &EmailConsumerService{
		rabbitMQService: rabbitMQService,
		emailService:    emailService,
	}
}

// StartEmailConsumer starts consuming email events from RabbitMQ
func (s *EmailConsumerService) StartEmailConsumer(ctx context.Context) error {
	return s.rabbitMQService.ConsumeWithType(ctx, "email.sending", &entities.EmailEvent{}, func(ctx context.Context, data interface{}) error {
		emailEvent := data.(*entities.EmailEvent)
		return s.processEmailEvent(ctx, emailEvent)
	})
}

// processEmailEvent processes an email event and sends the actual email
func (s *EmailConsumerService) processEmailEvent(ctx context.Context, event *entities.EmailEvent) error {
	log.Printf("Processing email event: Type=%s, To=%s, Subject=%s", event.Type, event.To, event.Subject)

	switch event.Type {
	case "welcome":
		return s.sendWelcomeEmail(event)
	case "password_reset":
		return s.sendPasswordResetEmail(event)
	case "user_registration":
		return s.sendUserRegistrationEmail(event)
	case "custom":
		return s.sendCustomEmail(event)
	case "template":
		return s.sendTemplateEmail(event)
	default:
		return fmt.Errorf("unknown email type: %s", event.Type)
	}
}

// sendWelcomeEmail sends a welcome email
func (s *EmailConsumerService) sendWelcomeEmail(event *entities.EmailEvent) error {
	username, ok := event.Data["username"].(string)
	if !ok {
		return fmt.Errorf("username not found in email event data")
	}

	// Use the original email service to send the actual email
	return s.emailService.SendWelcomeEmail(event.To, username)
}

// sendPasswordResetEmail sends a password reset email
func (s *EmailConsumerService) sendPasswordResetEmail(event *entities.EmailEvent) error {
	otp, ok := event.Data["otp"].(string)
	if !ok {
		return fmt.Errorf("OTP not found in email event data")
	}

	// Use the original email service to send the actual email
	return s.emailService.SendPasswordResetEmail(event.To, otp)
}

// sendUserRegistrationEmail sends a user registration email
func (s *EmailConsumerService) sendUserRegistrationEmail(event *entities.EmailEvent) error {
	username, ok := event.Data["username"].(string)
	if !ok {
		return fmt.Errorf("username not found in email event data")
	}

	// Use the original email service to send the actual email
	// For user registration, we'll send a welcome email
	return s.emailService.SendWelcomeEmail(event.To, username)
}

// sendCustomEmail sends a custom email
func (s *EmailConsumerService) sendCustomEmail(event *entities.EmailEvent) error {
	// For custom emails, we might need to implement a method in the original email service
	// For now, we'll log the email content
	log.Printf("Sending custom email to %s: %s", event.To, event.Subject)
	log.Printf("Email body: %s", event.Body)

	// You can implement actual email sending here
	// For example, using a third-party email service like SendGrid, AWS SES, etc.

	return nil
}

// sendTemplateEmail sends a template email
func (s *EmailConsumerService) sendTemplateEmail(event *entities.EmailEvent) error {
	// For template emails, we might need to implement a method in the original email service
	log.Printf("Sending template email to %s: %s (Template: %s)", event.To, event.Subject, event.Template)

	// You can implement template-based email sending here
	// For example, using HTML templates with data interpolation

	return nil
}

// StartEmailConsumerWithRetry starts the email consumer with retry logic
func (s *EmailConsumerService) StartEmailConsumerWithRetry(ctx context.Context, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := s.StartEmailConsumer(ctx); err != nil {
			lastErr = err
			log.Printf("Failed to start email consumer (attempt %d/%d): %v", i+1, maxRetries, err)

			if i < maxRetries-1 {
				// Wait before retrying
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		} else {
			log.Println("Email consumer started successfully")
			return nil
		}
	}

	return fmt.Errorf("failed to start email consumer after %d attempts: %w", maxRetries, lastErr)
}

// GetEmailQueueInfo gets information about the email queue
func (s *EmailConsumerService) GetEmailQueueInfo(ctx context.Context) (map[string]interface{}, error) {
	return s.rabbitMQService.GetQueueInfo(ctx, "email.sending")
}

// HealthCheck performs a health check on the email consumer service
func (s *EmailConsumerService) HealthCheck(ctx context.Context) error {
	return s.rabbitMQService.HealthCheck(ctx)
}
