// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Taxonomy entity for managing
// hierarchical classification systems with nested set model support.
package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Taxonomy represents a taxonomy entity in the domain layer that provides
// hierarchical classification and categorization functionality. It supports
// nested set model implementation for efficient tree traversal and querying.
//
// The entity includes:
// - Taxonomy identification (name, slug, code, description)
// - Hierarchical structure support through nested set model
// - Parent-child relationships for classification trees
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for taxonomy preservation
type Taxonomy struct {
	ID             uuid.UUID  `json:"id"`                    // Unique identifier for the taxonomy
	Name           string     `json:"name"`                  // Display name of the taxonomy
	Slug           string     `json:"slug"`                  // URL-friendly identifier for the taxonomy
	Code           string     `json:"code,omitempty"`        // Optional taxonomy code/identifier
	Description    string     `json:"description,omitempty"` // Optional description of the taxonomy
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`   // ID of parent taxonomy (nil for root items)
	RecordLeft     *int64     `json:"record_left" db:"record_left"`
	RecordRight    *int64     `json:"record_right" db:"record_right"`
	RecordDepth    *int64     `json:"record_depth" db:"record_depth"`
	RecordOrdering *int64     `json:"record_ordering" db:"record_ordering"`
	CreatedBy      uuid.UUID  `json:"created_by"`           // ID of user who created this taxonomy
	UpdatedBy      uuid.UUID  `json:"updated_by"`           // ID of user who last updated this taxonomy
	DeletedBy      *uuid.UUID `json:"deleted_by,omitempty"` // ID of user who deleted this taxonomy (soft delete)
	CreatedAt      time.Time  `json:"created_at"`           // Timestamp when taxonomy was created
	UpdatedAt      time.Time  `json:"updated_at"`           // Timestamp when taxonomy was last updated
	DeletedAt      *time.Time `json:"deleted_at,omitempty"` // Timestamp when taxonomy was soft deleted

	// Relationships
	Parent   *Taxonomy   `json:"parent,omitempty"`   // Reference to parent taxonomy
	Children []*Taxonomy `json:"children,omitempty"` // Collection of child taxonomies
}

// NewTaxonomy creates a new taxonomy instance with the provided details.
// This constructor initializes required fields and sets default values
// for timestamps and generates a new UUID for the taxonomy.
//
// Parameters:
//   - name: Display name of the taxonomy
//   - slug: URL-friendly identifier for the taxonomy
//   - code: Optional taxonomy code/identifier
//   - description: Optional description of the taxonomy
//   - parentID: Optional ID of parent taxonomy (nil for root items)
//
// Returns:
//   - *Taxonomy: Pointer to the newly created taxonomy entity
//
// Note: This constructor does not perform validation - use Validate() method for data integrity checks
func NewTaxonomy(name, slug, code, description string, parentID *uuid.UUID) *Taxonomy {
	return &Taxonomy{
		ID:          uuid.New(),    // Generate new unique identifier
		Name:        name,          // Set taxonomy name
		Slug:        slug,          // Set taxonomy slug
		Code:        code,          // Set taxonomy code
		Description: description,   // Set taxonomy description
		ParentID:    parentID,      // Set parent taxonomy ID
		CreatedAt:   time.Now(),    // Set creation timestamp
		UpdatedAt:   time.Now(),    // Set initial update timestamp
		Children:    []*Taxonomy{}, // Initialize empty children slice
	}
}

// UpdateTaxonomy updates taxonomy information with new values.
// This method modifies the taxonomy fields and automatically updates
// the UpdatedAt timestamp to reflect the change.
//
// Parameters:
//   - name: New display name for the taxonomy
//   - slug: New URL-friendly identifier
//   - code: New taxonomy code/identifier
//   - description: New taxonomy description
//   - parentID: New parent taxonomy ID
//
// Note: This method automatically updates the UpdatedAt timestamp
func (t *Taxonomy) UpdateTaxonomy(name, slug, code, description string, parentID *uuid.UUID) {
	t.Name = name               // Update taxonomy name
	t.Slug = slug               // Update taxonomy slug
	t.Code = code               // Update taxonomy code
	t.Description = description // Update taxonomy description
	t.ParentID = parentID       // Update parent taxonomy ID
	t.UpdatedAt = time.Now()    // Update modification timestamp
}

// SoftDelete marks the taxonomy as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The taxonomy will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (t *Taxonomy) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now // Set deletion timestamp
	t.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the taxonomy has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
// This method is useful for filtering out deleted taxonomies from queries.
//
// Returns:
//   - bool: true if taxonomy is deleted, false if active
func (t *Taxonomy) IsDeleted() bool {
	return t.DeletedAt != nil
}

// AddChild adds a child taxonomy to this taxonomy's children collection.
// This method maintains the parent-child relationship structure.
//
// Parameters:
//   - child: Pointer to the child taxonomy to add
//
// Note: This method only adds the child if it's not nil
func (t *Taxonomy) AddChild(child *Taxonomy) {
	if child != nil {
		t.Children = append(t.Children, child)
	}
}

// RemoveChild removes a child taxonomy from this taxonomy's children collection.
// This method searches for the child by ID and removes it if found.
//
// Parameters:
//   - childID: UUID of the child taxonomy to remove
//
// Note: This method modifies the Children slice in place
func (t *Taxonomy) RemoveChild(childID uuid.UUID) {
	for i, child := range t.Children {
		if child.ID == childID {
			// Remove child by slicing around the found index
			t.Children = append(t.Children[:i], t.Children[i+1:]...)
			break
		}
	}
}

// Validate validates the taxonomy data to ensure data integrity.
// This method checks that all required fields are properly set.
//
// Returns:
//   - error: Validation error if any required field is invalid, nil if valid
//
// Validation rules:
// - name cannot be empty or only whitespace
// - slug cannot be empty or only whitespace
func (t *Taxonomy) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("taxonomy name is required")
	}

	if strings.TrimSpace(t.Slug) == "" {
		return errors.New("taxonomy slug is required")
	}

	return nil
}

// IsRoot checks if the taxonomy is a root taxonomy (no parent)
func (t *Taxonomy) IsRoot() bool {
	return t.ParentID == nil
}

// IsLeaf checks if the taxonomy is a leaf taxonomy (no children)
func (t *Taxonomy) IsLeaf() bool {
	return len(t.Children) == 0
}

// GetDepth calculates the depth of the taxonomy in the hierarchy
func (t *Taxonomy) GetDepth() int {
	if t.IsRoot() {
		return 0
	}

	depth := 1
	parent := t.Parent
	for parent != nil && !parent.IsRoot() {
		depth++
		parent = parent.Parent
	}

	return depth
}

// GetWidth returns the width of the node in the nested set
func (t *Taxonomy) GetWidth() int64 {
	return *t.RecordRight - *t.RecordLeft + 1
}

// IsDescendantOf checks if this taxonomy is a descendant of the given taxonomy
func (t *Taxonomy) IsDescendantOf(ancestor *Taxonomy) bool {
	return *t.RecordLeft > *ancestor.RecordLeft && *t.RecordRight < *ancestor.RecordRight
}

// IsAncestorOf checks if this taxonomy is an ancestor of the given taxonomy
func (t *Taxonomy) IsAncestorOf(descendant *Taxonomy) bool {
	return *t.RecordLeft < *descendant.RecordLeft && *t.RecordRight > *descendant.RecordRight
}
