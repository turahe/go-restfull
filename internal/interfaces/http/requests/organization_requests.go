// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateOrganizationRequest represents the request for creating a new organization entity.
// This struct defines the required and optional fields for organization creation,
// including validation tags for field constraints and business rules.
// The request supports hierarchical organization structures through parent_id.
type CreateOrganizationRequest struct {
	// Name is the organization's display name (required, 1-255 characters)
	Name string `json:"name" validate:"required,min=1,max=255"`
	// Description provides additional details about the organization (optional, max 1000 characters)
	Description string `json:"description,omitempty" validate:"max=1000"`
	// Code is a unique identifier for the organization (optional, max 50 characters, alphanumeric + hyphens)
	Code string `json:"code,omitempty" validate:"max=50"`
	// Email is the organization's contact email address (optional, must be valid email format if provided)
	Email string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	// Phone is the organization's contact phone number (optional, max 50 characters)
	Phone string `json:"phone,omitempty" validate:"max=50"`
	// Address is the organization's physical address (optional, max 500 characters)
	Address string `json:"address,omitempty" validate:"max=500"`
	// Website is the organization's website URL (optional, must be valid URL format if provided)
	Website string `json:"website,omitempty" validate:"omitempty,url,max=255"`
	// LogoURL is the URL to the organization's logo image (optional, must be valid URL format if provided)
	LogoURL string `json:"logo_url,omitempty" validate:"omitempty,url,max=500"`
	// ParentID is the UUID of the parent organization for hierarchical structures (optional, must be valid UUID if provided)
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
}

// Validate performs validation on the CreateOrganizationRequest using the validator package.
// This method checks all field constraints including required fields, length limits,
// and custom business rules for the organization code format.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateOrganizationRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Custom validation for code format (alphanumeric and hyphens only)
	if r.Code != "" {
		if !isValidCode(r.Code) {
			return errors.New("code must contain only alphanumeric characters and hyphens")
		}
	}

	return nil
}

// ToEntity transforms the CreateOrganizationRequest to an Organization domain entity.
// This method parses the parent_id string to UUID if provided, handles optional fields,
// and sets default values for required entity fields like status.
//
// Returns:
//   - *entities.Organization: The created organization entity
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *CreateOrganizationRequest) ToEntity() (*entities.Organization, error) {
	// Parse parent_id string to UUID if provided
	var parentID *uuid.UUID
	if r.ParentID != "" {
		parsedID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &parsedID
	}

	// Handle optional description field
	var description *string
	if r.Description != "" {
		description = &r.Description
	}

	// Handle optional code field
	var code *string
	if r.Code != "" {
		code = &r.Code
	}

	// Create and populate the organization entity
	organization := &entities.Organization{
		ID:          uuid.New(),
		Name:        r.Name,
		Description: description,
		Code:        code,
		Status:      entities.OrganizationStatusActive, // Default status
		ParentID:    parentID,
	}

	return organization, nil
}

// UpdateOrganizationRequest represents the request for updating an existing organization entity.
// This struct uses omitempty tags to make all fields optional, allowing partial updates.
// Only provided fields will be updated in the existing organization entity.
type UpdateOrganizationRequest struct {
	// Name is the organization's display name (optional, 1-255 characters if provided)
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Description provides additional details about the organization (optional, max 1000 characters if provided)
	Description string `json:"description,omitempty" validate:"max=1000"`
	// Code is a unique identifier for the organization (optional, max 50 characters, alphanumeric + hyphens if provided)
	Code string `json:"code,omitempty" validate:"max=50"`
	// Email is the organization's contact email address (optional, must be valid email format if provided)
	Email string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	// Phone is the organization's contact phone number (optional, max 50 characters if provided)
	Phone string `json:"phone,omitempty" validate:"max=50"`
	// Address is the organization's physical address (optional, max 500 characters if provided)
	Address string `json:"address,omitempty" validate:"max=500"`
	// Website is the organization's website URL (optional, must be valid URL format if provided)
	Website string `json:"website,omitempty" validate:"omitempty,url,max=255"`
	// LogoURL is the URL to the organization's logo image (optional, must be valid URL format if provided)
	LogoURL string `json:"logo_url,omitempty" validate:"omitempty,url,max=500"`
}

// Validate performs validation on the UpdateOrganizationRequest using the validator package.
// This method checks field constraints for any provided fields and ensures at least
// one field is provided for the update operation.
//
// Returns:
//   - error: Validation error if any provided field fails validation or no fields provided, nil if valid
func (r *UpdateOrganizationRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Check if at least one field is provided for update
	if r.Name == "" && r.Description == "" && r.Code == "" &&
		r.Email == "" && r.Phone == "" && r.Address == "" &&
		r.Website == "" && r.LogoURL == "" {
		return errors.New("at least one field must be provided for update")
	}

	// Custom validation for code format (alphanumeric and hyphens only)
	if r.Code != "" {
		if !isValidCode(r.Code) {
			return errors.New("code must contain only alphanumeric characters and hyphens")
		}
	}

	return nil
}

// ToEntity transforms the UpdateOrganizationRequest to update an existing Organization entity.
// This method applies only the provided fields to the existing organization, preserving
// unchanged values. It's designed for partial updates where not all fields are provided.
//
// Parameters:
//   - existingOrganization: The existing organization entity to update
//
// Returns:
//   - *entities.Organization: The updated organization entity
//   - error: Any error that occurred during transformation
func (r *UpdateOrganizationRequest) ToEntity(existingOrganization *entities.Organization) (*entities.Organization, error) {
	// Update fields only if provided in the request
	if r.Name != "" {
		existingOrganization.Name = r.Name
	}
	if r.Description != "" {
		existingOrganization.Description = &r.Description
	}
	if r.Code != "" {
		existingOrganization.Code = &r.Code
	}

	return existingOrganization, nil
}

