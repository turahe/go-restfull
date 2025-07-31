// Package services provides application-level business logic for email management.
// This package contains the email service implementation that handles email sending,
// template processing, and asynchronous email delivery via RabbitMQ while ensuring
// reliable communication with users.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/pkg/rabbitmq"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// EmailService handles email operations using RabbitMQ for asynchronous processing.
// This service provides a reliable and scalable email delivery system that supports
// various email types including welcome emails, password resets, custom emails,
// and template-based emails.
type EmailService struct {
	rabbitMQService *rabbitmq.Service
}

// NewEmailService creates a new email service instance with RabbitMQ integration.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the messaging infrastructure.
//
// Parameters:
//   - rabbitMQService: RabbitMQ service for asynchronous email processing
//
// Returns:
//   - *EmailService: The email service instance
func NewEmailService(rabbitMQService *rabbitmq.Service) *EmailService {
	return &EmailService{
		rabbitMQService: rabbitMQService,
	}
}

// SendWelcomeEmail sends a welcome email to new users via RabbitMQ.
// This method creates a welcome email event and publishes it to the email queue
// for asynchronous processing, ensuring non-blocking user registration.
//
// Business Rules:
//   - Email address must be valid and provided
//   - Username is used for personalization
//   - Email is sent asynchronously to prevent registration delays
//   - Uses standardized welcome email template
//
// Parameters:
//   - email: Email address of the new user
//   - username: Username for personalization
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailService) SendWelcomeEmail(email, username string) error {
	ctx := context.Background()

	// Create welcome email event with user information
	event := entities.NewWelcomeEmailEvent(email, username)

	// Publish to email queue for asynchronous processing
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendPasswordResetEmail sends a password reset email with OTP via RabbitMQ.
// This method creates a password reset email event and publishes it to the email
// queue for secure and asynchronous delivery.
//
// Security Features:
//   - OTP is included in the email for secure password reset
//   - Asynchronous processing prevents timing attacks
//   - Uses standardized password reset template
//
// Business Rules:
//   - Email address must be valid and provided
//   - OTP must be generated and provided
//   - Email is sent asynchronously for security
//
// Parameters:
//   - email: Email address of the user requesting password reset
//   - otp: One-time password for secure password reset
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailService) SendPasswordResetEmail(email, otp string) error {
	ctx := context.Background()

	// Create password reset email event with OTP
	event := entities.NewPasswordResetEmailEvent(email, otp)

	// Publish to email queue for asynchronous processing
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendUserRegistrationEmail sends a user registration confirmation email via RabbitMQ.
// This method creates a user registration event and publishes it to the email queue
// for notification purposes.
//
// Business Rules:
//   - User ID, username, and email must be provided
//   - Email is sent asynchronously to prevent registration delays
//   - Uses standardized registration confirmation template
//
// Parameters:
//   - userID: Unique identifier of the registered user
//   - username: Username of the registered user
//   - email: Email address of the registered user
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailService) SendUserRegistrationEmail(userID, username, email string) error {
	ctx := context.Background()

	// Create user registration email event with user details
	event := entities.NewUserRegistrationEvent(userID, username, email)

	// Publish to email queue for asynchronous processing
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendCustomEmail sends a custom email with specified content via RabbitMQ.
// This method allows for flexible email sending with custom subject, body,
// and optional headers for advanced email configurations.
//
// Business Rules:
//   - Recipient email address must be provided
//   - Subject and body must be specified
//   - Headers are optional but can be used for advanced configurations
//   - Email is sent asynchronously for non-blocking operation
//
// Parameters:
//   - to: Recipient email address
//   - subject: Email subject line
//   - body: Email body content
//   - headers: Optional email headers for advanced configurations
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailService) SendCustomEmail(to, subject, body string, headers map[string]string) error {
	ctx := context.Background()

	// Create custom email event with provided content
	event := &entities.EmailEvent{
		Type:      "custom",
		To:        to,
		Subject:   subject,
		Body:      body,
		Headers:   headers,
		Timestamp: time.Now(),
		Priority:  "normal",
	}

	// Publish to email queue for asynchronous processing
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendEmailWithTemplate sends an email using a template with dynamic data via RabbitMQ.
// This method supports template-based emails with dynamic content rendering
// for consistent and professional email communications.
//
// Business Rules:
//   - Recipient email address must be provided
//   - Template name must be specified
//   - Data map is used for template variable substitution
//   - Email is sent asynchronously for non-blocking operation
//
// Parameters:
//   - to: Recipient email address
//   - subject: Email subject line
//   - template: Template name to use for email generation
//   - data: Dynamic data for template variable substitution
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailService) SendEmailWithTemplate(to, subject, template string, data map[string]interface{}) error {
	ctx := context.Background()

	// Create template-based email event with dynamic data
	event := &entities.EmailEvent{
		Type:      "template",
		To:        to,
		Subject:   subject,
		Template:  template,
		Data:      data,
		Timestamp: time.Now(),
		Priority:  "normal",
	}

	// Publish to email queue for asynchronous processing
	return s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", event)
}

// SendBulkEmail sends multiple emails via RabbitMQ in a batch operation.
// This method processes multiple email events and publishes them to the queue
// for efficient bulk email delivery.
//
// Business Rules:
//   - All emails in the batch must be valid
//   - Each email is processed individually for error tracking
//   - Failed emails are reported with specific error details
//   - Emails are sent asynchronously for non-blocking operation
//
// Parameters:
//   - emails: Slice of email events to send
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *EmailService) SendBulkEmail(emails []entities.EmailEvent) error {
	ctx := context.Background()

	// Process each email in the batch
	for _, email := range emails {
		if err := s.rabbitMQService.PublishToQueueJSON(ctx, "email.sending", email); err != nil {
			return fmt.Errorf("failed to publish email to %s: %w", email.To, err)
		}
	}

	return nil
}

// GetEmailQueueInfo retrieves information about the email queue status.
// This method provides monitoring capabilities for the email delivery system
// including queue length, processing status, and performance metrics.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - map[string]interface{}: Queue information including metrics and status
//   - error: Any error that occurred during the operation
func (s *EmailService) GetEmailQueueInfo(ctx context.Context) (map[string]interface{}, error) {
	return s.rabbitMQService.GetQueueInfo(ctx, "email.sending")
}

// HealthCheck performs a health check on the email service and RabbitMQ connection.
// This method verifies that the email service is operational and can process
// email requests effectively.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error that occurred during the health check
func (s *EmailService) HealthCheck(ctx context.Context) error {
	return s.rabbitMQService.HealthCheck(ctx)
}
