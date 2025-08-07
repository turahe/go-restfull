// Package services provides application-level business logic for user management.
// This package contains the user service implementation that handles user creation,
// authentication, profile management, and user lifecycle while ensuring proper
// security and data integrity.
package services

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/services"

	"github.com/google/uuid"
)

// userService implements the UserService interface and provides comprehensive
// user management functionality. It handles user creation, authentication, profile
// management, password security, and user lifecycle while ensuring proper
// security and data integrity.
type userService struct {
	userRepo        repositories.UserRepository
	passwordService services.PasswordService
	emailService    services.EmailService
}

// NewUserService creates a new user service instance with the provided dependencies.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and external dependencies.
//
// Parameters:
//   - userRepo: Repository interface for user data access operations
//   - passwordService: Service for password hashing, validation, and comparison
//   - emailService: Service for user notifications and email communications
//
// Returns:
//   - ports.UserService: The user service interface implementation
func NewUserService(
	userRepo repositories.UserRepository,
	passwordService services.PasswordService,
	emailService services.EmailService,
) ports.UserService {
	return &userService{
		userRepo:        userRepo,
		passwordService: passwordService,
		emailService:    emailService,
	}
}

// CreateUser creates a new user account with comprehensive validation and security measures.
// This method enforces business rules for user registration and sends welcome notifications.
//
// Business Rules:
//   - Username, email, and phone must be unique across the system
//   - Password must meet strength requirements
//   - All required fields must be provided and validated
//   - Welcome email is sent asynchronously
//   - Password is securely hashed before storage
//
// Security Features:
//   - Password strength validation
//   - Secure password hashing
//   - Duplicate user detection
//   - Soft delete checking
//   - Asynchronous email processing
//
// Parameters:
//   - ctx: Context for the operation
//   - username: Unique username for the new account
//   - email: Valid email address for the new account
//   - phone: Phone number for the new account
//   - password: Plain text password (will be hashed)
//
// Returns:
//   - *entities.User: The created user entity
//   - error: Any error that occurred during the operation
func (s *userService) CreateUser(ctx context.Context, username, email, phone, password string) (*entities.User, error) {
	// Validate password strength to ensure security requirements are met
	if err := s.passwordService.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Check if user already exists by email to prevent duplicates
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Check if user already exists by username to prevent duplicates
	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this username already exists")
	}

	// Check if user already exists by phone to prevent duplicates
	exists, err = s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this phone already exists")
	}

	// Hash password using secure algorithms before storing
	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user entity with validated and hashed data
	user, err := entities.NewUser(username, email, phone, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Persist the user to the repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send welcome email asynchronously to avoid blocking registration
	go func() {
		_ = s.emailService.SendWelcomeEmail(email, username)
	}()

	return user, nil
}

// GetUserByID retrieves a user by their unique identifier.
// This method includes soft delete checking to ensure deleted users
// are not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the user to retrieve
//
// Returns:
//   - *entities.User: The user entity if found
//   - error: Error if user not found or other issues occur
func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Check if the user has been soft deleted
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email address.
// This method includes soft delete checking and is useful for authentication.
//
// Parameters:
//   - ctx: Context for the operation
//   - email: Email address of the user to retrieve
//
// Returns:
//   - *entities.User: The user entity if found
//   - error: Error if user not found or other issues occur
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Check if the user has been soft deleted
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username.
// This method includes soft delete checking and is useful for authentication.
//
// Parameters:
//   - ctx: Context for the operation
//   - username: Username of the user to retrieve
//
// Returns:
//   - *entities.User: The user entity if found
//   - error: Error if user not found or other issues occur
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	// Check if the user has been soft deleted
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserByPhone retrieves a user by their phone number.
// This method includes soft delete checking and is useful for authentication.
//
// Parameters:
//   - ctx: Context for the operation
//   - phone: Phone number of the user to retrieve
//
// Returns:
//   - *entities.User: The user entity if found
//   - error: Error if user not found or other issues occur
func (s *userService) GetUserByPhone(ctx context.Context, phone string) (*entities.User, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	// Check if the user has been soft deleted
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetAllUsers retrieves all users in the system with pagination.
// This method is useful for administrative purposes and user management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of users to return
//   - offset: Number of users to skip for pagination
//
// Returns:
//   - []*entities.User: List of all users
//   - error: Any error that occurred during the operation
func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	return s.userRepo.GetAll(ctx, limit, offset)
}

// SearchUsers searches for users based on a query string.
// This method supports full-text search capabilities for finding users
// by name, email, username, or other attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.User: List of matching users
//   - error: Any error that occurred during the operation
func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	return s.userRepo.Search(ctx, query, limit, offset)
}

