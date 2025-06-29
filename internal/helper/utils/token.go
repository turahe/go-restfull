package utils

import (
	"errors"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"webapi/config"
)

type TokenClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Type     string    `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateAccessToken generates a JWT access token
func GenerateAccessToken(userID uuid.UUID, username, email string) (string, error) {
	cfg := config.GetConfig()
	
	claims := TokenClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.Name,
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.App.JWTSecret))
}

// GenerateRefreshToken generates a JWT refresh token
func GenerateRefreshToken(userID uuid.UUID, username, email string) (string, error) {
	cfg := config.GetConfig()
	
	claims := TokenClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.Name,
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.App.JWTSecret))
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uuid.UUID, username, email string) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, username, email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(userID, username, email)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*TokenClaims, error) {
	cfg := config.GetConfig()
	
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.App.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAccessToken validates an access token specifically
func ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token specifically
func ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return GenerateTokenPair(claims.UserID, claims.Username, claims.Email)
}

// GenerateOTP generates a random numeric OTP of the given length
func GenerateOTP(length int) string {
	digits := "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = digits[rand.Intn(len(digits))]
	}
	return string(result)
}
