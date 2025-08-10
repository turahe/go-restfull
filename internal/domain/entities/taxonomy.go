package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Taxonomy represents a taxonomy entity in the domain layer
type Taxonomy struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Code        string     `json:"code,omitempty"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	RecordLeft  *uint64    `json:"record_left,omitempty"`
	RecordRight *uint64    `json:"record_right,omitempty"`
	RecordDepth *uint64    `json:"record_depth,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   uuid.UUID  `json:"updated_by"`
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	Parent   *Taxonomy   `json:"parent,omitempty"`
	Children []*Taxonomy `json:"children,omitempty"`
}

// NewTaxonomy creates a new taxonomy instance
func NewTaxonomy(name, slug, code, description string, parentID *uuid.UUID) *Taxonomy {
	return &Taxonomy{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Code:        code,
		Description: description,
		ParentID:    parentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Children:    []*Taxonomy{},
	}
}

// UpdateTaxonomy updates taxonomy information
func (t *Taxonomy) UpdateTaxonomy(name, slug, code, description string, parentID *uuid.UUID) {
	t.Name = name
	t.Slug = slug
	t.Code = code
	t.Description = description
	t.ParentID = parentID
	t.UpdatedAt = time.Now()
}

// SoftDelete marks the taxonomy as deleted
func (t *Taxonomy) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now
	t.UpdatedAt = now
}

// IsDeleted checks if the taxonomy is soft deleted
func (t *Taxonomy) IsDeleted() bool {
	return t.DeletedAt != nil
}

// AddChild adds a child taxonomy
func (t *Taxonomy) AddChild(child *Taxonomy) {
	if child != nil {
		t.Children = append(t.Children, child)
	}
}

// RemoveChild removes a child taxonomy
func (t *Taxonomy) RemoveChild(childID uuid.UUID) {
	for i, child := range t.Children {
		if child.ID == childID {
			t.Children = append(t.Children[:i], t.Children[i+1:]...)
			break
		}
	}
}

// Validate validates the taxonomy data
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
func (t *Taxonomy) GetWidth() uint64 {
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
