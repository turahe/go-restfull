package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OrganizationStatus represents the status of an organization
type OrganizationStatus string

type OrganizationType string

const (
	OrganizationTypeCompany            OrganizationType = "COMPANY"
	OrganizationTypeCompanySubsidiary  OrganizationType = "COMPANY_SUBSIDIARY"
	OrganizationTypeCompanyAgent       OrganizationType = "COMPANY_AGENT"
	OrganizationTypeCompanyLicensee    OrganizationType = "COMPANY_LICENSEE"
	OrganizationTypeCompanyDistributor OrganizationType = "COMPANY_DISTRIBUTOR"
	OrganizationTypeCompanyConsignee   OrganizationType = "COMPANY_CONSIGNEE"
	OrganizationTypeCompanyConsignor   OrganizationType = "COMPANY_CONSIGNOR"
	OrganizationTypeCompanyConsigner   OrganizationType = "COMPANY_CONSIGNER"
	OrganizationTypeOutlet             OrganizationType = "OUTLET"
	OrganizationTypeStore              OrganizationType = "STORE"
	OrganizationTypeDepartment         OrganizationType = "DEPARTMENT"
	OrganizationTypeSubDepartment      OrganizationType = "SUB_DEPARTMENT"
	OrganizationTypeDivision           OrganizationType = "DIVISION"
	OrganizationTypeSubDivision        OrganizationType = "SUB_DIVISION"
	OrganizationTypeDesignation        OrganizationType = "DESIGNATION"
	OrganizationTypeInstitution        OrganizationType = "INSTITUTION"
	OrganizationTypeCommunity          OrganizationType = "COMMUNITY"
	OrganizationTypeOrganization       OrganizationType = "ORGANIZATION"
	OrganizationTypeFoundation         OrganizationType = "FOUNDATION"
	OrganizationTypeBranchOffice       OrganizationType = "BRANCH_OFFICE"
	OrganizationTypeBranchOutlet       OrganizationType = "BRANCH_OUTLET"
	OrganizationTypeBranchStore        OrganizationType = "BRANCH_STORE"
	OrganizationTypeRegional           OrganizationType = "REGIONAL"
	OrganizationTypeFranchisee         OrganizationType = "FRANCHISEE"
	OrganizationTypePartner            OrganizationType = "PARTNER"
)

const (
	OrganizationStatusActive    OrganizationStatus = "ACTIVE"
	OrganizationStatusInactive  OrganizationStatus = "INACTIVE"
	OrganizationStatusSuspended OrganizationStatus = "SUSPENDED"
)

// Organization represents the core organization domain entity with nested set hierarchy
type Organization struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description,omitempty"`
	Code           *string            `json:"code,omitempty"`
	Type           *OrganizationType  `json:"type,omitempty"`
	Status         OrganizationStatus `json:"status"`
	ParentID       *uuid.UUID         `json:"parent_id,omitempty"`
	RecordLeft     *uint64            `json:"record_left,omitempty"`
	RecordRight    *uint64            `json:"record_right,omitempty"`
	RecordDepth    *uint64            `json:"record_depth,omitempty"`
	RecordOrdering *uint64            `json:"record_ordering,omitempty"`
	CreatedBy      uuid.UUID          `json:"created_by"`
	UpdatedBy      uuid.UUID          `json:"updated_by"`
	DeletedBy      *uuid.UUID         `json:"deleted_by,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	DeletedAt      *time.Time         `json:"deleted_at,omitempty"`
	// Relationships
	Parent   *Organization   `json:"parent,omitempty"`
	Children []*Organization `json:"children,omitempty"`
}

// NewOrganization creates a new organization with validation
func NewOrganization(name, description, code string, organizationType OrganizationType, parentID *uuid.UUID) (*Organization, error) {
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
	if organizationType != "" {
		org.Type = &organizationType
	}

	return org, nil
}

// UpdateOrganization updates organization information
func (o *Organization) UpdateOrganization(name, description, code string, organizationType OrganizationType) error {
	if name != "" {
		o.Name = name
	}
	if description != "" {
		o.Description = &description
	}
	if code != "" {
		o.Code = &code
	}
	if organizationType != "" {
		o.Type = &organizationType
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
func (o *Organization) GetChildrenCount() uint64 {
	if o.RecordLeft == nil || o.RecordRight == nil {
		return 0
	}
	return (*o.RecordRight - *o.RecordLeft - 1) / 2
}

// GetDescendantsCount returns the number of all descendants
func (o *Organization) GetDescendantsCount() uint64 {
	if o.RecordLeft == nil || o.RecordRight == nil {
		return 0
	}
	return (*o.RecordRight - *o.RecordLeft - 1) / 2
}
