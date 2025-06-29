package test

import (
	"context"
	"testing"
	"time"

	"webapi/config"
	"webapi/internal/helper/utils"

	"github.com/stretchr/testify/assert"
)

func TestOTPService(t *testing.T) {
	// Set up test config
	config.SetConfig("config/config.testing.yaml")

	// Create OTP service
	otpService := utils.NewOTPService()
	ctx := context.Background()

	// Test data
	email := "test@example.com"
	phone := "1234567890"
	otpLength := 6

	t.Run("Generate and store OTP", func(t *testing.T) {
		// Generate and store OTP
		otp, err := otpService.GenerateAndStoreOTP(ctx, email, otpLength, 5*time.Minute)
		assert.NoError(t, err)
		assert.NotEmpty(t, otp)
		assert.Len(t, otp, otpLength)

		// Check if OTP exists
		exists, err := otpService.IsOTPExists(ctx, email)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Validate OTP
		isValid, err := otpService.ValidateOTP(ctx, email, otp)
		assert.NoError(t, err)
		assert.True(t, isValid)

		// OTP should be removed after validation
		exists, err = otpService.IsOTPExists(ctx, email)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Generate and store OTP for phone", func(t *testing.T) {
		// Generate and store OTP for phone
		otp, err := otpService.GenerateAndStoreOTP(ctx, phone, otpLength, 5*time.Minute)
		assert.NoError(t, err)
		assert.NotEmpty(t, otp)
		assert.Len(t, otp, otpLength)

		// Check if OTP exists
		exists, err := otpService.IsOTPExists(ctx, phone)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Validate OTP
		isValid, err := otpService.ValidateOTP(ctx, phone, otp)
		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("Invalid OTP validation", func(t *testing.T) {
		// Generate OTP
		_, err := otpService.GenerateAndStoreOTP(ctx, email, otpLength, 5*time.Minute)
		assert.NoError(t, err)

		// Try to validate with wrong OTP
		isValid, err := otpService.ValidateOTP(ctx, email, "000000")
		assert.NoError(t, err)
		assert.False(t, isValid)

		// Clean up
		_ = otpService.RemoveOTP(ctx, email)
	})

	t.Run("OTP expiration", func(t *testing.T) {
		// Generate OTP with very short expiration
		otp, err := otpService.GenerateAndStoreOTP(ctx, email, otpLength, 1*time.Millisecond)
		assert.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Try to validate expired OTP
		isValid, err := otpService.ValidateOTP(ctx, email, otp)
		assert.Error(t, err)
		assert.False(t, isValid)
	})

	t.Run("Resend OTP", func(t *testing.T) {
		// Generate initial OTP
		otp1, err := otpService.GenerateAndStoreOTP(ctx, email, otpLength, 5*time.Minute)
		assert.NoError(t, err)

		// Resend OTP
		otp2, err := otpService.ResendOTP(ctx, email, otpLength, 5*time.Minute)
		assert.NoError(t, err)

		// OTPs should be different
		assert.NotEqual(t, otp1, otp2)

		// First OTP should not be valid anymore
		isValid, err := otpService.ValidateOTP(ctx, email, otp1)
		assert.Error(t, err)
		assert.False(t, isValid)

		// Second OTP should be valid
		isValid, err = otpService.ValidateOTP(ctx, email, otp2)
		assert.NoError(t, err)
		assert.True(t, isValid)
	})
}
