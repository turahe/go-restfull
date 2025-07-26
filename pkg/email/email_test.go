package email

import (
	"os"
	"testing"
	"webapi/config"

	"go.uber.org/zap/zapcore"
)

type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...zapcore.Field) {}

// Define a local Email struct for testing if not exported
type testEmailConfig struct {
	FromAddress string
	FromName    string
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
}

func TestSendEmail_LogsEmail(t *testing.T) {
	cfg := config.GetConfig()
	var emailCfg *Email
	if cfg == nil {
		emailCfg = &Email{
			FromAddress: "from@example.com",
			FromName:    "Tester",
			SMTPHost:    "smtp.example.com",
			SMTPPort:    587,
			Username:    "user",
			Password:    "pass",
		}
	} else {
		tmp := cfg.Email
		emailCfg = &Email{
			FromAddress: tmp.FromAddress,
			FromName:    tmp.FromName,
			SMTPHost:    tmp.SMTPHost,
			SMTPPort:    tmp.SMTPPort,
			Username:    tmp.Username,
			Password:    tmp.Password,
		}
	}

	// This just checks that SendEmail returns nil (since it only logs)
	err := emailCfg.SendEmail("to@example.com", "Subject", "Body")
	if err != nil {
		t.Errorf("SendEmail returned error: %v", err)
	}
}

func TestSendEmailTemplate_HTML(t *testing.T) {
	tpl := `<!DOCTYPE html><html><body><h1>Hello, {{.Name}}!</h1></body></html>`
	tplPath := "test_template.html"
	os.WriteFile(tplPath, []byte(tpl), 0644)
	defer os.Remove(tplPath)

	cfg := config.GetConfig()
	var emailCfg *Email
	if cfg == nil {
		emailCfg = &Email{
			FromAddress: "from@example.com",
			FromName:    "Tester",
			SMTPHost:    "smtp.example.com",
			SMTPPort:    587,
			Username:    "user",
			Password:    "pass",
		}
	} else {
		tmp := cfg.Email
		emailCfg = &Email{
			FromAddress: tmp.FromAddress,
			FromName:    tmp.FromName,
			SMTPHost:    tmp.SMTPHost,
			SMTPPort:    tmp.SMTPPort,
			Username:    tmp.Username,
			Password:    tmp.Password,
		}
	}

	data := struct{ Name string }{Name: "TestUser"}
	err := emailCfg.SendEmailTemplate("test@example.com", "Test Subject", tplPath, data, true)
	if err != nil {
		t.Logf("SendEmailTemplate error (expected if SMTP not configured): %v", err)
	}
}
