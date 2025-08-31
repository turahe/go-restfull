// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Organization entity for managing
// organizational structures with hierarchical relationships and nested set model support.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OrganizationStatus represents the current operational status of an organization.
// This enum provides predefined values for tracking the state of organizations
// throughout their lifecycle in the system.
type OrganizationStatus string

// OrganizationType represents the category or classification of an organization.
// This enum provides predefined values for different types of organizations
// to enable proper categorization and business logic handling.
type OrganizationType string

// Organization type constants defining the standard categories for organizations.
// These values represent different business entity types and organizational structures.
const (
	OrganizationTypeCompany            OrganizationType = "COMPANY"             // Standard business company
	OrganizationTypeCompanySubsidiary  OrganizationType = "COMPANY_SUBSIDIARY"  // Subsidiary of a parent company
	OrganizationTypeCompanyAgent       OrganizationType = "COMPANY_AGENT"       // Agent representing a company
	OrganizationTypeCompanyLicensee    OrganizationType = "COMPANY_LICENSEE"    // Company with licensing rights
	OrganizationTypeCompanyDistributor OrganizationType = "COMPANY_DISTRIBUTOR" // Company distributing products
	OrganizationTypeCompanyConsignee   OrganizationType = "COMPANY_CONSIGNEE"   // Company receiving consigned goods
	OrganizationTypeCompanyConsignor   OrganizationType = "COMPANY_CONSIGNOR"   // Company sending consigned goods
	OrganizationTypeCompanyConsigner   OrganizationType = "COMPANY_CONSIGNER"   // Company acting as consigner
	OrganizationTypeOutlet             OrganizationType = "OUTLET"              // Retail outlet or point of sale
	OrganizationTypeStore              OrganizationType = "STORE"               // Retail store
	OrganizationTypeDepartment         OrganizationType = "DEPARTMENT"          // Department within an organization
	OrganizationTypeSubDepartment      OrganizationType = "SUB_DEPARTMENT"      // Sub-department within a department
	OrganizationTypeDivision           OrganizationType = "DIVISION"            // Division within an organization
	OrganizationTypeSubDivision        OrganizationType = "SUB_DIVISION"        // Sub-division within a division
	OrganizationTypeDesignation        OrganizationType = "DESIGNATION"         // Specific designation or role
	OrganizationTypeInstitution        OrganizationType = "INSTITUTION"         // Educational or research institution
	OrganizationTypeCommunity          OrganizationType = "COMMUNITY"           // Community organization
	OrganizationTypeOrganization       OrganizationType = "ORGANIZATION"        // Generic organization type
	OrganizationTypeFoundation         OrganizationType = "FOUNDATION"          // Non-profit foundation
	OrganizationTypeBranchOffice       OrganizationType = "BRANCH_OFFICE"       // Branch office location
	OrganizationTypeBranchOutlet       OrganizationType = "BRANCH_OUTLET"       // Branch outlet location
	OrganizationTypeBranchStore        OrganizationType = "BRANCH_STORE"        // Branch store location
	OrganizationTypeRegional           OrganizationType = "REGIONAL"            // Regional organization
	OrganizationTypeFranchisee         OrganizationType = "FRANCHISEE"          // Franchisee organization
	OrganizationTypePartner            OrganizationType = "PARTNER"             // Partner organization
)

// Organization status constants defining the operational states.
// These values represent the current status of organizations in the system.
const (
	OrganizationStatusActive    OrganizationStatus = "ACTIVE"    // Organization is fully operational
	OrganizationStatusInactive  OrganizationStatus = "INACTIVE"  // Organization is temporarily inactive
	OrganizationStatusSuspended OrganizationStatus = "SUSPENDED" // Organization is suspended from operations
)

// Organization represents the core organization domain entity with nested set hierarchy.
// It supports hierarchical organization structures through parent-child relationships
// and nested set model implementation for efficient tree traversal and querying.
//
// The entity includes:
// - Basic organization information (name, description, code, type)
// - Status management for operational control
// - Hierarchical structure support through nested set model
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for data retention
type Organization struct {
	ID             uuid.UUID          `json:"id"`                    // Unique identifier for the organization
	Name           string             `json:"name"`                  // Display name of the organization
	Description    *string            `json:"description,omitempty"` // Optional description of the organization
	Code           *string            `json:"code,omitempty"`        // Optional organization code/identifier
	Type           *OrganizationType  `json:"type,omitempty"`        // Optional organization type classification
	Status         OrganizationStatus `json:"status"`                // Current operational status of the organization
	ParentID       *uuid.UUID         `json:"parent_id,omitempty"`   // ID of parent organization (nil for root)
	RecordLeft     *int64             `json:"record_left" db:"record_left"`
	RecordRight    *int64             `json:"record_right" db:"record_right"`
	RecordDepth    *int64             `json:"record_depth" db:"record_depth"`
	RecordOrdering *int64             `json:"record_ordering" db:"record_ordering"` // Display order within the same level
	CreatedBy      uuid.UUID          `json:"created_by"`                           // ID of user who created this organization
	UpdatedBy      uuid.UUID          `json:"updated_by"`                           // ID of user who last updated this organization
	DeletedBy      *uuid.UUID         `json:"deleted_by,omitempty"`                 // ID of user who deleted this organization (soft delete)
	CreatedAt      time.Time          `json:"created_at"`                           // Timestamp when organization was created
	UpdatedAt      time.Time          `json:"updated_at"`                           // Timestamp when organization was last updated
	DeletedAt      *time.Time         `json:"deleted_at,omitempty"`                 // Timestamp when organization was soft deleted

	// Relationships
	Parent   *Organization   `json:"parent,omitempty"`   // Reference to parent organization
	Children []*Organization `json:"children,omitempty"` // Collection of child organizations
}

// NewOrganization creates a new organization with validation.
// This constructor validates required fields and initializes the organization
// with default status and generated UUID and timestamps.
//
// Parameters:
//   - name: Display name of the organization (required)
//   - description: Optional description of the organization
//   - code: Optional organization code/identifier
//   - organizationType: Optional organization type classification
//   - parentID: Optional ID of parent organization (nil for root organizations)
//
// Returns:
//   - *Organization: Pointer to the newly created organization entity
//   - error: Validation error if name is empty
//
// Default values:
//   - Status: OrganizationStatusActive (organization is active by default)
//   - CreatedAt/UpdatedAt: Current timestamp
func NewOrganization(name, description, code string, organizationType OrganizationType, parentID *uuid.UUID) (*Organization, error) {
	// Validate required fields
	if name == "" {
		return nil, errors.New("name is required")
	}

	// Create organization with current timestamp
	now := time.Now()
	org := &Organization{
		ID:        uuid.New(),               // Generate new unique identifier
		Name:      name,                     // Set organization name
		Status:    OrganizationStatusActive, // Set as active by default
		ParentID:  parentID,                 // Set parent organization ID
		CreatedAt: now,                      // Set creation timestamp
		UpdatedAt: now,                      // Set initial update timestamp
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
func (o *Organization) GetChildrenCount() int64 {
	if o.RecordLeft == nil || o.RecordRight == nil {
		return 0
	}
	return (*o.RecordRight - *o.RecordLeft - 1) / 2
}

// GetDescendantsCount returns the number of all descendants
func (o *Organization) GetDescendantsCount() int64 {
	if o.RecordLeft == nil || o.RecordRight == nil {
		return 0
	}
	return (*o.RecordRight - *o.RecordLeft - 1) / 2
}
