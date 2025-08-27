// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"fmt"
	"regexp"

	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/pkg/email"
)

// smtpEmailService implements the EmailService interface using SMTP for email delivery.
// This service provides email functionality including sending general emails, welcome emails,
// password reset emails, and email validation. It wraps the underlying email client
// to provide a clean domain interface.
type smtpEmailService struct {
	// emailClient holds the underlying email client for SMTP operations
	emailClient *email.Email
}

// NewSmtpEmailService creates a new SMTP email service instance.
// This factory function returns a concrete implementation of the EmailService interface
// that uses SMTP for email delivery.
//
// Parameters:
//   - emailClient: The email client instance to use for SMTP operations
//
// Returns:
//   - services.EmailService: A new email service instance
func NewSmtpEmailService(emailClient *email.Email) services.EmailService {
	return &smtpEmailService{
		emailClient: emailClient,
	}
}

// SendEmail sends a general email with the specified recipient, subject, and body.
// This method delegates to the underlying email client to perform the actual SMTP operation.
// The body can contain HTML content for rich email formatting.
//
// Parameters:
//   - to: The recipient's email address
//   - subject: The email subject line
//   - body: The email body content (can include HTML)
//
// Returns:
//   - error: Any error that occurred during email sending
func (s *smtpEmailService) SendEmail(to, subject, body string) error {
	return s.emailClient.SendEmail(to, subject, body)
}

// SendWelcomeEmail sends a personalized welcome email to new users.
// This method creates a formatted HTML welcome message with the user's username
// and sends it using the underlying email client. The email includes a welcome
// message and contact information for support.
//
// Parameters:
//   - to: The recipient's email address
//   - username: The username to personalize the welcome message
//
// Returns:
//   - error: Any error that occurred during email sending
func (s *smtpEmailService) SendWelcomeEmail(to, username string) error {
	subject := "Welcome to Our Platform!"
	body := fmt.Sprintf(`
		<h1>Welcome %s!</h1>
		<p>Thank you for joining our platform. We're excited to have you on board!</p>
		<p>If you have any questions, feel free to reach out to our support team.</p>
		<br>
		<p>Best regards,<br>The Team</p>
	`, username)

	return s.emailClient.SendEmail(to, subject, body)
}

// SendPasswordResetEmail sends a password reset email with a reset token.
// This method creates a formatted HTML password reset email that includes
// a clickable link for the user to reset their password. The email also
// includes security information about the reset process.
//
// Parameters:
//   - to: The recipient's email address
//   - resetToken: The unique token for password reset verification
//
// Returns:
//   - error: Any error that occurred during email sending
func (s *smtpEmailService) SendPasswordResetEmail(to, resetToken string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<h1>Password Reset Request</h1>
		<p>You have requested to reset your password.</p>
		<p>Click the link below to reset your password:</p>
		<a href="https://yourapp.com/reset-password?token=%s">Reset Password</a>
		<p>If you didn't request this, please ignore this email.</p>
		<p>This link will expire in 1 hour.</p>
		<br>
		<p>Best regards,<br>The Team</p>
	`, resetToken)

	return s.emailClient.SendEmail(to, subject, body)
}

// ValidateEmail validates the format of an email address using regex.
// This method checks if the provided email string matches a standard email format
// including local part, @ symbol, domain, and TLD. It's useful for form validation
// and ensuring email addresses are properly formatted before sending.
//
// Parameters:
//   - email: The email address string to validate
//
// Returns:
//   - error: Validation error if the email format is invalid, nil if valid
func (s *smtpEmailService) ValidateEmail(email string) error {
	// Use regex to validate email format according to standard conventions
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}
