package email

import (
	"bytes"
	"fmt"
	"text/template"
	"webapi/config"

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

	// Uncomment the lines below to actually send emails
	d := gomail.NewDialer(e.config.SMTPHost, e.config.SMTPPort, e.config.Username, e.config.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendEmailTemplate sends an email using a Go template file and data
func (e *EmailService) SendEmailTemplate(to, subject, templatePath string, data interface{}, isHTML bool) error {
	var body bytes.Buffer
	if isHTML {
		tmpl, err := template.New("email").ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("failed to parse HTML template: %w", err)
		}
		err = tmpl.Execute(&body, data)
		if err != nil {
			return fmt.Errorf("failed to execute HTML template: %w", err)
		}
	} else {
		tmpl, err := template.New("email").ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("failed to parse text template: %w", err)
		}
		err = tmpl.Execute(&body, data)
		if err != nil {
			return fmt.Errorf("failed to execute text template: %w", err)
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.config.FromAddress, e.config.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	if isHTML {
		m.SetBody("text/html", body.String())
	} else {
		m.SetBody("text/plain", body.String())
	}

	d := gomail.NewDialer(e.config.SMTPHost, e.config.SMTPPort, e.config.Username, e.config.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
