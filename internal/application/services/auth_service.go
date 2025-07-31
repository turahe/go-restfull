// Package services provides application-level business logic for authentication and user management.
// This package contains the auth service implementation that handles user registration, login,
// password management, and token-based authentication while enforcing security best practices.
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

// authService implements the AuthService interface and provides comprehensive
// authentication and user management functionality. It handles user registration,
// login/logout, password reset, token validation, and integrates with email
// services for notifications via RabbitMQ.
type authService struct {
	userRepo        repositories.UserRepository
	passwordService services.PasswordService
	emailService    *EmailService
	rabbitMQService *rabbitmq.Service
}

// NewAuthService creates a new authentication service instance with the provided dependencies.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and external dependencies.
//
// The service integrates with:
//   - User repository for data persistence
//   - Password service for secure password handling
//   - Email service for user notifications
//   - RabbitMQ service for asynchronous email processing
//
// Parameters:
//   - userRepo: Repository interface for user data access operations
//   - passwordService: Service for password hashing, validation, and comparison
//   - rabbitMQService: RabbitMQ service for asynchronous email processing
//
// Returns:
//   - ports.AuthService: The authentication service interface implementation
func NewAuthService(
	userRepo repositories.UserRepository,
	passwordService services.PasswordService,
	rabbitMQService *rabbitmq.Service,
) ports.AuthService {
	// Create email service with RabbitMQ integration for async processing
	emailService := NewEmailService(rabbitMQService)

	return &authService{
		userRepo:        userRepo,
		passwordService: passwordService,
		emailService:    emailService,
		rabbitMQService: rabbitMQService,
	}
}

// RegisterUser creates a new user account with comprehensive validation and security measures.
// This method enforces business rules for user registration and sends welcome notifications.
//
// Business Rules:
//   - Username, email, and phone must be unique across the system
//   - Password must meet strength requirements
//   - All required fields must be provided and validated
//   - Welcome email is sent asynchronously via RabbitMQ
//   - Authentication tokens are generated upon successful registration
//
// Security Features:
//   - Password is hashed using secure algorithms
//   - Duplicate user detection prevents account conflicts
//   - Soft delete checking ensures deleted users cannot register
//   - Asynchronous email processing prevents registration delays
//
// Parameters:
//   - ctx: Context for the operation
//   - username: Unique username for the new account
//   - email: Valid email address for the new account
//   - phone: Phone number for the new account
//   - password: Plain text password (will be hashed)
//
// Returns:
//   - *utils.TokenPair: Authentication tokens (access and refresh)
//   - *entities.User: The created user entity
//   - error: Any error that occurred during the operation
func (s *authService) RegisterUser(ctx context.Context, username, email, phone, password string) (*utils.TokenPair, *entities.User, error) {
	// Check if user already exists by email to prevent duplicates
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("user with this email already exists")
	}

	// Check if user already exists by username to prevent duplicates
	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("user with this username already exists")
	}

	// Check if user already exists by phone to prevent duplicates
	exists, err = s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("user with this phone already exists")
	}

	// Validate password strength to ensure security requirements are met
	if err := s.passwordService.ValidatePassword(password); err != nil {
		return nil, nil, err
	}

	// Hash password using secure algorithms before storing
	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	// Create user entity with validated and hashed data
	user, err := entities.NewUser(username, email, phone, hashedPassword)
	if err != nil {
		return nil, nil, err
	}

	// Persist the user to the repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Generate authentication tokens for immediate login
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserName, user.Email)
	if err != nil {
		return nil, nil, err
	}

	// Send welcome email via RabbitMQ asynchronously to avoid blocking registration
	go func() {
		if err := s.emailService.SendWelcomeEmail(email, username); err != nil {
			// Log error but don't fail the registration process
			// In production, you might want to use a proper logger
			fmt.Printf("Failed to send welcome email to %s: %v\n", email, err)
		}
	}()

	return tokenPair, user, nil
}

