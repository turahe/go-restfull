package test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"webapi/config"
	"webapi/internal/helper/utils"
)

func TestJWTTokenGeneration(t *testing.T) {
	// Set up test config
	config.SetConfig("config/config.testing.yaml")
	
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Test token pair generation
	tokenPair, err := utils.GenerateTokenPair(userID, username, email)
	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, int64(15*60), tokenPair.ExpiresIn)

	// Test access token validation
	accessClaims, err := utils.ValidateAccessToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, username, accessClaims.Username)
	assert.Equal(t, email, accessClaims.Email)
	assert.Equal(t, "access", accessClaims.Type)

	// Test refresh token validation
	refreshClaims, err := utils.ValidateRefreshToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, username, refreshClaims.Username)
	assert.Equal(t, email, refreshClaims.Email)
	assert.Equal(t, "refresh", refreshClaims.Type)

	// Test token refresh
	newTokenPair, err := utils.RefreshAccessToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, newTokenPair)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)

	// Verify new tokens are different
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)
}

func TestJWTTokenValidation(t *testing.T) {
	// Set up test config
	config.SetConfig("config/config.testing.yaml")
	
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Test invalid token
	_, err := utils.ValidateToken("invalid-token")
	assert.Error(t, err)

	// Test expired token (this would require mocking time)
	tokenPair, err := utils.GenerateTokenPair(userID, username, email)
	assert.NoError(t, err)

	// Test wrong token type
	_, err = utils.ValidateAccessToken(tokenPair.RefreshToken)
	assert.Error(t, err)

	_, err = utils.ValidateRefreshToken(tokenPair.AccessToken)
	assert.Error(t, err)
}

func TestJWTTokenClaims(t *testing.T) {
	// Set up test config
	config.SetConfig("config/config.testing.yaml")
	
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(userID, username, email)
	assert.NoError(t, err)

	// Validate and check claims
	claims, err := utils.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "access", claims.Type)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
} 