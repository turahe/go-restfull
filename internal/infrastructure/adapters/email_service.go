package adapters

import (
	"fmt"
	"regexp"

	"webapi/internal/domain/services"
	"webapi/pkg/email"
)

// smtpEmailService implements EmailService interface
type smtpEmailService struct {
	emailClient *email.Email
}

// NewSmtpEmailService creates a new SMTP email service
func NewSmtpEmailService(emailClient *email.Email) services.EmailService {
	return &smtpEmailService{
		emailClient: emailClient,
	}
}

func (s *smtpEmailService) SendEmail(to, subject, body string) error {
	return s.emailClient.SendEmail(to, subject, body)
}

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

func (s *smtpEmailService) ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
} 