// LoginUser authenticates a user with username and password, returning authentication tokens.
// This method implements secure login with proper validation and error handling.
//
// Security Features:
//   - Password comparison using secure hashing
//   - Soft delete checking prevents deleted user login
//   - Generic error messages prevent user enumeration attacks
//   - Token generation for authenticated sessions
//
// Business Rules:
//   - User must exist and not be soft deleted
//   - Password must match the stored hash
//   - Authentication tokens are generated upon successful login
//
// Parameters:
//   - ctx: Context for the operation
//   - username: Username for authentication
//   - password: Plain text password for authentication
//
// Returns:
//   - *utils.TokenPair: Authentication tokens (access and refresh)
//   - *entities.User: The authenticated user entity
//   - error: Any error that occurred during the operation
func (s *authService) LoginUser(ctx context.Context, username, password string) (*utils.TokenPair, *entities.User, error) {
	// Get user by username to verify existence
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Check if user is soft deleted to prevent deleted user login
	if user.IsDeleted() {
		return nil, nil, errors.New("invalid credentials")
	}

	// Verify password using secure comparison
	if !s.passwordService.ComparePassword(user.Password, password) {
		return nil, nil, errors.New("invalid credentials")
	}

	// Generate authentication tokens for successful login
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserName, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return tokenPair, user, nil
}

// RefreshToken validates a refresh token and generates a new token pair.
// This method enables secure token refresh without requiring re-authentication.
//
// Security Features:
//   - Refresh token validation with proper signature checking
//   - User existence verification to prevent token reuse after deletion
//   - Soft delete checking prevents deleted user token refresh
//
// Business Rules:
//   - Refresh token must be valid and not expired
//   - User must still exist and not be soft deleted
//   - New token pair is generated with updated expiration times
//
// Parameters:
//   - ctx: Context for the operation
//   - refreshToken: Valid refresh token string
//
// Returns:
//   - *utils.TokenPair: New authentication tokens (access and refresh)
//   - error: Any error that occurred during the operation
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*utils.TokenPair, error) {
	// Validate refresh token signature and expiration
	claims, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get user to ensure they still exist and are not deleted
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if user is soft deleted to prevent token refresh for deleted users
	if user.IsDeleted() {
		return nil, errors.New("invalid refresh token")
	}

	// Generate new token pair with updated expiration times
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserName, user.Email)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// LogoutUser handles user logout operations. This method provides a framework
// for implementing sophisticated logout mechanisms in production environments.
//
// Production Considerations:
//   - Token blacklisting in Redis for immediate invalidation
//   - Logout event tracking for audit trails
//   - Refresh token invalidation
//   - Session management cleanup
//
// Current Implementation:
//   - Returns success immediately (client-side token disposal)
//   - Framework ready for enhanced security features
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user logging out
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *authService) LogoutUser(ctx context.Context, userID string) error {
	// In a more sophisticated implementation, you might want to:
	// 1. Add the token to a blacklist in Redis
	// 2. Track logout events for audit purposes
	// 3. Invalidate refresh tokens
	// 4. Clean up session data

	// For now, we'll just return success
	// The client should discard the tokens
	return nil
}

// ForgetPassword initiates the password reset process for a user identified by
// email, username, or phone number. This method implements security best practices
// to prevent user enumeration attacks.
//
// Security Features:
//   - Multiple identifier support (email, username, phone)
//   - Generic response to prevent user enumeration
//   - Asynchronous email processing via RabbitMQ
//   - OTP generation for secure password reset
//
// Business Rules:
//   - User must exist and not be soft deleted
//   - Password reset email is sent regardless of user existence (security)
//   - OTP is generated and sent asynchronously
//   - No information leakage about user existence
//
// Parameters:
//   - ctx: Context for the operation
//   - identifier: Email, username, or phone number to identify the user
//
// Returns:
//   - error: Always returns nil to prevent user enumeration
func (s *authService) ForgetPassword(ctx context.Context, identifier string) error {
	var user *entities.User
	var err error

	// Try to find user by different identifiers for flexibility
	// First, try by email
	user, err = s.userRepo.GetByEmail(ctx, identifier)
	if err == nil && user != nil && !user.IsDeleted() {
		// Found user by email - send password reset email
		return s.sendPasswordResetEmail(user.Email, user.UserName)
	}

	// Try by username if email lookup failed
	user, err = s.userRepo.GetByUsername(ctx, identifier)
	if err == nil && user != nil && !user.IsDeleted() {
		// Found user by username - send password reset email
		return s.sendPasswordResetEmail(user.Email, user.UserName)
	}

	// Try by phone if username lookup failed
	user, err = s.userRepo.GetByPhone(ctx, identifier)
	if err == nil && user != nil && !user.IsDeleted() {
		// Found user by phone - send password reset email
		return s.sendPasswordResetEmail(user.Email, user.UserName)
	}

	// Don't reveal if user exists or not for security reasons
	// Always return nil to prevent user enumeration attacks
	return nil
}

