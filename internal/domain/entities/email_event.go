package entities

import (
	"time"
)

// EmailEvent represents an email event for RabbitMQ
type EmailEvent struct {
	Type      string                 `json:"type"`
	To        string                 `json:"to"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Template  string                 `json:"template,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Priority  string                 `json:"priority,omitempty"`
}

// NewWelcomeEmailEvent creates a new welcome email event
func NewWelcomeEmailEvent(email, username string) *EmailEvent {
	return &EmailEvent{
		Type:      "welcome",
		To:        email,
		Subject:   "Welcome to Our Platform!",
		Template:  "welcome",
		Data:      map[string]interface{}{"username": username},
		Timestamp: time.Now(),
		Priority:  "normal",
	}
}

// NewPasswordResetEmailEvent creates a new password reset email event
func NewPasswordResetEmailEvent(email, otp string) *EmailEvent {
	return &EmailEvent{
		Type:      "password_reset",
		To:        email,
		Subject:   "Password Reset Request",
		Template:  "reset_password",
		Data:      map[string]interface{}{"otp": otp},
		Timestamp: time.Now(),
		Priority:  "high",
	}
}

// NewUserRegistrationEvent creates a new user registration event
func NewUserRegistrationEvent(userID, username, email string) *EmailEvent {
	return &EmailEvent{
		Type:     "user_registration",
		To:       email,
		Subject:  "Account Registration Successful",
		Template: "registration",
		Data: map[string]interface{}{
			"user_id":  userID,
			"username": username,
			"email":    email,
		},
		Timestamp: time.Now(),
		Priority:  "normal",
	}
}
