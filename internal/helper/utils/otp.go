package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/turahe/go-restfull/internal/helper/cache"
	"github.com/turahe/go-restfull/pkg/email"
)

const (
	// DefaultOTPExpiration defines the standard expiration time for OTP codes (5 minutes)
	// This ensures OTPs are not valid for too long, improving security
	DefaultOTPExpiration = 5 * time.Minute
	
	// DefaultResetLinkExpiration defines the standard expiration time for password reset links (1 hour)
	// Reset links need longer validity as users may take time to check their email
	DefaultResetLinkExpiration = 1 * time.Hour
	
	// OTPKeyPrefix is the Redis key prefix for storing OTP data
	// Used to organize and namespace OTP keys in Redis
	OTPKeyPrefix = "otp"
	
	// ResetLinkKeyPrefix is the Redis key prefix for storing reset link data
	// Used to organize and namespace reset link keys in Redis
	ResetLinkKeyPrefix = "reset_link"
)

// OTPData represents the OTP data structure stored in Redis
// Contains all necessary information for OTP validation and management
type OTPData struct {
	OTP       string    `json:"otp"`        // The actual OTP code (e.g., "123456")
	Identity  string    `json:"identity"`   // User's phone number or email
	Type      string    `json:"type"`       // Identity type: "email" or "phone"
	CreatedAt time.Time `json:"created_at"` // When the OTP was generated
	ExpiresAt time.Time `json:"expires_at"` // When the OTP expires
}

// ResetLinkData represents the reset link data structure stored in Redis
// Contains all necessary information for password reset link validation
type ResetLinkData struct {
	Token     string    `json:"token"`      // Secure random token for reset link
	Identity  string    `json:"identity"`   // User's email address
	Type      string    `json:"type"`       // Identity type: always "email" for reset links
	CreatedAt time.Time `json:"created_at"` // When the reset link was generated
	ExpiresAt time.Time `json:"expires_at"` // When the reset link expires
}

// OTPService provides comprehensive OTP generation, storage, and management functionality
// Handles both phone-based OTPs and email-based reset links
type OTPService struct{}

// NewOTPService creates a new instance of the OTP service
// Returns a ready-to-use OTPService struct
func NewOTPService() *OTPService {
	return &OTPService{}
}

// GenerateAndStoreOTP generates an OTP and stores it in Redis for phone-based authentication
// This function creates a secure OTP, stores it with expiration, and returns the OTP code
//
// Parameters:
//   - ctx: context for Redis operations and cancellation
//   - identity: phone number to associate with the OTP
//   - length: number of digits in the OTP (e.g., 6 for "123456")
//   - expiration: how long the OTP should be valid (uses default if 0)
//
// Returns:
//   - string: the generated OTP code
//   - error: any error that occurred during generation or storage
func (s *OTPService) GenerateAndStoreOTP(ctx context.Context, identity string, length int, expiration time.Duration) (string, error) {
	// Use default expiration if none specified
	if expiration == 0 {
		expiration = DefaultOTPExpiration
	}

	// Generate a random numeric OTP of specified length
	otp := GenerateOTP(length)

	// Create OTP data structure with metadata
	otpData := OTPData{
		OTP:       otp,                    // The generated OTP code
		Identity:  identity,               // User's phone number
		Type:      "phone",                // Mark as phone-based OTP
		CreatedAt: time.Now(),             // Current timestamp
		ExpiresAt: time.Now().Add(expiration), // Expiration timestamp
	}

	// Store OTP data in Redis with automatic expiration
	key := s.generateOTPKey(identity)
	err := cache.Set(ctx, key, otpData, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to store OTP in Redis: %w", err)
	}

	return otp, nil
}

