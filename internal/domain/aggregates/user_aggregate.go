package aggregates

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/events"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
)

// UserAggregate represents the user aggregate root
type UserAggregate struct {
	ID              uuid.UUID                   `json:"id"`
	UserName        string                      `json:"username"`
	Email           valueobjects.Email          `json:"email"`
	Phone           valueobjects.Phone          `json:"phone"`
	Password        valueobjects.HashedPassword `json:"-"`
	EmailVerifiedAt *time.Time                  `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time                  `json:"phone_verified_at,omitempty"`
	Roles           []valueobjects.Role         `json:"roles,omitempty"`
	Profile         *valueobjects.UserProfile   `json:"profile,omitempty"`
	Avatar          *string                     `json:"avatar,omitempty"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	DeletedAt       *time.Time                  `json:"deleted_at,omitempty"`
	Version         int                         `json:"version"`

	// Domain events
	events []events.DomainEvent
}

// NewUserAggregate creates a new user aggregate with validation
func NewUserAggregate(username string, email valueobjects.Email, phone valueobjects.Phone, password valueobjects.HashedPassword) (*UserAggregate, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	now := time.Now()
	user := &UserAggregate{
		ID:        uuid.New(),
		UserName:  username,
		Email:     email,
		Phone:     phone,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
		events:    make([]events.DomainEvent, 0),
	}

	// Add domain event
	user.addEvent(events.NewUserCreatedEvent(user.ID, user.UserName, user.Email.String()))

	return user, nil
}

// VerifyEmail verifies the user's email address
func (u *UserAggregate) VerifyEmail() error {
	if u.EmailVerifiedAt != nil {
		return errors.New("email is already verified")
	}

	now := time.Now()
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
	u.Version++

	u.addEvent(events.NewUserEmailVerifiedEvent(u.ID, u.Email.String()))
	return nil
}

// VerifyPhone verifies the user's phone number
func (u *UserAggregate) VerifyPhone() error {
	if u.PhoneVerifiedAt != nil {
		return errors.New("phone is already verified")
	}

	now := time.Now()
	u.PhoneVerifiedAt = &now
	u.UpdatedAt = now
	u.Version++

	u.addEvent(events.NewUserPhoneVerifiedEvent(u.ID, u.Phone.String()))
	return nil
}

// ChangePassword changes the user's password
func (u *UserAggregate) ChangePassword(newPassword valueobjects.HashedPassword) error {
	u.Password = newPassword
	u.UpdatedAt = time.Now()
	u.Version++

	u.addEvent(events.NewUserPasswordChangedEvent(u.ID))
	return nil
}

// AssignRole assigns a role to the user
func (u *UserAggregate) AssignRole(role valueobjects.Role) error {
	// Check if role already exists
	for _, existingRole := range u.Roles {
		if existingRole.ID == role.ID {
			return errors.New("role already assigned to user")
		}
	}

	u.Roles = append(u.Roles, role)
	u.UpdatedAt = time.Now()
	u.Version++

	u.addEvent(events.NewUserRoleAssignedEvent(u.ID, role.ID, role.Name))
	return nil
}

// RemoveRole removes a role from the user
func (u *UserAggregate) RemoveRole(roleID uuid.UUID) error {
	for i, role := range u.Roles {
		if role.ID == roleID {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			u.UpdatedAt = time.Now()
			u.Version++

			u.addEvent(events.NewUserRoleRemovedEvent(u.ID, roleID, role.Name))
			return nil
		}
	}

	return errors.New("role not found")
}

// UpdateProfile updates the user's profile
func (u *UserAggregate) UpdateProfile(profile valueobjects.UserProfile) error {
	u.Profile = &profile
	u.UpdatedAt = time.Now()
	u.Version++

	u.addEvent(events.NewUserProfileUpdatedEvent(u.ID))
	return nil
}

// SoftDelete soft deletes the user
func (u *UserAggregate) SoftDelete() error {
	if u.DeletedAt != nil {
		return errors.New("user is already deleted")
	}

	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
	u.Version++

	u.addEvent(events.NewUserDeletedEvent(u.ID, u.UserName))
	return nil
}

// IsDeleted checks if the user is deleted
func (u *UserAggregate) IsDeleted() bool {
	return u.DeletedAt != nil
}

// IsEmailVerified checks if the user's email is verified
func (u *UserAggregate) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsPhoneVerified checks if the user's phone is verified
func (u *UserAggregate) IsPhoneVerified() bool {
	return u.PhoneVerifiedAt != nil
}

// HasRole checks if the user has a specific role
func (u *UserAggregate) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// GetEvents returns the domain events
func (u *UserAggregate) GetEvents() []events.DomainEvent {
	return u.events
}

// ClearEvents clears the domain events
func (u *UserAggregate) ClearEvents() {
	u.events = make([]events.DomainEvent, 0)
}

// addEvent adds a domain event
func (u *UserAggregate) addEvent(event events.DomainEvent) {
	u.events = append(u.events, event)
}
