package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OrganizationStatus represents the status of an organization
type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusInactive  OrganizationStatus = "inactive"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
)

// Organization represents the core organization domain entity with nested set hierarchy
type Organization struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description,omitempty"`
	Code           *string            `json:"code,omitempty"`
	Email          *string            `json:"email,omitempty"`
	Phone          *string            `json:"phone,omitempty"`
	Address        *string            `json:"address,omitempty"`
	Website        *string            `json:"website,omitempty"`
	LogoURL        *string            `json:"logo_url,omitempty"`
	Status         OrganizationStatus `json:"status"`
	ParentID       *uuid.UUID         `json:"parent_id,omitempty"`
	RecordLeft     *int               `json:"record_left,omitempty"`
	RecordRight    *int               `json:"record_right,omitempty"`
	RecordDepth    *int               `json:"record_depth,omitempty"`
	RecordOrdering *int               `json:"record_ordering,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	DeletedAt      *time.Time         `json:"deleted_at,omitempty"`
	// Relationships
	Parent   *Organization   `json:"parent,omitempty"`
	Children []*Organization `json:"children,omitempty"`
}

// NewOrganization creates a new organization with validation
func NewOrganization(name, description, code, email, phone, address, website, logoURL string, parentID *uuid.UUID) (*Organization, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	now := time.Now()
	org := &Organization{
		ID:        uuid.New(),
		Name:      name,
		Status:    OrganizationStatusActive,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set optional fields if provided
	if description != "" {
		org.Description = &description
	}
	if code != "" {
		org.Code = &code
	}
	if email != "" {
		org.Email = &email
	}
	if phone != "" {
		org.Phone = &phone
	}
	if address != "" {
		org.Address = &address
	}
	if website != "" {
		org.Website = &website
	}
	if logoURL != "" {
		org.LogoURL = &logoURL
	}

	return org, nil
}

// UpdateOrganization updates organization information
func (o *Organization) UpdateOrganization(name, description, code, email, phone, address, website, logoURL string) error {
	if name != "" {
		o.Name = name
	}
	if description != "" {
		o.Description = &description
	}
	if code != "" {
		o.Code = &code
	}
	if email != "" {
		o.Email = &email
	}
	if phone != "" {
		o.Phone = &phone
	}
	if address != "" {
		o.Address = &address
	}
	if website != "" {
		o.Website = &website
	}
	if logoURL != "" {
		o.LogoURL = &logoURL
	}
	o.UpdatedAt = time.Now()
	return nil
}

// SetStatus updates the organization status
func (o *Organization) SetStatus(status OrganizationStatus) error {
	switch status {
	case OrganizationStatusActive, OrganizationStatusInactive, OrganizationStatusSuspended:
		o.Status = status
		o.UpdatedAt = time.Now()
		return nil
	default:
		return errors.New("invalid organization status")
	}
}

// SetParent updates the parent organization
func (o *Organization) SetParent(parentID *uuid.UUID) {
	o.ParentID = parentID
	o.UpdatedAt = time.Now()
}

// SetNestedSetValues sets the nested set hierarchy values
func (o *Organization) SetNestedSetValues(left, right, depth, ordering *int) {
	o.RecordLeft = left
	o.RecordRight = right
	o.RecordDepth = depth
	o.RecordOrdering = ordering
	o.UpdatedAt = time.Now()
}

// SoftDelete marks the organization as deleted
func (o *Organization) SoftDelete() {
	now := time.Now()
	o.DeletedAt = &now
	o.UpdatedAt = now
}

// IsDeleted checks if the organization is soft deleted
func (o *Organization) IsDeleted() bool {
	return o.DeletedAt != nil
}

// IsActive checks if the organization is active
func (o *Organization) IsActive() bool {
	return o.Status == OrganizationStatusActive
}

// IsRoot checks if the organization is a root node (no parent)
func (o *Organization) IsRoot() bool {
	return o.ParentID == nil
}

// HasChildren checks if the organization has children
func (o *Organization) HasChildren() bool {
	return o.RecordLeft != nil && o.RecordRight != nil &&
		*o.RecordRight-*o.RecordLeft > 1
}

// GetChildrenCount returns the number of direct children
func (o *Organization) GetChildrenCount() int {
	if o.RecordLeft == nil || o.RecordRight == nil {
		return 0
	}
	return (*o.RecordRight - *o.RecordLeft - 1) / 2
}

// GetDescendantsCount returns the number of all descendants
func (o *Organization) GetDescendantsCount() int {
	if o.RecordLeft == nil || o.RecordRight == nil {
		return 0
	}
	return (*o.RecordRight - *o.RecordLeft - 1) / 2
}
