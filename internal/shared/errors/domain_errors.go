package errors

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to the domain error
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	e.Details[key] = value
	return e
}

// Common domain error codes
const (
	// Validation errors
	ValidationErrorCode       = "VALIDATION_ERROR"
	InvalidEmailErrorCode     = "INVALID_EMAIL"
	InvalidPhoneErrorCode     = "INVALID_PHONE"
	InvalidPasswordErrorCode  = "INVALID_PASSWORD"
	RequiredFieldErrorCode    = "REQUIRED_FIELD"

	// Business rule errors
	BusinessRuleErrorCode     = "BUSINESS_RULE_VIOLATION"
	EmailAlreadyExistsCode    = "EMAIL_ALREADY_EXISTS"
	UsernameAlreadyExistsCode = "USERNAME_ALREADY_EXISTS"
	PhoneAlreadyExistsCode    = "PHONE_ALREADY_EXISTS"
	EmailAlreadyVerifiedCode  = "EMAIL_ALREADY_VERIFIED"
	PhoneAlreadyVerifiedCode  = "PHONE_ALREADY_VERIFIED"
	RoleAlreadyAssignedCode   = "ROLE_ALREADY_ASSIGNED"

	// Not found errors
	NotFoundErrorCode         = "NOT_FOUND"
	UserNotFoundCode          = "USER_NOT_FOUND"
	RoleNotFoundCode          = "ROLE_NOT_FOUND"

	// Authorization errors
	UnauthorizedErrorCode     = "UNAUTHORIZED"
	ForbiddenErrorCode        = "FORBIDDEN"
	InvalidCredentialsCode    = "INVALID_CREDENTIALS"

	// Concurrency errors
	ConcurrencyErrorCode      = "CONCURRENCY_ERROR"
	OptimisticLockErrorCode   = "OPTIMISTIC_LOCK_ERROR"
)

// Validation errors
var (
	ErrValidation       = NewDomainError(ValidationErrorCode, "Validation failed")
	ErrInvalidEmail     = NewDomainError(InvalidEmailErrorCode, "Invalid email format")
	ErrInvalidPhone     = NewDomainError(InvalidPhoneErrorCode, "Invalid phone format")
	ErrInvalidPassword  = NewDomainError(InvalidPasswordErrorCode, "Invalid password")
	ErrRequiredField    = func(field string) *DomainError {
		return NewDomainError(RequiredFieldErrorCode, fmt.Sprintf("%s is required", field))
	}
)

// Business rule errors
var (
	ErrBusinessRule        = NewDomainError(BusinessRuleErrorCode, "Business rule violation")
	ErrEmailAlreadyExists  = NewDomainError(EmailAlreadyExistsCode, "Email already exists")
	ErrUsernameAlreadyExists = NewDomainError(UsernameAlreadyExistsCode, "Username already exists")
	ErrPhoneAlreadyExists  = NewDomainError(PhoneAlreadyExistsCode, "Phone already exists")
	ErrEmailAlreadyVerified = NewDomainError(EmailAlreadyVerifiedCode, "Email is already verified")
	ErrPhoneAlreadyVerified = NewDomainError(PhoneAlreadyVerifiedCode, "Phone is already verified")
	ErrRoleAlreadyAssigned = NewDomainError(RoleAlreadyAssignedCode, "Role is already assigned to user")
)

// Not found errors
var (
	ErrNotFound    = NewDomainError(NotFoundErrorCode, "Resource not found")
	ErrUserNotFound = func(id uuid.UUID) *DomainError {
		return NewDomainError(UserNotFoundCode, "User not found").WithDetails("user_id", id)
	}
	ErrRoleNotFound = func(id uuid.UUID) *DomainError {
		return NewDomainError(RoleNotFoundCode, "Role not found").WithDetails("role_id", id)
	}
)

// Authorization errors
var (
	ErrUnauthorized      = NewDomainError(UnauthorizedErrorCode, "Unauthorized access")
	ErrForbidden         = NewDomainError(ForbiddenErrorCode, "Forbidden access")
	ErrInvalidCredentials = NewDomainError(InvalidCredentialsCode, "Invalid credentials")
)

// Concurrency errors
var (
	ErrConcurrency      = NewDomainError(ConcurrencyErrorCode, "Concurrency error")
	ErrOptimisticLock   = NewDomainError(OptimisticLockErrorCode, "Optimistic lock error")
)

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}

// GetDomainError extracts a domain error from an error
func GetDomainError(err error) *DomainError {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}
	return nil
}