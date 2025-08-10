package utils

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/turahe/go-restfull/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims represents the JWT token claims structure
// This struct extends the standard JWT registered claims with custom user information
// and is used for both access and refresh tokens
type TokenClaims struct {
	UserID   uuid.UUID `json:"user_id"`   // Unique identifier for the authenticated user
	Username string    `json:"username"`   // User's display name or username
	Email    string    `json:"email"`      // User's email address
	Type     string    `json:"type"`       // Token type: "access" or "refresh"
	jwt.RegisteredClaims                    // Standard JWT claims (exp, iat, nbf, iss, sub, jti)
}

// TokenPair represents a complete set of authentication tokens
// This struct contains both access and refresh tokens along with expiration information
// for complete JWT-based authentication flow
type TokenPair struct {
	AccessToken  string `json:"access_token"`  // Short-lived access token for API calls
	RefreshToken string `json:"refresh_token"` // Long-lived refresh token for getting new access tokens
	ExpiresIn    int64  `json:"expires_in"`    // Unix timestamp when access token expires
}

// GenerateAccessToken creates a JWT access token for user authentication
// Access tokens are short-lived (default 15 minutes) and used for API authorization
// The token contains user identity information and standard JWT claims
//
// Parameters:
//   - userID: unique identifier for the user
//   - username: user's display name
//   - email: user's email address
//
// Returns:
//   - string: the signed JWT access token
//   - error: any error that occurred during token generation
//
// Security notes:
//   - Access tokens have short expiration for security
//   - Tokens are signed with the application's JWT secret
//   - Each token has a unique ID (jti claim) for tracking
func GenerateAccessToken(userID uuid.UUID, username, email string) (string, error) {
	// Get application configuration for JWT settings
	cfg := config.GetConfig()
	
	// Set default expiration to 15 minutes for security
	expiration := 15 * time.Minute
	
	// Override with configured value if available
	if cfg.App.AccessTokenExpiration > 0 {
		expiration = time.Duration(cfg.App.AccessTokenExpiration) * time.Hour
	}

	// Create JWT claims with user information and standard fields
	claims := TokenClaims{
		UserID:   userID,   // User's unique identifier
		Username: username, // User's display name
		Email:    email,    // User's email address
		Type:     "access", // Mark as access token
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)), // Token expiration time
			IssuedAt:  jwt.NewNumericDate(time.Now()),                 // When token was issued
			NotBefore: jwt.NewNumericDate(time.Now()),                 // Token valid from now
			Issuer:    cfg.App.Name,                                   // Application name as issuer
			Subject:   userID.String(),                                // User ID as subject
			ID:        uuid.New().String(),                            // Unique token ID
		},
	}

	// Create and sign the JWT token with HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.App.JWTSecret))
}

// GenerateRefreshToken creates a JWT refresh token for obtaining new access tokens
// Refresh tokens are long-lived (7 days) and used to get new access tokens without
// requiring user re-authentication
//
// Parameters:
//   - userID: unique identifier for the user
//   - username: user's display name
//   - email: user's email address
//
// Returns:
//   - string: the signed JWT refresh token
//   - error: any error that occurred during token generation
//
// Security notes:
//   - Refresh tokens have longer expiration for user convenience
//   - Should be stored securely (HTTP-only cookies recommended)
//   - Can be revoked independently of access tokens
func GenerateRefreshToken(userID uuid.UUID, username, email string) (string, error) {
	// Get application configuration for JWT settings
	cfg := config.GetConfig()

	// Create JWT claims with user information and standard fields
	claims := TokenClaims{
		UserID:   userID,    // User's unique identifier
		Username: username,  // User's display name
		Email:    email,     // User's email address
		Type:     "refresh", // Mark as refresh token
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days expiration
			IssuedAt:  jwt.NewNumericDate(time.Now()),                          // When token was issued
			NotBefore: jwt.NewNumericDate(time.Now()),                          // Token valid from now
			Issuer:    cfg.App.Name,                                            // Application name as issuer
			Subject:   userID.String(),                                         // User ID as subject
			ID:        uuid.New().String(),                                     // Unique token ID
		},
	}

	// Create and sign the JWT token with HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.App.JWTSecret))
}