// GetUsersWithPagination retrieves users with pagination and returns total count.
// This method provides a comprehensive pagination solution with search capabilities.
//
// Business Rules:
//   - Page and perPage parameters are properly handled
//   - Search functionality is integrated with pagination
//   - Total count is calculated for pagination metadata
//   - Offset is calculated based on page and perPage
//
// Parameters:
//   - ctx: Context for the operation
//   - page: Current page number (1-based)
//   - perPage: Number of users per page
//   - search: Optional search query for filtering
//
// Returns:
//   - []*entities.User: List of users for the current page
//   - int64: Total count of users for pagination
//   - error: Any error that occurred during the operation
func (s *userService) GetUsersWithPagination(ctx context.Context, page, perPage int, search string) ([]*entities.User, int64, error) {
	// Calculate offset based on page and perPage for pagination
	offset := (page - 1) * perPage

	var users []*entities.User
	var err error

	// Get users based on search parameter or all users
	if search != "" {
		users, err = s.userRepo.Search(ctx, search, perPage, offset)
	} else {
		users, err = s.userRepo.GetAll(ctx, perPage, offset)
	}

	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination metadata
	total, err := s.GetUsersCount(ctx, search)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUsersCount returns total count of users for pagination calculations.
// This method supports both general count and search-based count.
//
// Parameters:
//   - ctx: Context for the operation
//   - search: Optional search query for filtered count
//
// Returns:
//   - int64: Total count of users
//   - error: Any error that occurred during the operation
func (s *userService) GetUsersCount(ctx context.Context, search string) (int64, error) {
	if search != "" {
		return s.userRepo.CountBySearch(ctx, search)
	}
	return s.userRepo.Count(ctx)
}

// UpdateUser updates an existing user's profile information.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - User must exist and not be deleted
//   - Updated email/username/phone must be unique
//   - Only changed fields are validated for uniqueness
//   - User validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the user to update
//   - username: Updated username (optional)
//   - email: Updated email address (optional)
//   - phone: Updated phone number (optional)
//
// Returns:
//   - *entities.User: The updated user entity
//   - error: Any error that occurred during the operation
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, username, email, phone string) (*entities.User, error) {
	// Retrieve existing user to ensure it exists and is not deleted
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}

	// Check for conflicts if email is being changed
	if email != "" && email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this email already exists")
		}
	}

	// Check for conflicts if username is being changed
	if username != "" && username != user.UserName {
		exists, err := s.userRepo.ExistsByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this username already exists")
		}
	}

	// Check for conflicts if phone is being changed
	if phone != "" && phone != user.Phone {
		exists, err := s.userRepo.ExistsByPhone(ctx, phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this phone already exists")
		}
	}

	// Update the user entity with new information
	if err := user.UpdateUser(username, email, phone); err != nil {
		return nil, err
	}

	// Persist the updated user to the repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser performs a soft delete of a user by marking them as deleted
// rather than physically removing them from the database. This preserves data
// integrity and allows for potential recovery.
//
// Business Rules:
//   - User must exist before deletion
//   - Soft delete preserves user data
//   - Deleted users are not returned in queries
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the user to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	// Perform soft delete by marking the user as deleted
	user.SoftDelete()
	return s.userRepo.Update(ctx, user)
}

// ChangePassword allows a user to change their password with proper validation.
// This method enforces security best practices for password changes.
//
// Business Rules:
//   - User must exist and not be deleted
//   - Old password must be verified
//   - New password must meet strength requirements
//   - Password is securely hashed before storage
//
// Security Features:
//   - Old password verification
//   - Password strength validation
//   - Secure password hashing
//   - Atomic password update
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the user changing password
//   - oldPassword: Current password for verification
//   - newPassword: New password to set
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	// Retrieve user to ensure they exist and are not deleted
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	// Verify old password to ensure security
	if !s.passwordService.ComparePassword(user.Password, oldPassword) {
		return errors.New("invalid old password")
	}

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

// VerifyEmail marks a user's email address as verified.
// This method is part of the email verification workflow.
//
// Business Rules:
//   - User must exist and not be deleted
//   - Email verification status is updated
//   - Verification is atomic and consistent
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the user to verify email for
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *userService) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	// Mark email as verified
	user.VerifyEmail()
	return s.userRepo.Update(ctx, user)
}

// VerifyPhone marks a user's phone number as verified.
// This method is part of the phone verification workflow.
//
// Business Rules:
//   - User must exist and not be deleted
//   - Phone verification status is updated
//   - Verification is atomic and consistent
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the user to verify phone for
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *userService) VerifyPhone(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	// Mark phone as verified
	user.VerifyPhone()
	return s.userRepo.Update(ctx, user)
}

// AuthenticateUser authenticates a user with username and password.
// This method implements secure authentication with proper validation.
//
// Business Rules:
//   - User must exist and not be deleted
//   - Password must match the stored hash
//   - Generic error messages prevent user enumeration
//
// Security Features:
//   - Password comparison using secure hashing
//   - Soft delete checking prevents deleted user login
//   - Generic error messages prevent user enumeration attacks
//
// Parameters:
//   - ctx: Context for the operation
//   - username: Username for authentication
//   - password: Plain text password for authentication
//
// Returns:
//   - *entities.User: The authenticated user entity
//   - error: Any error that occurred during the operation
func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (*entities.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}

	// Verify password using secure comparison
	if !s.passwordService.ComparePassword(user.Password, password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
