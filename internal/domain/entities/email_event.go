// Package entities provides the core domain models and business logic entities
// for the application. This file contains the EmailEvent entity for managing
// email events in the messaging system, particularly for RabbitMQ integration.
package entities

import (
	"time"
)

// EmailEvent represents an email event for RabbitMQ messaging system.
// This entity encapsulates all the necessary information needed to process
// and send emails through the messaging infrastructure.
//
// The entity supports:
// - Multiple email types (welcome, password reset, registration, etc.)
// - Template-based email generation
// - Custom data injection for dynamic content
// - Priority-based email processing
// - Header customization for advanced email features
type EmailEvent struct {
	Type      string                 `json:"type"`               // Type of email event (welcome, password_reset, etc.)
	To        string                 `json:"to"`                 // Recipient email address
	Subject   string                 `json:"subject"`            // Email subject line
	Body      string                 `json:"body"`               // Email body content (if not using template)
	Template  string                 `json:"template,omitempty"` // Template name for email generation
	Data      map[string]interface{} `json:"data,omitempty"`     // Dynamic data for template rendering
	Headers   map[string]string      `json:"headers,omitempty"`  // Custom email headers
	Timestamp time.Time              `json:"timestamp"`          // When the event was created
	Priority  string                 `json:"priority,omitempty"` // Email priority (normal, high, urgent)
}

// NewWelcomeEmailEvent creates a new welcome email event for new users.
// This constructor sets up a standard welcome email with appropriate
// subject, template, and user data for personalization.
//
// Parameters:
//   - email: Recipient email address
//   - username: Username for personalization
//
// Returns:
//   - *EmailEvent: Pointer to the newly created welcome email event
//
// Note: Uses "normal" priority and "welcome" template
func NewWelcomeEmailEvent(email, username string) *EmailEvent {
	return &EmailEvent{
		Type:      "welcome",                                    // Set email type as welcome
		To:        email,                                        // Set recipient email
		Subject:   "Welcome to Our Platform!",                   // Set welcome subject
		Template:  "welcome",                                    // Use welcome template
		Data:      map[string]interface{}{"username": username}, // Include username for personalization
		Timestamp: time.Now(),                                   // Set current timestamp
		Priority:  "normal",                                     // Set normal priority
	}
}

// NewPasswordResetEmailEvent creates a new password reset email event.
// This constructor sets up a high-priority email for password reset
// requests with OTP verification code.
//
// Parameters:
//   - email: Recipient email address
//   - otp: One-time password for verification
//
// Returns:
//   - *EmailEvent: Pointer to the newly created password reset email event
//
// Note: Uses "high" priority and "reset_password" template
func NewPasswordResetEmailEvent(email, otp string) *EmailEvent {
	return &EmailEvent{
		Type:      "password_reset",                   // Set email type as password reset
		To:        email,                              // Set recipient email
		Subject:   "Password Reset Request",           // Set password reset subject
		Template:  "reset_password",                   // Use reset password template
		Data:      map[string]interface{}{"otp": otp}, // Include OTP for verification
		Timestamp: time.Now(),                         // Set current timestamp
		Priority:  "high",                             // Set high priority for security
	}
}

// NewUserRegistrationEvent creates a new user registration email event.
// This constructor sets up a confirmation email for successful user
// registration with account details.
//
// Parameters:
//   - userID: Unique identifier for the registered user
//   - username: Username for the new account
//   - email: Email address for the new account
//
// Returns:
//   - *EmailEvent: Pointer to the newly created user registration email event
//
// Note: Uses "normal" priority and "registration" template
func NewUserRegistrationEvent(userID, username, email string) *EmailEvent {
	return &EmailEvent{
		Type:     "user_registration",               // Set email type as user registration
		To:       email,                             // Set recipient email
		Subject:  "Account Registration Successful", // Set registration success subject
		Template: "registration",                    // Use registration template
		Data: map[string]interface{}{ // Include account details for confirmation
			"user_id":  userID,
			"username": username,
			"email":    email,
		},
		Timestamp: time.Now(), // Set current timestamp
		Priority:  "normal",   // Set normal priority
	}
}
