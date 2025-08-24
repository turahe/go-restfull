package requests

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateOrganizationRequest represents the request for creating an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=1000"`
	Code        string `json:"code,omitempty" validate:"max=50"`
	Email       string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone       string `json:"phone,omitempty" validate:"max=50"`
	Address     string `json:"address,omitempty" validate:"max=500"`
	Website     string `json:"website,omitempty" validate:"omitempty,url,max=255"`
	LogoURL     string `json:"logo_url,omitempty" validate:"omitempty,url,max=500"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
}

// Validate validates the CreateOrganizationRequest
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

// ToEntity transforms CreateOrganizationRequest to an Organization entity
func (r *CreateOrganizationRequest) ToEntity() (*entities.Organization, error) {
	var parentID *uuid.UUID
	if r.ParentID != "" {
		parsedID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &parsedID
	}

	var description *string
	if r.Description != "" {
		description = &r.Description
	}

	var code *string
	if r.Code != "" {
		code = &r.Code
	}

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

// UpdateOrganizationRequest represents the request for updating an organization
type UpdateOrganizationRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=1000"`
	Code        string `json:"code,omitempty" validate:"max=50"`
	Email       string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone       string `json:"phone,omitempty" validate:"max=50"`
	Address     string `json:"address,omitempty" validate:"max=500"`
	Website     string `json:"website,omitempty" validate:"omitempty,url,max=255"`
	LogoURL     string `json:"logo_url,omitempty" validate:"omitempty,url,max=500"`
}

// Validate validates the UpdateOrganizationRequest
func (r *UpdateOrganizationRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Check if at least one field is provided
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

// ToEntity transforms UpdateOrganizationRequest to update an existing Organization entity
func (r *UpdateOrganizationRequest) ToEntity(existingOrganization *entities.Organization) (*entities.Organization, error) {
	// Update fields only if provided
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

// MoveOrganizationRequest represents the request for moving an organization
type MoveOrganizationRequest struct {
	NewParentID string `json:"new_parent_id" validate:"required,uuid4"`
}

// Validate validates the MoveOrganizationRequest
func (r *MoveOrganizationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms MoveOrganizationRequest to update an existing Organization entity
func (r *MoveOrganizationRequest) ToEntity(existingOrganization *entities.Organization) (*entities.Organization, error) {
	newParentID, err := uuid.Parse(r.NewParentID)
	if err != nil {
		return nil, err
	}

	existingOrganization.ParentID = &newParentID
	return existingOrganization, nil
}

// SetOrganizationStatusRequest represents the request for setting organization status
type SetOrganizationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
}

// Validate validates the SetOrganizationStatusRequest
func (r *SetOrganizationStatusRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms SetOrganizationStatusRequest to update an existing Organization entity
func (r *SetOrganizationStatusRequest) ToEntity(existingOrganization *entities.Organization) *entities.Organization {
	existingOrganization.Status = entities.OrganizationStatus(r.Status)
	return existingOrganization
}

// SearchOrganizationsRequest represents the request for searching organizations
type SearchOrganizationsRequest struct {
	Query   string `json:"query" validate:"required,min=1,max=100"`
	Page    int    `json:"page,omitempty" validate:"omitempty,min=1"`
	PerPage int    `json:"per_page,omitempty" validate:"omitempty,min=1,max=100"`
}

// Validate validates the SearchOrganizationsRequest
func (r *SearchOrganizationsRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Set default values
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PerPage == 0 {
		r.PerPage = 10
	}

	return nil
}

// GetOrganizationsRequest represents the request for getting organizations with filters
type GetOrganizationsRequest struct {
	Page    int    `json:"page,omitempty" validate:"omitempty,min=1"`
	PerPage int    `json:"per_page,omitempty" validate:"omitempty,min=1,max=100"`
	Search  string `json:"search,omitempty" validate:"max=100"`
	Status  string `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended"`
}

// Validate validates the GetOrganizationsRequest
func (r *GetOrganizationsRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Set default values
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PerPage == 0 {
		r.PerPage = 10
	}

	return nil
}

// isValidCode checks if the code contains only alphanumeric characters and hyphens
func isValidCode(code string) bool {
	if code == "" {
		return false
	}

	// Check if code starts or ends with hyphen
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
