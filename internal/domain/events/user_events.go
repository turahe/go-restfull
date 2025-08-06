package events

import (
	"github.com/google/uuid"
)

// User event types
const (
	UserCreatedEventType        = "user.created"
	UserEmailVerifiedEventType  = "user.email_verified"
	UserPhoneVerifiedEventType  = "user.phone_verified"
	UserPasswordChangedEventType = "user.password_changed"
	UserRoleAssignedEventType   = "user.role_assigned"
	UserRoleRemovedEventType    = "user.role_removed"
	UserProfileUpdatedEventType = "user.profile_updated"
	UserDeletedEventType        = "user.deleted"
)

// UserCreatedEvent represents a user created event
type UserCreatedEvent struct {
	BaseDomainEvent
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID uuid.UUID, username, email string) DomainEvent {
	eventData := map[string]interface{}{
		"username": username,
		"email":    email,
	}
	
	return UserCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserCreatedEventType, userID, "user", eventData),
	}
}

// UserEmailVerifiedEvent represents a user email verified event
type UserEmailVerifiedEvent struct {
	BaseDomainEvent
}

// NewUserEmailVerifiedEvent creates a new user email verified event
func NewUserEmailVerifiedEvent(userID uuid.UUID, email string) DomainEvent {
	eventData := map[string]interface{}{
		"email": email,
	}
	
	return UserEmailVerifiedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserEmailVerifiedEventType, userID, "user", eventData),
	}
}

// UserPhoneVerifiedEvent represents a user phone verified event
type UserPhoneVerifiedEvent struct {
	BaseDomainEvent
}

// NewUserPhoneVerifiedEvent creates a new user phone verified event
func NewUserPhoneVerifiedEvent(userID uuid.UUID, phone string) DomainEvent {
	eventData := map[string]interface{}{
		"phone": phone,
	}
	
	return UserPhoneVerifiedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserPhoneVerifiedEventType, userID, "user", eventData),
	}
}

// UserPasswordChangedEvent represents a user password changed event
type UserPasswordChangedEvent struct {
	BaseDomainEvent
}

// NewUserPasswordChangedEvent creates a new user password changed event
func NewUserPasswordChangedEvent(userID uuid.UUID) DomainEvent {
	eventData := map[string]interface{}{}
	
	return UserPasswordChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserPasswordChangedEventType, userID, "user", eventData),
	}
}

// UserRoleAssignedEvent represents a user role assigned event
type UserRoleAssignedEvent struct {
	BaseDomainEvent
}

// NewUserRoleAssignedEvent creates a new user role assigned event
func NewUserRoleAssignedEvent(userID, roleID uuid.UUID, roleName string) DomainEvent {
	eventData := map[string]interface{}{
		"role_id":   roleID,
		"role_name": roleName,
	}
	
	return UserRoleAssignedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserRoleAssignedEventType, userID, "user", eventData),
	}
}

// UserRoleRemovedEvent represents a user role removed event
type UserRoleRemovedEvent struct {
	BaseDomainEvent
}

// NewUserRoleRemovedEvent creates a new user role removed event
func NewUserRoleRemovedEvent(userID, roleID uuid.UUID, roleName string) DomainEvent {
	eventData := map[string]interface{}{
		"role_id":   roleID,
		"role_name": roleName,
	}
	
	return UserRoleRemovedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserRoleRemovedEventType, userID, "user", eventData),
	}
}

// UserProfileUpdatedEvent represents a user profile updated event
type UserProfileUpdatedEvent struct {
	BaseDomainEvent
}

// NewUserProfileUpdatedEvent creates a new user profile updated event
func NewUserProfileUpdatedEvent(userID uuid.UUID) DomainEvent {
	eventData := map[string]interface{}{}
	
	return UserProfileUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserProfileUpdatedEventType, userID, "user", eventData),
	}
}

// UserDeletedEvent represents a user deleted event
type UserDeletedEvent struct {
	BaseDomainEvent
}

// NewUserDeletedEvent creates a new user deleted event
func NewUserDeletedEvent(userID uuid.UUID, username string) DomainEvent {
	eventData := map[string]interface{}{
		"username": username,
	}
	
	return UserDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(UserDeletedEventType, userID, "user", eventData),
	}
}