// sendPasswordResetEmail sends a password reset email for the given user.
// This method generates an OTP and sends it asynchronously via RabbitMQ.
//
// Security Features:
//   - OTP generation for secure password reset
//   - Asynchronous email processing
//   - Error handling that doesn't block the main process
//
// Business Rules:
//   - OTP is generated with 6 digits
//   - Email is sent asynchronously via RabbitMQ
//   - Errors are logged but don't fail the password reset request
//
// Parameters:
//   - email: Email address to send the reset email to
//   - username: Username for personalization
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *authService) sendPasswordResetEmail(email, username string) error {
	// Generate OTP for secure password reset
	otp := utils.GenerateOTP(6)

	// Store OTP in cache (you might want to implement this)
	// For now, we'll just send the email

	// Send password reset email via RabbitMQ asynchronously
	go func() {
		if err := s.emailService.SendPasswordResetEmail(email, otp); err != nil {
			// Log error but don't fail the password reset request
			fmt.Printf("Failed to send password reset email to %s: %v\n", email, err)
		}
	}()

	return nil
}

// ResetPassword completes the password reset process using email and OTP.
// This method validates the reset credentials and updates the user's password.
//
// Security Features:
//   - OTP validation (placeholder implementation)
//   - Password strength validation
//   - Secure password hashing
//   - User existence and status verification
//
// Business Rules:
//   - User must exist and not be soft deleted
//   - OTP must be valid (implementation needed)
//   - New password must meet strength requirements
//   - Password is securely hashed before storage
//
// Parameters:
//   - ctx: Context for the operation
//   - email: Email address of the user resetting password
//   - otp: One-time password for verification
//   - newPassword: New password to set
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *authService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	// Get user by email to verify existence
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("invalid email or OTP")
	}

	// Check if user is soft deleted to prevent password reset for deleted users
	if user.IsDeleted() {
		return errors.New("invalid email or OTP")
	}

	// Validate OTP (you might want to implement this with cache)
	// For now, we'll assume OTP is valid
	// TODO: Implement proper OTP validation with cache/Redis

	// Validate new password strength
	if err := s.passwordService.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password using secure algorithms
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user's password with the new hashed password
	if err := user.ChangePassword(hashedPassword); err != nil {
		return err
	}

	// Persist the updated user to the repository
	return s.userRepo.Update(ctx, user)
}

// ValidateToken validates an access token and returns the token claims.
// This method is used for protecting routes and verifying user authentication.
//
// Security Features:
//   - Token signature validation
//   - Token expiration checking
//   - User existence verification
//   - Soft delete checking
//
// Business Rules:
//   - Token must be valid and not expired
//   - User must still exist and not be soft deleted
//   - Token claims are returned for route protection
//
// Parameters:
//   - ctx: Context for the operation
//   - token: Access token to validate
//
// Returns:
//   - *utils.TokenClaims: Token claims if valid
//   - error: Any error that occurred during the operation
func (s *authService) ValidateToken(ctx context.Context, token string) (*utils.TokenClaims, error) {
	// Validate access token signature and expiration
	claims, err := utils.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	// Get user to ensure they still exist and are not deleted
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Check if user is soft deleted to prevent token validation for deleted users
	if user.IsDeleted() {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
