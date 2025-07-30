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

// userService implements the UserService interface
type userService struct {
	userRepo        repositories.UserRepository
	passwordService services.PasswordService
	emailService    services.EmailService
}

// NewUserService creates a new user service instance
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

func (s *userService) CreateUser(ctx context.Context, username, email, phone, password string) (*entities.User, error) {
	// Validate password strength
	if err := s.passwordService.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this username already exists")
	}

	exists, err = s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this phone already exists")
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user, err := entities.NewUser(username, email, phone, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send welcome email (async)
	go func() {
		_ = s.emailService.SendWelcomeEmail(email, username)
	}()

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetUserByPhone(ctx context.Context, phone string) (*entities.User, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	return s.userRepo.GetAll(ctx, limit, offset)
}

func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	return s.userRepo.Search(ctx, query, limit, offset)
}

// GetUsersWithPagination retrieves users with pagination and returns total count
func (s *userService) GetUsersWithPagination(ctx context.Context, page, perPage int, search string) ([]*entities.User, int64, error) {
	// Calculate offset
	offset := (page - 1) * perPage

	var users []*entities.User
	var err error

	// Get users based on search parameter
	if search != "" {
		users, err = s.userRepo.Search(ctx, search, perPage, offset)
	} else {
		users, err = s.userRepo.GetAll(ctx, perPage, offset)
	}

	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := s.GetUsersCount(ctx, search)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUsersCount returns total count of users (for pagination)
func (s *userService) GetUsersCount(ctx context.Context, search string) (int64, error) {
	if search != "" {
		return s.userRepo.CountBySearch(ctx, search)
	}
	return s.userRepo.Count(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, username, email, phone string) (*entities.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}

	// Check for conflicts if email/username/phone is being changed
	if email != "" && email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this email already exists")
		}
	}

	if username != "" && username != user.UserName {
		exists, err := s.userRepo.ExistsByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this username already exists")
		}
	}

	if phone != "" && phone != user.Phone {
		exists, err := s.userRepo.ExistsByPhone(ctx, phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this phone already exists")
		}
	}

	// Update user
	if err := user.UpdateUser(username, email, phone); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	user.SoftDelete()
	return s.userRepo.Update(ctx, user)
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	// Verify old password
	if !s.passwordService.ComparePassword(user.Password, oldPassword) {
		return errors.New("invalid old password")
	}

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

func (s *userService) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	user.VerifyEmail()
	return s.userRepo.Update(ctx, user)
}

func (s *userService) VerifyPhone(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsDeleted() {
		return errors.New("user not found")
	}

	user.VerifyPhone()
	return s.userRepo.Update(ctx, user)
}

func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (*entities.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted() {
		return nil, errors.New("user not found")
	}

	if !s.passwordService.ComparePassword(user.Password, password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