// MoveOrganizationRequest represents the request for moving an organization within the hierarchy.
// This request allows changing an organization's parent, which affects its position
// in the organizational tree structure.
type MoveOrganizationRequest struct {
	// NewParentID is the UUID of the new parent organization (required, must be valid UUID)
	NewParentID string `json:"new_parent_id" validate:"required,uuid4"`
}

// Validate performs validation on the MoveOrganizationRequest using the validator package.
// This method ensures the new_parent_id is a valid UUID format.
//
// Returns:
//   - error: Validation error if the new_parent_id is invalid, nil if valid
func (r *MoveOrganizationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the MoveOrganizationRequest to update an existing Organization entity.
// This method parses the new_parent_id string to UUID and updates the organization's
// parent relationship.
//
// Parameters:
//   - existingOrganization: The existing organization entity to move
//
// Returns:
//   - *entities.Organization: The organization entity with updated parent
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *MoveOrganizationRequest) ToEntity(existingOrganization *entities.Organization) (*entities.Organization, error) {
	// Parse the new parent ID string to UUID
	newParentID, err := uuid.Parse(r.NewParentID)
	if err != nil {
		return nil, err
	}

	// Update the organization's parent
	existingOrganization.ParentID = &newParentID
	return existingOrganization, nil
}

// SetOrganizationStatusRequest represents the request for changing an organization's status.
// This request allows updating the organization's operational state (active, inactive, suspended)
// which may affect its visibility and functionality in the system.
type SetOrganizationStatusRequest struct {
	// Status is the new status for the organization (required, must be valid enum value)
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
}

// Validate performs validation on the SetOrganizationStatusRequest using the validator package.
// This method ensures the status is one of the allowed enumerated values.
//
// Returns:
//   - error: Validation error if the status is invalid, nil if valid
func (r *SetOrganizationStatusRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the SetOrganizationStatusRequest to update an existing Organization entity.
// This method updates the organization's status field directly.
//
// Parameters:
//   - existingOrganization: The existing organization entity to update
//
// Returns:
//   - *entities.Organization: The organization entity with updated status
func (r *SetOrganizationStatusRequest) ToEntity(existingOrganization *entities.Organization) *entities.Organization {
	existingOrganization.Status = entities.OrganizationStatus(r.Status)
	return existingOrganization
}

// SearchOrganizationsRequest represents the request for searching organizations by query.
// This struct supports text-based search with pagination parameters for result management.
type SearchOrganizationsRequest struct {
	// Query is the search term to match against organization names and descriptions (required, 1-100 characters)
	Query string `json:"query" validate:"required,min=1,max=100"`
	// Page is the page number for pagination (optional, minimum 1, defaults to 1)
	Page int `json:"page,omitempty" validate:"omitempty,min=1"`
	// PerPage is the number of results per page (optional, 1-100, defaults to 10)
	PerPage int `json:"per_page,omitempty" validate:"omitempty,min=1,max=100"`
}

// Validate performs validation on the SearchOrganizationsRequest using the validator package.
// This method checks field constraints and sets sensible defaults for pagination
// parameters to ensure consistent search behavior.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *SearchOrganizationsRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Set default pagination values for consistent behavior
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PerPage == 0 {
		r.PerPage = 10
	}

	return nil
}

// GetOrganizationsRequest represents the request for retrieving organizations with filters.
// This struct supports filtering by status and search terms, with pagination parameters
// for result management.
type GetOrganizationsRequest struct {
	// Page is the page number for pagination (optional, minimum 1, defaults to 1)
	Page int `json:"page,omitempty" validate:"omitempty,min=1"`
	// PerPage is the number of results per page (optional, 1-100, defaults to 10)
	PerPage int `json:"per_page,omitempty" validate:"omitempty,min=1,max=100"`
	// Search is an optional search term for filtering organizations (optional, max 100 characters)
	Search string `json:"search,omitempty" validate:"max=100"`
	// Status filters organizations by their operational status (optional, must be valid enum if provided)
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended"`
}

// Validate performs validation on the GetOrganizationsRequest using the validator package.
// This method checks field constraints and sets sensible defaults for pagination
// parameters to ensure consistent retrieval behavior.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *GetOrganizationsRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Set default pagination values for consistent behavior
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PerPage == 0 {
		r.PerPage = 10
	}

	return nil
}

// isValidCode checks if the organization code contains only alphanumeric characters and hyphens.
// This helper function enforces business rules for organization codes, ensuring they
// are URL-safe and consistent with naming conventions.
//
// Business Rules:
// - Code cannot be empty
// - Code cannot start or end with a hyphen
// - Code can only contain letters (a-z, A-Z), numbers (0-9), and hyphens
//
// Parameters:
//   - code: The organization code string to validate
//
// Returns:
//   - bool: True if the code meets all business rules, false otherwise
func isValidCode(code string) bool {
	if code == "" {
		return false
	}

	// Check if code starts or ends with hyphen (not allowed)
	if strings.HasPrefix(code, "-") || strings.HasSuffix(code, "-") {
		return false
	}

	// Check if code contains only alphanumeric characters and hyphens
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return false
		}
	}

	return true
}
