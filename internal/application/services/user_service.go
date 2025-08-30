// Package services provides application-level business logic for user management.
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	domainservices "github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"

	"github.com/google/uuid"
)

type userService struct {
	userRepo        repositories.UserRepository
	passwordService domainservices.PasswordService
	emailService    domainservices.EmailService
	mediaService    ports.MediaService
}

func NewUserService(
	userRepo repositories.UserRepository,
	passwordService domainservices.PasswordService,
	emailService domainservices.EmailService,
	mediaService ports.MediaService,
) ports.UserService {
	return &userService{
		userRepo:        userRepo,
		passwordService: passwordService,
		emailService:    emailService,
		mediaService:    mediaService,
	}
}

// Helpers
func aggregateToEntity(u *aggregates.UserAggregate) *entities.User {
	return &entities.User{
		ID:              u.ID,
		UserName:        u.UserName,
		Email:           u.Email.String(),
		Phone:           u.Phone.String(),
		Password:        u.Password.Hash(),
		Avatar:          "", // Will be populated by service methods
		EmailVerifiedAt: u.EmailVerifiedAt,
		PhoneVerifiedAt: u.PhoneVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		DeletedAt:       u.DeletedAt,
	}
}

// aggregateToEntityWithAvatar converts a user aggregate to a user entity and fetches the avatar
func (s *userService) aggregateToEntityWithAvatar(ctx context.Context, u *aggregates.UserAggregate) *entities.User {
	user := aggregateToEntity(u)

	// Fetch avatar if media service is available
	if s.mediaService != nil {
		media, err := s.mediaService.GetAvatarByUserID(ctx, u.ID)
		if err == nil && media != nil {
			user.Avatar = media.GetURL()
		}
	}

	return user
}

// GetUserMediaByGroup retrieves media by group for a specific user
func (s *userService) GetUserMediaByGroup(ctx context.Context, userID uuid.UUID, group string) (*entities.Media, error) {
	if s.mediaService == nil {
		return nil, fmt.Errorf("media service not available")
	}
	return s.mediaService.GetMediaByGroup(ctx, userID, "User", group)
}

// GetUserMediaGallery retrieves all media in a specific group for a user
func (s *userService) GetUserMediaGallery(ctx context.Context, userID uuid.UUID, group string, limit, offset int) ([]*entities.Media, error) {
	if s.mediaService == nil {
		return nil, fmt.Errorf("media service not available")
	}
	return s.mediaService.GetAllMediaByGroup(ctx, userID, "User", group, limit, offset)
}

func (s *userService) CreateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	// Validate password, uniqueness
	if err := s.passwordService.ValidatePassword(user.Password); err != nil {
		return nil, err
	}
	if exists, err := s.userRepo.ExistsByEmail(ctx, user.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("user with this email already exists")
	}
	if exists, err := s.userRepo.ExistsByUsername(ctx, user.UserName); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("user with this username already exists")
	}
	if exists, err := s.userRepo.ExistsByPhone(ctx, user.Phone); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("user with this phone already exists")
	}

	// Build aggregate
	emailVO, err := valueobjects.NewEmail(user.Email)
	if err != nil {
		return nil, err
	}
	phoneVO, err := valueobjects.NewPhone(user.Phone)
	if err != nil {
		return nil, err
	}
	hashed, err := s.passwordService.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	passVO, err := valueobjects.NewHashedPasswordFromHash(hashed)
	if err != nil {
		return nil, err
	}
	agg, err := aggregates.NewUserAggregate(user.UserName, emailVO, phoneVO, passVO)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Save(ctx, agg); err != nil {
		return nil, err
	}

	// Fire-and-forget welcome email
	go func(to, name string) { _ = s.emailService.SendWelcomeEmail(to, name) }(user.Email, user.UserName)

	return aggregateToEntity(agg), nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	agg, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if agg.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return s.aggregateToEntityWithAvatar(ctx, agg), nil
}

// GetUserProfileWithRelations retrieves a user by ID with roles and menus populated
func (s *userService) GetUserProfileWithRelations(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	// First get the basic user
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch user roles
	roles, err := s.userRepo.GetUserRoles(ctx, id)
	if err != nil {
		// Log error but don't fail the request
		// You might want to add logging here
		roles = []*entities.Role{}
	}
	user.Roles = roles

	// Fetch user menus
	menus, err := s.userRepo.GetUserMenus(ctx, id)
	if err != nil {
		// Log error but don't fail the request
		// You might want to add logging here
		menus = []*entities.Menu{}
	}
	user.Menus = menus

	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	agg, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if agg.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return s.aggregateToEntityWithAvatar(ctx, agg), nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	agg, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if agg.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return s.aggregateToEntityWithAvatar(ctx, agg), nil
}