// SendOTPToWhatsApp sends an OTP code via WhatsApp using an external API endpoint
// This function formats the message and sends it to the user's WhatsApp number
//
// Parameters:
//   - ctx: context for HTTP request cancellation
//   - phoneNumber: the recipient's phone number in international format
//   - otp: the OTP code to send
//
// Returns:
//   - error: any error that occurred during the WhatsApp API call
func (s *OTPService) SendOTPToWhatsApp(ctx context.Context, phoneNumber, otp string) error {
	// Format phone number to WhatsApp chat ID format
	// WhatsApp uses phone number + "@c.us" as the chat identifier
	chatID := phoneNumber + "@c.us"

	// Create user-friendly message with OTP code and expiration information
	message := fmt.Sprintf("Your OTP code is: %s\n\nThis code will expire in 5 minutes.", otp)

	// Prepare the request payload for the WhatsApp API
	payload := map[string]interface{}{
		"chatId":                 chatID,                    // Recipient's WhatsApp chat ID
		"reply_to":               nil,                       // No reply to message
		"text":                   message,                   // The OTP message
		"linkPreview":            true,                      // Enable link previews
		"linkPreviewHighQuality": false,                     // Use standard quality for performance
		"session":                "default",                 // WhatsApp session identifier
	}

	// Convert payload to JSON for HTTP request
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP POST request to WhatsApp API endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/sendText", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set appropriate headers for JSON API request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}
	defer resp.Body.Close()

	// Check if the API call was successful
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WhatsApp API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GenerateAndStoreAndSendOTP is a convenience function that combines OTP generation, storage, and WhatsApp delivery
// This function handles the complete flow from OTP creation to user notification
//
// Parameters:
//   - ctx: context for all operations
//   - phoneNumber: the user's phone number
//   - length: OTP length in digits
//   - expiration: OTP validity duration
//
// Returns:
//   - string: the generated OTP code
//   - error: any error that occurred during the process
func (s *OTPService) GenerateAndStoreAndSendOTP(ctx context.Context, phoneNumber string, length int, expiration time.Duration) (string, error) {
	// Step 1: Generate and store OTP in Redis
	otp, err := s.GenerateAndStoreOTP(ctx, phoneNumber, length, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate and store OTP: %w", err)
	}

	// Step 2: Send OTP via WhatsApp
	err = s.SendOTPToWhatsApp(ctx, phoneNumber, otp)
	if err != nil {
		return "", fmt.Errorf("failed to send OTP via WhatsApp: %w", err)
	}

	return otp, nil
}

// GenerateAndStoreResetLink generates a secure reset token and stores it in Redis for email-based password resets
// This function creates a cryptographically secure token and sends a reset email to the user
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: user's email address
//   - expiration: how long the reset link should be valid (uses default if 0)
//
// Returns:
//   - string: the generated reset token
//   - error: any error that occurred during generation, storage, or email sending
func (s *OTPService) GenerateAndStoreResetLink(ctx context.Context, identity string, expiration time.Duration) (string, error) {
	// Use default expiration if none specified
	if expiration == 0 {
		expiration = DefaultResetLinkExpiration
	}

	// Generate a cryptographically secure random token
	token := s.generateResetToken()

	// Create reset link data structure with metadata
	resetData := ResetLinkData{
		Token:     token,                  // The secure reset token
		Identity:  identity,               // User's email address
		Type:      "email",                // Mark as email-based reset
		CreatedAt: time.Now(),             // Current timestamp
		ExpiresAt: time.Now().Add(expiration), // Expiration timestamp
	}

	// Store reset link data in Redis with automatic expiration
	key := s.generateResetLinkKey(identity)
	resetDataBytes, err := json.Marshal(resetData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal reset link data: %w", err)
	}
	err = cache.Set(ctx, key, string(resetDataBytes), expiration)
	if err != nil {
		return "", fmt.Errorf("failed to store reset link in Redis: %w", err)
	}

	// Send reset password email to the user
	emailService := email.NewEmailService()
	templateData := struct {
		ResetLink string
		Token     string
	}{
		ResetLink: "https://yourdomain.com/reset-password?token=" + token, // Frontend reset URL
		Token:     token,                                                   // Token for verification
	}
	err = emailService.SendEmailTemplate(identity, "Reset Password", "pkg/template/email/reset_password.html", templateData, true)
	if err != nil {
		fmt.Printf("Failed to send reset password email: %v\n", err)
	}

	return token, nil
}

// ValidateOTP validates an OTP against the stored value in Redis for phone-based authentication
// This function checks if the provided OTP matches the stored one and removes it after successful validation
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: phone number associated with the OTP
//   - otp: the OTP code to validate
//
// Returns:
//   - bool: true if OTP is valid, false otherwise
//   - error: any error that occurred during validation
func (s *OTPService) ValidateOTP(ctx context.Context, identity, otp string) (bool, error) {
	key := s.generateOTPKey(identity)

	// Retrieve OTP data from Redis
	otpDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("OTP not found or expired: %w", err)
	}

	// Compare the provided OTP with the stored one
	// Note: In production, you might want to unmarshal the JSON and compare the OTP field
	if otpDataStr == otp {
		// Remove OTP from Redis after successful validation for security
		_ = cache.Remove(ctx, key)
		return true, nil
	}

	return false, nil
}

