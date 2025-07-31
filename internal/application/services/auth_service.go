package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/turahe/go-restfull/pkg/rabbitmq"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/helper/utils"
)

// authService implements the AuthService interface
type authService struct {
	userRepo        repositories.UserRepository
	passwordService services.PasswordService
	emailService    *EmailService
	rabbitMQService *rabbitmq.Service
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	userRepo repositories.UserRepository,
	passwordService services.PasswordService,
	rabbitMQService *rabbitmq.Service,
) ports.AuthService {
	emailService := NewEmailService(rabbitMQService)

	return &authService{
		userRepo:        userRepo,
		passwordService: passwordService,
		emailService:    emailService,
		rabbitMQService: rabbitMQService,
	}
}

func (s *authService) RegisterUser(ctx context.Context, username, email, phone, password string) (*utils.TokenPair, *entities.User, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("user with this email already exists")
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("user with this username already exists")
	}

	exists, err = s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("user with this phone already exists")
	}

	// Validate password strength
	if err := s.passwordService.ValidatePassword(password); err != nil {
		return nil, nil, err
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	// Create user entity
	user, err := entities.NewUser(username, email, phone, hashedPassword)
	if err != nil {
		return nil, nil, err
	}

	// Save to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Generate authentication tokens
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserName, user.Email)
	if err != nil {
		return nil, nil, err
	}

	// Send welcome email via RabbitMQ (async)
	go func() {
		if err := s.emailService.SendWelcomeEmail(email, username); err != nil {
			// Log error but don't fail the registration
			// In production, you might want to use a proper logger
			fmt.Printf("Failed to send welcome email to %s: %v\n", email, err)
		}
	}()

	return tokenPair, user, nil
}

func (s *authService) LoginUser(ctx context.Context, username, password string) (*utils.TokenPair, *entities.User, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return nil, nil, errors.New("invalid credentials")
	}

	// Verify password
	if !s.passwordService.ComparePassword(user.Password, password) {
		return nil, nil, errors.New("invalid credentials")
	}

	// Generate authentication tokens
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserName, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return tokenPair, user, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*utils.TokenPair, error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get user to ensure they still exist
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return nil, errors.New("invalid refresh token")
	}

	// Generate new token pair
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserName, user.Email)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *authService) LogoutUser(ctx context.Context, userID string) error {
	// In a more sophisticated implementation, you might want to:
	// 1. Add the token to a blacklist in Redis
	// 2. Track logout events
	// 3. Invalidate refresh tokens

	// For now, we'll just return success
	// The client should discard the tokens
	return nil
}

func (s *authService) ForgetPassword(ctx context.Context, identifier string) error {
	var user *entities.User
	var err error

	// Try to find user by different identifiers
	// First, try by email
	user, err = s.userRepo.GetByEmail(ctx, identifier)
	if err == nil && user != nil && !user.IsDeleted() {
		// Found user by email
		return s.sendPasswordResetEmail(user.Email, user.UserName)
	}

	// Try by username
	user, err = s.userRepo.GetByUsername(ctx, identifier)
	if err == nil && user != nil && !user.IsDeleted() {
		// Found user by username
		return s.sendPasswordResetEmail(user.Email, user.UserName)
	}

	// Try by phone
	user, err = s.userRepo.GetByPhone(ctx, identifier)
	if err == nil && user != nil && !user.IsDeleted() {
		// Found user by phone
		return s.sendPasswordResetEmail(user.Email, user.UserName)
	}

	// Don't reveal if user exists or not for security
	return nil
}

// sendPasswordResetEmail sends a password reset email for the given user
func (s *authService) sendPasswordResetEmail(email, username string) error {
	// Generate OTP
	otp := utils.GenerateOTP(6)

	// Store OTP in cache (you might want to implement this)
	// For now, we'll just send the email

	// Send password reset email via RabbitMQ (async)
	go func() {
		if err := s.emailService.SendPasswordResetEmail(email, otp); err != nil {
			// Log error but don't fail the password reset request
			fmt.Printf("Failed to send password reset email to %s: %v\n", email, err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("invalid email or OTP")
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return errors.New("invalid email or OTP")
	}

	// Validate OTP (you might want to implement this with cache)
	// For now, we'll assume OTP is valid

	// Validate new password
	if err := s.passwordService.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := user.ChangePassword(hashedPassword); err != nil {
		return err
	}

	// Save to repository
	return s.userRepo.Update(ctx, user)
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*utils.TokenClaims, error) {
	// Validate access token
	claims, err := utils.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	// Get user to ensure they still exist
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
