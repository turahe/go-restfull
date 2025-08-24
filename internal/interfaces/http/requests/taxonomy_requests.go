package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
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

// ToEntity transforms CreateTaxonomyRequest to a Taxonomy entity
func (r *CreateTaxonomyRequest) ToEntity() (*entities.Taxonomy, error) {
	var parentID *uuid.UUID
	if r.ParentID != "" {
		parsedID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &parsedID
	}

	taxonomy := &entities.Taxonomy{
		ID:          uuid.New(),
		Name:        r.Name,
		Slug:        r.Slug,
		Code:        r.Code,
		Description: r.Description,
		ParentID:    parentID,
	}

	return taxonomy, nil
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

// ToEntity transforms UpdateTaxonomyRequest to update an existing Taxonomy entity
func (r *UpdateTaxonomyRequest) ToEntity(existingTaxonomy *entities.Taxonomy) (*entities.Taxonomy, error) {
	// Update fields only if provided
	if r.Name != "" {
		existingTaxonomy.Name = r.Name
	}
	if r.Slug != "" {
		existingTaxonomy.Slug = r.Slug
	}
	if r.Code != "" {
		existingTaxonomy.Code = r.Code
	}
	if r.Description != "" {
		existingTaxonomy.Description = r.Description
	}
	if r.ParentID != "" {
		parentID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		existingTaxonomy.ParentID = &parentID
	}

	return existingTaxonomy, nil
}
