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
	// Default OTP expiration time (5 minutes)
	DefaultOTPExpiration = 5 * time.Minute
	// Default reset link expiration time (1 hour)
	DefaultResetLinkExpiration = 1 * time.Hour
	// OTP key prefix for Redis
	OTPKeyPrefix = "otp"
	// Reset link key prefix for Redis
	ResetLinkKeyPrefix = "reset_link"
)

// OTPData represents the OTP data stored in Redis
type OTPData struct {
	OTP       string    `json:"otp"`
	Identity  string    `json:"identity"`
	Type      string    `json:"type"` // "email" or "phone"
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ResetLinkData represents the reset link data stored in Redis
type ResetLinkData struct {
	Token     string    `json:"token"`
	Identity  string    `json:"identity"`
	Type      string    `json:"type"` // "email"
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OTPService provides OTP generation and management functionality
type OTPService struct{}

// NewOTPService creates a new OTP service instance
func NewOTPService() *OTPService {
	return &OTPService{}
}

// GenerateAndStoreOTP generates an OTP and stores it in Redis (for phone)
func (s *OTPService) GenerateAndStoreOTP(ctx context.Context, identity string, length int, expiration time.Duration) (string, error) {
	if expiration == 0 {
		expiration = DefaultOTPExpiration
	}

	// Generate OTP
	otp := GenerateOTP(length)

	// Create OTP data
	otpData := OTPData{
		OTP:       otp,
		Identity:  identity,
		Type:      "phone",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiration),
	}

	// Store in Redis with expiration
	key := s.generateOTPKey(identity)
	err := cache.Set(ctx, key, otpData, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to store OTP in Redis: %w", err)
	}

	return otp, nil
}

// SendOTPToWhatsApp sends OTP via WhatsApp using the provided endpoint
func (s *OTPService) SendOTPToWhatsApp(ctx context.Context, phoneNumber, otp string) error {
	// Format phone number to WhatsApp format if needed
	// Assuming phoneNumber is already in international format (e.g., "6285225440150")
	chatID := phoneNumber + "@c.us"

	// Prepare the message with OTP
	message := fmt.Sprintf("Your OTP code is: %s\n\nThis code will expire in 5 minutes.", otp)

	// Prepare request payload
	payload := map[string]interface{}{
		"chatId":                 chatID,
		"reply_to":               nil,
		"text":                   message,
		"linkPreview":            true,
		"linkPreviewHighQuality": false,
		"session":                "default",
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/sendText", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WhatsApp API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GenerateAndStoreAndSendOTP generates OTP, stores it, and sends it via WhatsApp
func (s *OTPService) GenerateAndStoreAndSendOTP(ctx context.Context, phoneNumber string, length int, expiration time.Duration) (string, error) {
	// Generate and store OTP
	otp, err := s.GenerateAndStoreOTP(ctx, phoneNumber, length, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate and store OTP: %w", err)
	}

	// Send OTP via WhatsApp
	err = s.SendOTPToWhatsApp(ctx, phoneNumber, otp)
	if err != nil {
		return "", fmt.Errorf("failed to send OTP via WhatsApp: %w", err)
	}

	return otp, nil
}

// GenerateAndStoreResetLink generates a reset link token and stores it in Redis (for email)
func (s *OTPService) GenerateAndStoreResetLink(ctx context.Context, identity string, expiration time.Duration) (string, error) {
	if expiration == 0 {
		expiration = DefaultResetLinkExpiration
	}

	// Generate reset token
	token := s.generateResetToken()

	// Create reset link data
	resetData := ResetLinkData{
		Token:     token,
		Identity:  identity,
		Type:      "email",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiration),
	}

	// Store in Redis with expiration (marshal to JSON)
	key := s.generateResetLinkKey(identity)
	resetDataBytes, err := json.Marshal(resetData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal reset link data: %w", err)
	}
	err = cache.Set(ctx, key, string(resetDataBytes), expiration)
	if err != nil {
		return "", fmt.Errorf("failed to store reset link in Redis: %w", err)
	}

	// send email link to email
	emailService := email.NewEmailService()
	templateData := struct {
		ResetLink string
		Token     string
	}{
		ResetLink: "https://yourdomain.com/reset-password?token=" + token,
		Token:     token,
	}
	err = emailService.SendEmailTemplate(identity, "Reset Password", "pkg/template/email/reset_password.html", templateData, true)
	if err != nil {
		fmt.Printf("Failed to send reset password email: %v\n", err)
	}

	return token, nil
}

// ValidateOTP validates an OTP against the stored value in Redis (for phone)
func (s *OTPService) ValidateOTP(ctx context.Context, identity, otp string) (bool, error) {
	key := s.generateOTPKey(identity)

	// Get OTP data from Redis
	otpDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("OTP not found or expired: %w", err)
	}

	// Parse OTP data (for now, we'll just compare the OTP string directly)
	// In a real implementation, you might want to unmarshal the JSON
	if otpDataStr == otp {
		// Remove OTP from Redis after successful validation
		_ = cache.Remove(ctx, key)
		return true, nil
	}

	return false, nil
}

// ValidateResetLink validates a reset link token against the stored value in Redis (for email)
func (s *OTPService) ValidateResetLink(ctx context.Context, identity, token string) (bool, error) {
	key := s.generateResetLinkKey(identity)

	// Get reset link data from Redis
	resetDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("Reset link not found or expired: %w", err)
	}

	var resetData ResetLinkData
	err = json.Unmarshal([]byte(resetDataStr), &resetData)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal reset link data: %w", err)
	}

	if resetData.Token == token {
		_ = cache.Remove(ctx, key)
		return true, nil
	}

	return false, nil
}