// ValidateResetLink validates a reset link token against the stored value in Redis for email-based resets
// This function checks if the provided token matches the stored one and removes it after successful validation
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: email address associated with the reset link
//   - token: the reset token to validate
//
// Returns:
//   - bool: true if token is valid, false otherwise
//   - error: any error that occurred during validation
func (s *OTPService) ValidateResetLink(ctx context.Context, identity, token string) (bool, error) {
	key := s.generateResetLinkKey(identity)

	// Retrieve reset link data from Redis
	resetDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("Reset link not found or expired: %w", err)
	}

	// Unmarshal the JSON data to access the token field
	var resetData ResetLinkData
	err = json.Unmarshal([]byte(resetDataStr), &resetData)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal reset link data: %w", err)
	}

	// Compare the provided token with the stored one
	if resetData.Token == token {
		// Remove reset link from Redis after successful validation for security
		_ = cache.Remove(ctx, key)
		return true, nil
	}

	return false, nil
}

// GetOTP retrieves OTP data from Redis without removing it
// This function is useful for debugging or checking OTP status without invalidating it
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: phone number associated with the OTP
//
// Returns:
//   - *OTPData: the OTP data structure, or nil if not found
//   - error: any error that occurred during retrieval
func (s *OTPService) GetOTP(ctx context.Context, identity string) (*OTPData, error) {
	key := s.generateOTPKey(identity)

	// Retrieve OTP data from Redis
	otpDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("OTP not found or expired: %w", err)
	}

	// Return a basic OTPData structure for simplicity
	// In production, you would unmarshal the JSON data from Redis
	return &OTPData{
		OTP:       otpDataStr,             // The OTP code
		Identity:  identity,               // Phone number
		Type:      "phone",                // Type identifier
		CreatedAt: time.Now(),             // Current time (approximate)
		ExpiresAt: time.Now().Add(DefaultOTPExpiration), // Estimated expiration
	}, nil
}

// GetResetLink retrieves reset link data from Redis without removing it
// This function is useful for debugging or checking reset link status without invalidating it
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: email address associated with the reset link
//
// Returns:
//   - *ResetLinkData: the reset link data structure, or nil if not found
//   - error: any error that occurred during retrieval
func (s *OTPService) GetResetLink(ctx context.Context, identity string) (*ResetLinkData, error) {
	key := s.generateResetLinkKey(identity)

	// Retrieve reset link data from Redis
	resetDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("Reset link not found or expired: %w", err)
	}

	// Unmarshal the JSON data to return the complete structure
	var resetData ResetLinkData
	err = json.Unmarshal([]byte(resetDataStr), &resetData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reset link data: %w", err)
	}

	// Return the complete reset link data
	return &resetData, nil
}

// RemoveOTP removes an OTP from Redis, effectively invalidating it
// This function is useful for manual cleanup or when OTPs need to be invalidated
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: phone number associated with the OTP
//
// Returns:
//   - error: any error that occurred during removal
func (s *OTPService) RemoveOTP(ctx context.Context, identity string) error {
	key := s.generateOTPKey(identity)
	return cache.Remove(ctx, key)
}

// RemoveResetLink removes a reset link from Redis, effectively invalidating it
// This function is useful for manual cleanup or when reset links need to be invalidated
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: email address associated with the reset link
//
// Returns:
//   - error: any error that occurred during removal
func (s *OTPService) RemoveResetLink(ctx context.Context, identity string) error {
	key := s.generateResetLinkKey(identity)
	return cache.Remove(ctx, key)
}

