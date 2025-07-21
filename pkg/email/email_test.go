package email

import (
	"testing"
	"webapi/config"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...zapcore.Field) {}

func TestSendEmail_LogsEmail(t *testing.T) {
	// Save and restore the original logger
	origLogger := config.GetConfig().Log
	cfg := config.GetConfig()
	cfg.Email = config.Email{
		FromAddress: "from@example.com",
		FromName:    "Tester",
		SMTPHost:    "smtp.example.com",
	}

	e := &EmailService{config: &cfg.Email}

	// This just checks that SendEmail returns nil (since it only logs)
	err := e.SendEmail("to@example.com", "Subject", "Body")
	assert.NoError(t, err)
}
