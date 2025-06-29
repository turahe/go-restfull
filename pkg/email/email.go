package email

import (
	"webapi/config"
	"webapi/internal/logger"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// EmailService provides email sending functionality
type EmailService struct {
	config *config.Email
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	cfg := config.GetConfig()
	return &EmailService{config: &cfg.Email}
}

// SendEmail sends an email using gomail
func (e *EmailService) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(e.config.FromAddress, e.config.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// For development/testing, just log the email
	logger.Log.Info("Email sent",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("body", body),
		zap.String("smtp_host", e.config.SMTPHost))

	// Uncomment the lines below to actually send emails
	// d := gomail.NewDialer(e.config.SMTPHost, e.config.SMTPPort, e.config.Username, e.config.Password)
	// if err := d.DialAndSend(m); err != nil {
	//     return fmt.Errorf("failed to send email: %w", err)
	// }

	return nil
}