// GetOTP retrieves OTP data from Redis without removing it
func (s *OTPService) GetOTP(ctx context.Context, identity string) (*OTPData, error) {
	key := s.generateOTPKey(identity)

	otpDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("OTP not found or expired: %w", err)
	}

	// For simplicity, we'll return a basic OTPData structure
	// In a real implementation, you would unmarshal the JSON
	return &OTPData{
		OTP:       otpDataStr,
		Identity:  identity,
		Type:      "phone",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(DefaultOTPExpiration),
	}, nil
}

// GetResetLink retrieves reset link data from Redis without removing it
func (s *OTPService) GetResetLink(ctx context.Context, identity string) (*ResetLinkData, error) {
	key := s.generateResetLinkKey(identity)

	resetDataStr, err := cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("Reset link not found or expired: %w", err)
	}

	var resetData ResetLinkData
	err = json.Unmarshal([]byte(resetDataStr), &resetData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reset link data: %w", err)
	}

	// For simplicity, we'll return a basic ResetLinkData structure
	// In a real implementation, you would unmarshal the JSON
	return &resetData, nil
}

// RemoveOTP removes an OTP from Redis
func (s *OTPService) RemoveOTP(ctx context.Context, identity string) error {
	key := s.generateOTPKey(identity)
	return cache.Remove(ctx, key)
}

// RemoveResetLink removes a reset link from Redis
func (s *OTPService) RemoveResetLink(ctx context.Context, identity string) error {
	key := s.generateResetLinkKey(identity)
	return cache.Remove(ctx, key)
}

// IsOTPExists checks if an OTP exists for the given identity
func (s *OTPService) IsOTPExists(ctx context.Context, identity string) (bool, error) {
	key := s.generateOTPKey(identity)

	_, err := cache.Get(ctx, key)
	if err != nil {
		return false, nil // OTP doesn't exist
	}

	return true, nil
}

// IsResetLinkExists checks if a reset link exists for the given identity
func (s *OTPService) IsResetLinkExists(ctx context.Context, identity string) (bool, error) {
	key := s.generateResetLinkKey(identity)

	_, err := cache.Get(ctx, key)
	if err != nil {
		return false, nil // Reset link doesn't exist
	}

	return true, nil
}

// generateOTPKey generates a unique key for storing OTP in Redis
func (s *OTPService) generateOTPKey(identity string) string {
	return fmt.Sprintf("%s:phone:%s", OTPKeyPrefix, identity)
}

// generateResetLinkKey generates a unique key for storing reset link in Redis
func (s *OTPService) generateResetLinkKey(identity string) string {
	return fmt.Sprintf("%s:email:%s", ResetLinkKeyPrefix, identity)
}

// generateResetToken generates a secure random token for reset links
func (s *OTPService) generateResetToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ResendOTP generates a new OTP and replaces the existing one
func (s *OTPService) ResendOTP(ctx context.Context, identity string, length int, expiration time.Duration) (string, error) {
	// Remove existing OTP if it exists
	_ = s.RemoveOTP(ctx, identity)

	// Generate and store new OTP
	return s.GenerateAndStoreOTP(ctx, identity, length, expiration)
}

// ResendResetLink generates a new reset link and replaces the existing one
func (s *OTPService) ResendResetLink(ctx context.Context, identity string, expiration time.Duration) (string, error) {
	// Remove existing reset link if it exists
	_ = s.RemoveResetLink(ctx, identity)

	// Generate and store new reset link
	return s.GenerateAndStoreResetLink(ctx, identity, expiration)
}
