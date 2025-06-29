package test

import (
	"testing"

	"webapi/config"
	"webapi/internal/helper/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTTokenGenerationSimple(t *testing.T) {
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
