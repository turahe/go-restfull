// Package services provides application-level business logic for email consumer processing.
// This package contains the email consumer service implementation that handles email
// event processing from RabbitMQ queues, email type routing, and actual email delivery
// while ensuring reliable asynchronous email processing.
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/turahe/go-restfull/pkg/rabbitmq"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/services"
)

// EmailConsumerService handles email processing from RabbitMQ queues.
// This service consumes email events from the message queue and processes
// them based on their type, ensuring reliable asynchronous email delivery.
type EmailConsumerService struct {
	rabbitMQService *rabbitmq.Service
	emailService    services.EmailService // Original email service for actual sending
}

// NewEmailConsumerService creates a new email consumer service instance with
// RabbitMQ integration and email service dependencies.
// This function follows the dependency injection pattern to ensure loose coupling
// between the consumer layer and the messaging infrastructure.
//
// Parameters:
//   - rabbitMQService: RabbitMQ service for queue consumption
//   - emailService: Email service for actual email delivery
//
// Returns:
//   - *EmailConsumerService: The email consumer service instance
func NewEmailConsumerService(
	rabbitMQService *rabbitmq.Service,
	emailService services.EmailService,
) *EmailConsumerService {
	return &EmailConsumerService{
		rabbitMQService: rabbitMQService,
		emailService:    emailService,
	}
}

// StartEmailConsumer starts consuming email events from the RabbitMQ queue.
// This method establishes a connection to the email queue and processes
// incoming email events asynchronously.
//
// Business Rules:
//   - Consumes from "email.sending" queue
//   - Processes email events based on their type
//   - Handles different email types (welcome, password_reset, etc.)
//   - Logs processing activities for monitoring
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) StartEmailConsumer(ctx context.Context) error {
	return s.rabbitMQService.ConsumeWithType(ctx, "email.sending", &entities.EmailEvent{}, func(ctx context.Context, data interface{}) error {
		emailEvent := data.(*entities.EmailEvent)
		return s.processEmailEvent(ctx, emailEvent)
	})
}

// processEmailEvent processes an email event and routes it to the appropriate
// email sending method based on the event type.
//
// Business Rules:
//   - Validates email event type
//   - Routes to appropriate sending method
//   - Handles unknown email types with error
//   - Logs processing activities for monitoring
//
// Parameters:
//   - ctx: Context for the operation
//   - event: Email event to process
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) processEmailEvent(ctx context.Context, event *entities.EmailEvent) error {
	log.Printf("Processing email event: Type=%s, To=%s, Subject=%s", event.Type, event.To, event.Subject)

	// Route email event to appropriate handler based on type
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

// sendWelcomeEmail sends a welcome email to new users.
// This method extracts username from the email event data and sends
// a personalized welcome email.
//
// Business Rules:
//   - Username must be present in event data
//   - Uses original email service for actual sending
//   - Handles missing data gracefully
//
// Parameters:
//   - event: Email event containing welcome email data
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) sendWelcomeEmail(event *entities.EmailEvent) error {
	// Extract username from event data for personalization
	username, ok := event.Data["username"].(string)
	if !ok {
		return fmt.Errorf("username not found in email event data")
	}

	// Use the original email service to send the actual email
	return s.emailService.SendWelcomeEmail(event.To, username)
}

// sendPasswordResetEmail sends a password reset email with OTP.
// This method extracts OTP from the email event data and sends
// a secure password reset email.
//
// Business Rules:
//   - OTP must be present in event data
//   - Uses original email service for actual sending
//   - Handles missing data gracefully
//
// Parameters:
//   - event: Email event containing password reset data
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) sendPasswordResetEmail(event *entities.EmailEvent) error {
	// Extract OTP from event data for secure password reset
	otp, ok := event.Data["otp"].(string)
	if !ok {
		return fmt.Errorf("OTP not found in email event data")
	}

	// Use the original email service to send the actual email
	return s.emailService.SendPasswordResetEmail(event.To, otp)
}

