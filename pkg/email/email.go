package email

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"text/template"
	"github.com/turahe/go-restfull/config"

	"gopkg.in/gomail.v2"
)

// EmailService provides email sending functionality
type Email struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	FromAddress string
	FromName    string
}

// NewEmailService creates a new email service instance
func NewEmailService() *Email {
	cfg := config.GetConfig()
	return &Email{
		SMTPHost:    cfg.Email.SMTPHost,
		SMTPPort:    cfg.Email.SMTPPort,
		Username:    cfg.Email.Username,
		Password:    cfg.Email.Password,
		FromAddress: cfg.Email.FromAddress,
		FromName:    cfg.Email.FromName,
	}
}

// SendEmail sends an email using gomail
func (e *Email) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(e.FromAddress, e.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// Uncomment the lines below to actually send emails
	d := gomail.NewDialer(e.SMTPHost, e.SMTPPort, e.Username, e.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendEmailTemplate sends an email using a Go template file and data
func (e *Email) SendEmailTemplate(to, subject, templatePath string, data interface{}, isHTML bool) error {
	var body bytes.Buffer
	if isHTML {
		tmpl, err := htmltemplate.ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("failed to parse HTML template: %w", err)
		}
		err = tmpl.Execute(&body, data)
		if err != nil {
			return fmt.Errorf("failed to execute HTML template: %w", err)
		}
	} else {
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("failed to parse text template: %w", err)
		}
		err = tmpl.Execute(&body, data)
		if err != nil {
			return fmt.Errorf("failed to execute text template: %w", err)
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.FromAddress, e.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	if isHTML {
		m.SetBody("text/html", body.String())
	} else {
		m.SetBody("text/plain", body.String())
	}

	d := gomail.NewDialer(e.SMTPHost, e.SMTPPort, e.Username, e.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