// GenerateTokenPair creates both access and refresh tokens for a user
// This is a convenience function that generates a complete authentication
// token set in a single call
//
// Parameters:
//   - userID: unique identifier for the user
//   - username: user's display name
//   - email: user's email address
//
// Returns:
//   - *TokenPair: complete set of access and refresh tokens with expiration info
//   - error: any error that occurred during token generation
//
// Usage example:
//   tokenPair, err := GenerateTokenPair(user.ID, user.Username, user.Email)
//   if err != nil {
//       // Handle error
//   }
//   // Return tokens to client
func GenerateTokenPair(userID uuid.UUID, username, email string) (*TokenPair, error) {
	// Get application configuration for JWT settings
	cfg := config.GetConfig()
	
	// Generate access token
	accessToken, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(userID, username, email)
	if err != nil {
		return nil, err
	}

	// Calculate expiration time for access token
	expiration := 15 * time.Minute
	if cfg.App.AccessTokenExpiration > 0 {
		expiration = time.Duration(cfg.App.AccessTokenExpiration) * time.Hour
	}
	expiresAt := time.Now().Add(expiration).Unix()

	// Return complete token pair
	return &TokenPair{
		AccessToken:  accessToken,  // Short-lived access token
		RefreshToken: refreshToken, // Long-lived refresh token
		ExpiresIn:    expiresAt,    // Unix timestamp of expiration
	}, nil
}

// ValidateToken validates a JWT token and returns the decoded claims
// This function handles token parsing, signature verification, and basic validation
// It supports both raw tokens and "Bearer " prefixed tokens
//
// Parameters:
//   - tokenString: the JWT token string to validate
//
// Returns:
//   - *TokenClaims: the decoded token claims if valid
//   - error: validation error if the token is invalid
//
// Security notes:
//   - Verifies token signature using application secret
//   - Checks token expiration and validity
//   - Supports both raw and Bearer token formats
func ValidateToken(tokenString string) (*TokenClaims, error) {
	// Get application configuration for JWT secret
	cfg := config.GetConfig()

	// Sanitize token: remove leading/trailing whitespace
	tokenString = strings.TrimSpace(tokenString)
	
	// Remove "Bearer " prefix if present (common in Authorization headers)
	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}

	// Parse and validate the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify that the signing method is HMAC (not RSA or other methods)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Return the secret key for signature verification
		return []byte(cfg.App.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract and return claims if token is valid
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAccessToken validates an access token specifically
// This function ensures the token is not only valid but also of the correct type
// for access token operations
//
// Parameters:
//   - tokenString: the JWT token string to validate
//
// Returns:
//   - *TokenClaims: the decoded access token claims if valid
//   - error: validation error if the token is invalid or wrong type
//
// Usage example:
//   claims, err := ValidateAccessToken(authHeader)
//   if err != nil {
//       // Handle invalid or expired access token
//   }
//   // Use claims.UserID for authorization
func ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	// First validate the token structure and signature
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Ensure this is actually an access token
	if claims.Type != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token specifically
// This function ensures the token is not only valid but also of the correct type
// for refresh token operations
//
// Parameters:
//   - tokenString: the JWT token string to validate
//
// Returns:
//   - *TokenClaims: the decoded refresh token claims if valid
//   - error: validation error if the token is invalid or wrong type
//
// Usage example:
//   claims, err := ValidateRefreshToken(refreshToken)
//   if err != nil {
//       // Handle invalid or expired refresh token
//   }
//   // Generate new token pair
func ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	// First validate the token structure and signature
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Ensure this is actually a refresh token
	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
// This function allows users to get new access tokens without re-authenticating
// as long as their refresh token is still valid
//
// Parameters:
//   - refreshToken: the valid refresh token to use for generating new tokens
//
// Returns:
//   - *TokenPair: new set of access and refresh tokens
//   - error: any error that occurred during token generation
//
// Usage example:
//   newTokens, err := RefreshAccessToken(refreshToken)
//   if err != nil {
//       // Handle refresh failure
//   }
//   // Return new tokens to client
func RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	// Validate the refresh token first
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair using the refresh token claims
	return GenerateTokenPair(claims.UserID, claims.Username, claims.Email)
}

// GenerateOTP generates a random numeric OTP (One-Time Password) of specified length
// This function creates a secure random string of digits for authentication purposes
// such as SMS verification or two-factor authentication
//
// Parameters:
//   - length: the number of digits in the OTP (e.g., 6 for "123456")
//
// Returns:
//   - string: the generated OTP code
//
// Security notes:
//   - Uses math/rand for simplicity (consider crypto/rand for production)
//   - Generates only numeric digits (0-9)
//   - Length should be appropriate for the use case (typically 4-8 digits)
//
// Usage example:
//   otp := GenerateOTP(6) // Generates "123456" (example)
//   // Send OTP via SMS or email
func GenerateOTP(length int) string {
	// Define the character set for OTP generation (digits only)
	digits := "0123456789"
	
	// Create byte slice to hold the OTP
	result := make([]byte, length)
	
	// Generate random digits for each position
	for i := range result {
		result[i] = digits[rand.Intn(len(digits))]
	}
	
	return string(result)
}