// sendUserRegistrationEmail sends a user registration confirmation email.
// This method extracts username from the email event data and sends
// a registration confirmation email.
//
// Business Rules:
//   - Username must be present in event data
//   - Uses welcome email template for registration confirmation
//   - Handles missing data gracefully
//
// Parameters:
//   - event: Email event containing user registration data
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) sendUserRegistrationEmail(event *entities.EmailEvent) error {
	// Extract username from event data for personalization
	username, ok := event.Data["username"].(string)
	if !ok {
		return fmt.Errorf("username not found in email event data")
	}

	// Use the original email service to send the actual email
	// For user registration, we'll send a welcome email
	return s.emailService.SendWelcomeEmail(event.To, username)
}

// sendCustomEmail sends a custom email with specified content.
// This method handles custom email sending with flexible content
// and subject lines.
//
// Business Rules:
//   - Logs email content for monitoring
//   - Supports custom subject and body
//   - Can be extended with third-party email services
//
// Parameters:
//   - event: Email event containing custom email data
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) sendCustomEmail(event *entities.EmailEvent) error {
	// Log custom email details for monitoring and debugging
	log.Printf("Sending custom email to %s: %s", event.To, event.Subject)
	log.Printf("Email body: %s", event.Body)

	// You can implement actual email sending here
	// For example, using a third-party email service like SendGrid, AWS SES, etc.
	// This is a placeholder for custom email implementation

	return nil
}

// sendTemplateEmail sends a template-based email with dynamic data.
// This method handles template-based emails with data interpolation.
//
// Business Rules:
//   - Template name must be specified
//   - Supports dynamic data interpolation
//   - Logs template usage for monitoring
//
// Parameters:
//   - event: Email event containing template email data
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) sendTemplateEmail(event *entities.EmailEvent) error {
	// Log template email details for monitoring and debugging
	log.Printf("Sending template email to %s: %s (Template: %s)", event.To, event.Subject, event.Template)

	// You can implement template-based email sending here
	// For example, using HTML templates with data interpolation
	// This is a placeholder for template email implementation

	return nil
}

// StartEmailConsumerWithRetry starts the email consumer with retry logic
// for improved reliability and fault tolerance.
//
// Business Rules:
//   - Attempts to start consumer up to maxRetries times
//   - Implements exponential backoff between retries
//   - Logs retry attempts for monitoring
//   - Returns error if all retries fail
//
// Parameters:
//   - ctx: Context for the operation
//   - maxRetries: Maximum number of retry attempts
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) StartEmailConsumerWithRetry(ctx context.Context, maxRetries int) error {
	var lastErr error

	// Attempt to start email consumer with retry logic
	for i := 0; i < maxRetries; i++ {
		if err := s.StartEmailConsumer(ctx); err != nil {
			lastErr = err
			log.Printf("Failed to start email consumer (attempt %d/%d): %v", i+1, maxRetries, err)

			// Wait before retrying with exponential backoff
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		} else {
			log.Println("Email consumer started successfully")
			return nil
		}
	}

	return fmt.Errorf("failed to start email consumer after %d attempts: %w", maxRetries, lastErr)
}

// GetEmailQueueInfo retrieves information about the email queue status.
// This method provides monitoring capabilities for the email processing system
// including queue length, processing status, and performance metrics.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - map[string]interface{}: Queue information including metrics and status
//   - error: Any error that occurred during the operation
func (s *EmailConsumerService) GetEmailQueueInfo(ctx context.Context) (map[string]interface{}, error) {
	return s.rabbitMQService.GetQueueInfo(ctx, "email.sending")
}

// HealthCheck performs a health check on the email consumer service
// and RabbitMQ connection. This method verifies that the email consumer
// is operational and can process email requests effectively.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error that occurred during the health check
func (s *EmailConsumerService) HealthCheck(ctx context.Context) error {
	return s.rabbitMQService.HealthCheck(ctx)
}
