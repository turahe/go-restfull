package services

// EmailService defines the interface for email-related operations
type EmailService interface {
	// SendEmail sends an email
	SendEmail(to, subject, body string) error
	
	// SendWelcomeEmail sends a welcome email to new users
	SendWelcomeEmail(to, username string) error
	
	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(to, resetToken string) error
	
	// ValidateEmail validates email format
	ValidateEmail(email string) error
} 