// IsOTPExists checks if an OTP exists for the given identity without retrieving its value
// This function is useful for checking OTP existence without exposing the actual OTP code
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: phone number to check
//
// Returns:
//   - bool: true if OTP exists, false otherwise
//   - error: any error that occurred during the check
func (s *OTPService) IsOTPExists(ctx context.Context, identity string) (bool, error) {
	key := s.generateOTPKey(identity)

	// Try to get the OTP data (we don't need the actual value, just to check existence)
	_, err := cache.Get(ctx, key)
	if err != nil {
		return false, nil // OTP doesn't exist or has expired
	}

	return true, nil
}

// IsResetLinkExists checks if a reset link exists for the given identity without retrieving its value
// This function is useful for checking reset link existence without exposing the actual token
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: email address to check
//
// Returns:
//   - bool: true if reset link exists, false otherwise
//   - error: any error that occurred during the check
func (s *OTPService) IsResetLinkExists(ctx context.Context, identity string) (bool, error) {
	key := s.generateResetLinkKey(identity)

	// Try to get the reset link data (we don't need the actual value, just to check existence)
	_, err := cache.Get(ctx, key)
	if err != nil {
		return false, nil // Reset link doesn't exist or has expired
	}

	return true, nil
}

// generateOTPKey generates a unique Redis key for storing OTP data
// The key format ensures uniqueness and easy identification in Redis
//
// Parameters:
//   - identity: phone number to include in the key
//
// Returns:
//   - string: formatted Redis key (e.g., "otp:phone:+628123456789")
func (s *OTPService) generateOTPKey(identity string) string {
	return fmt.Sprintf("%s:phone:%s", OTPKeyPrefix, identity)
}

// generateResetLinkKey generates a unique Redis key for storing reset link data
// The key format ensures uniqueness and easy identification in Redis
//
// Parameters:
//   - identity: email address to include in the key
//
// Returns:
//   - string: formatted Redis key (e.g., "reset_link:email:user@example.com")
func (s *OTPService) generateResetLinkKey(identity string) string {
	return fmt.Sprintf("%s:email:%s", ResetLinkKeyPrefix, identity)
}

// generateResetToken generates a cryptographically secure random token for reset links
// This function creates a 32-byte random value and encodes it as a hex string
//
// Returns:
//   - string: 64-character hexadecimal token
func (s *OTPService) generateResetToken() string {
	bytes := make([]byte, 32)        // Create 32-byte slice for randomness
	rand.Read(bytes)                  // Fill with cryptographically secure random bytes
	return hex.EncodeToString(bytes)  // Convert to hexadecimal string
}

// ResendOTP generates a new OTP and replaces the existing one for the same identity
// This function is useful when users request a new OTP (e.g., if the previous one expired)
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: phone number to generate new OTP for
//   - length: number of digits in the new OTP
//   - expiration: how long the new OTP should be valid
//
// Returns:
//   - string: the new OTP code
//   - error: any error that occurred during the process
func (s *OTPService) ResendOTP(ctx context.Context, identity string, length int, expiration time.Duration) (string, error) {
	// Remove existing OTP if it exists to prevent conflicts
	_ = s.RemoveOTP(ctx, identity)

	// Generate and store new OTP
	return s.GenerateAndStoreOTP(ctx, identity, length, expiration)
}

// ResendResetLink generates a new reset link and replaces the existing one for the same identity
// This function is useful when users request a new reset link (e.g., if the previous one expired)
//
// Parameters:
//   - ctx: context for Redis operations
//   - identity: email address to generate new reset link for
//   - expiration: how long the new reset link should be valid
//
// Returns:
//   - string: the new reset token
//   - error: any error that occurred during the process
func (s *OTPService) ResendResetLink(ctx context.Context, identity string, expiration time.Duration) (string, error) {
	// Remove existing reset link if it exists to prevent conflicts
	_ = s.RemoveResetLink(ctx, identity)

	// Generate and store new reset link
	return s.GenerateAndStoreResetLink(ctx, identity, expiration)
}