func (s *userService) GetUserByPhone(ctx context.Context, phone string) (*entities.User, error) {
	agg, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if agg.IsDeleted() {
		return nil, errors.New("user not found")
	}
	return s.aggregateToEntityWithAvatar(ctx, agg), nil
}

func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	// Convert to page/pageSize
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10
	}
	page := 1
	if pageSize > 0 {
		page = offset/pageSize + 1
	}
	res, err := s.userRepo.FindAll(ctx, queries.ListUsersQuery{Page: page, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	out := make([]*entities.User, 0, len(res.Items))
	for _, agg := range res.Items {
		out = append(out, aggregateToEntity(agg))
	}
	return out, nil
}

func (s *userService) SearchUsers(ctx context.Context, q string, limit, offset int) ([]*entities.User, error) {
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10
	}
	page := 1
	if pageSize > 0 {
		page = offset/pageSize + 1
	}
	res, err := s.userRepo.Search(ctx, queries.SearchUsersQuery{Query: q, Page: page, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	out := make([]*entities.User, 0, len(res.Items))
	for _, agg := range res.Items {
		out = append(out, aggregateToEntity(agg))
	}
	return out, nil
}

func (s *userService) GetUsersWithPagination(ctx context.Context, page, perPage int, search string) ([]*entities.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}
	if search != "" {
		res, err := s.userRepo.Search(ctx, queries.SearchUsersQuery{Query: search, Page: page, PageSize: perPage})
		if err != nil {
			return nil, 0, err
		}
		out := make([]*entities.User, 0, len(res.Items))
		for _, agg := range res.Items {
			out = append(out, aggregateToEntity(agg))
		}
		return out, int64(res.TotalCount), nil
	}
	res, err := s.userRepo.FindAll(ctx, queries.ListUsersQuery{Page: page, PageSize: perPage})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*entities.User, 0, len(res.Items))
	for _, agg := range res.Items {
		out = append(out, aggregateToEntity(agg))
	}
	return out, int64(res.TotalCount), nil
}

func (s *userService) GetUsersCount(ctx context.Context, search string) (int64, error) {
	if search != "" {
		res, err := s.userRepo.Search(ctx, queries.SearchUsersQuery{Query: search, Page: 1, PageSize: 1})
		if err != nil {
			return 0, err
		}
		return int64(res.TotalCount), nil
	}
	return s.userRepo.Count(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	agg, err := s.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if agg.IsDeleted() {
		return nil, errors.New("user not found")
	}

	if user.UserName != "" {
		agg.UserName = user.UserName
	}
	if user.Email != "" {
		emailVO, err := valueobjects.NewEmail(user.Email)
		if err != nil {
			return nil, err
		}
		agg.Email = emailVO
	}
	if user.Phone != "" {
		phoneVO, err := valueobjects.NewPhone(user.Phone)
		if err != nil {
			return nil, err
		}
		agg.Phone = phoneVO
	}
	agg.UpdatedAt = time.Now()

	if err := s.userRepo.Save(ctx, agg); err != nil {
		return nil, err
	}
	return aggregateToEntity(agg), nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	agg, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if agg.IsDeleted() {
		return errors.New("user not found")
	}
	now := time.Now()
	agg.DeletedAt = &now
	agg.UpdatedAt = now
	return s.userRepo.Save(ctx, agg)
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	agg, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if agg.IsDeleted() {
		return errors.New("user not found")
	}
	if !s.passwordService.ComparePassword(agg.Password.Hash(), oldPassword) {
		return errors.New("invalid old password")
	}
	if err := s.passwordService.ValidatePassword(newPassword); err != nil {
		return err
	}
	newHash, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}
	newVO, err := valueobjects.NewHashedPasswordFromHash(newHash)
	if err != nil {
		return err
	}
	if err := agg.ChangePassword(newVO); err != nil {
		return err
	}
	return s.userRepo.Save(ctx, agg)
}

func (s *userService) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	agg, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if agg.IsDeleted() {
		return errors.New("user not found")
	}
	if err := agg.VerifyEmail(); err != nil {
		return err
	}
	return s.userRepo.Save(ctx, agg)
}

func (s *userService) VerifyPhone(ctx context.Context, id uuid.UUID) error {
	agg, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if agg.IsDeleted() {
		return errors.New("user not found")
	}
	if err := agg.VerifyPhone(); err != nil {
		return err
	}
	return s.userRepo.Save(ctx, agg)
}

func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (*entities.User, error) {
	agg, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if agg.IsDeleted() {
		return nil, errors.New("user not found")
	}
	if !s.passwordService.ComparePassword(agg.Password.Hash(), password) {
		return nil, errors.New("invalid credentials")
	}
	return aggregateToEntity(agg), nil
}
