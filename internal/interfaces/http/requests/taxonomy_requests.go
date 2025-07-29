package requests

import (
	"github.com/go-playground/validator/v10"
)

// CreateTaxonomyRequest represents the request for creating a taxonomy
type CreateTaxonomyRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Slug        string `json:"slug" validate:"required,min=1,max=255"`
	Code        string `json:"code,omitempty" validate:"max=50"`
	Description string `json:"description,omitempty" validate:"max=1000"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
}

// Validate validates the CreateTaxonomyRequest
func (r *CreateTaxonomyRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// UpdateTaxonomyRequest represents the request for updating a taxonomy
type UpdateTaxonomyRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Slug        string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	Code        string `json:"code,omitempty" validate:"max=50"`
	Description string `json:"description,omitempty" validate:"max=1000"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
}

// Validate validates the UpdateTaxonomyRequest
func (r *UpdateTaxonomyRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
