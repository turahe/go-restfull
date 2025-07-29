package utils

import (
	"testing"
	"time"

	"webapi/config"

	"github.com/google/uuid"
)

func init() {
	// Initialize test config
	testConfig := &config.Config{
		App: config.App{
			Name:                  "TestApp",
			NameSlug:              "test-app",
			JWTSecret:             "test-jwt-secret-key-for-testing-purposes-only",
			AccessTokenExpiration: 1, // 1 hour for testing
		},
	}
	config.SetConfig(testConfig)
}

func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	token, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		t.Errorf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Error("Generated access token should not be empty")
	}

	// Validate the generated token
	claims, err := ValidateAccessToken(token)
	if err != nil {
		t.Errorf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}
	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}
	if claims.Type != "access" {
		t.Errorf("Type = %v, want %v", claims.Type, "access")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	token, err := GenerateRefreshToken(userID, username, email)
	if err != nil {
		t.Errorf("GenerateRefreshToken() error = %v", err)
	}

	if token == "" {
		t.Error("Generated refresh token should not be empty")
	}

	// Validate the generated token
	claims, err := ValidateRefreshToken(token)
	if err != nil {
		t.Errorf("ValidateRefreshToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}
	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}
	if claims.Type != "refresh" {
		t.Errorf("Type = %v, want %v", claims.Type, "refresh")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	tokenPair, err := GenerateTokenPair(userID, username, email)
	if err != nil {
		t.Errorf("GenerateTokenPair() error = %v", err)
	}

	if tokenPair == nil {
		t.Error("TokenPair should not be nil")
	}

	if tokenPair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}

	if tokenPair.ExpiresIn <= 0 {
		t.Error("ExpiresIn should be greater than 0")
	}

	// Validate access token
	accessClaims, err := ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Errorf("ValidateAccessToken() error = %v", err)
	}

	if accessClaims.UserID != userID {
		t.Errorf("Access token UserID = %v, want %v", accessClaims.UserID, userID)
	}

	// Validate refresh token
	refreshClaims, err := ValidateRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Errorf("ValidateRefreshToken() error = %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("Refresh token UserID = %v, want %v", refreshClaims.UserID, userID)
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Test valid token
	token, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	// Test invalid token
	_, err = ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}

	// Test empty token
	_, err = ValidateToken("")
	if err == nil {
		t.Error("ValidateToken() should return error for empty token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Test valid access token
	accessToken, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := ValidateAccessToken(accessToken)
	if err != nil {
		t.Errorf("ValidateAccessToken() error = %v", err)
	}

	if claims.Type != "access" {
		t.Errorf("Type = %v, want %v", claims.Type, "access")
	}

	// Test with refresh token (should fail)
	refreshToken, err := GenerateRefreshToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	_, err = ValidateAccessToken(refreshToken)
	if err == nil {
		t.Error("ValidateAccessToken() should return error for refresh token")
	}
}

func TestValidateRefreshToken(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Test valid refresh token
	refreshToken, err := GenerateRefreshToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Errorf("ValidateRefreshToken() error = %v", err)
	}

	if claims.Type != "refresh" {
		t.Errorf("Type = %v, want %v", claims.Type, "refresh")
	}

	// Test with access token (should fail)
	accessToken, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ValidateRefreshToken(accessToken)
	if err == nil {
		t.Error("ValidateRefreshToken() should return error for access token")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Generate initial token pair
	initialPair, err := GenerateTokenPair(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	// Refresh access token using refresh token
	newPair, err := RefreshAccessToken(initialPair.RefreshToken)
	if err != nil {
		t.Errorf("RefreshAccessToken() error = %v", err)
	}

	if newPair == nil {
		t.Error("New token pair should not be nil")
	}

	if newPair.AccessToken == "" {
		t.Error("New access token should not be empty")
	}

	if newPair.RefreshToken == "" {
		t.Error("New refresh token should not be empty")
	}

	// Validate new access token
	claims, err := ValidateAccessToken(newPair.AccessToken)
	if err != nil {
		t.Errorf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	// Test with invalid refresh token
	_, err = RefreshAccessToken("invalid.refresh.token")
	if err == nil {
		t.Error("RefreshAccessToken() should return error for invalid refresh token")
	}
}

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "4 digit OTP",
			length: 4,
		},
		{
			name:   "6 digit OTP",
			length: 6,
		},
		{
			name:   "8 digit OTP",
			length: 8,
		},
		{
			name:   "1 digit OTP",
			length: 1,
		},
		{
			name:   "10 digit OTP",
			length: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp := GenerateOTP(tt.length)

			if len(otp) != tt.length {
				t.Errorf("OTP length = %d, want %d", len(otp), tt.length)
			}

			// Verify all characters are digits
			for _, char := range otp {
				if char < '0' || char > '9' {
					t.Errorf("OTP contains non-digit character: %c", char)
				}
			}
		})
	}
}

func TestGenerateOTP_Consistency(t *testing.T) {
	// Test that OTPs are different on each generation
	otp1 := GenerateOTP(6)
	otp2 := GenerateOTP(6)
	otp3 := GenerateOTP(6)

	if otp1 == otp2 || otp1 == otp3 || otp2 == otp3 {
		t.Error("Generated OTPs should be different")
	}

	// Verify all are 6 digits
	if len(otp1) != 6 || len(otp2) != 6 || len(otp3) != 6 {
		t.Error("All OTPs should be 6 digits")
	}
}

func TestTokenExpiration(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Generate access token
	token, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Token should be valid immediately
	claims, err := ValidateAccessToken(token)
	if err != nil {
		t.Errorf("ValidateAccessToken() error = %v", err)
	}

	// Check that expiration is in the future
	if claims.ExpiresAt == nil {
		t.Error("Token should have expiration time")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token expiration should be in the future")
	}
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateAccessToken(userID, username, email)
	}
}

func BenchmarkGenerateRefreshToken(b *testing.B) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRefreshToken(userID, username, email)
	}
}

func BenchmarkGenerateTokenPair(b *testing.B) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateTokenPair(userID, username, email)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	token, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		b.Fatalf("GenerateAccessToken() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(token)
	}
}

func BenchmarkGenerateOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateOTP(6)
	}
}
