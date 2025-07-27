package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Menu represents a menu entity in the domain layer
type Menu struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description,omitempty"`
	URL            string     `json:"url,omitempty"`
	Icon           string     `json:"icon,omitempty"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	RecordLeft     int64      `json:"record_left"`
	RecordRight    int64      `json:"record_right"`
	RecordOrdering int64      `json:"record_ordering"`
	IsActive       bool       `json:"is_active"`
	IsVisible      bool       `json:"is_visible"`
	Target         string     `json:"target,omitempty"` // _blank, _self, etc.
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	Parent   *Menu   `json:"parent,omitempty"`
	Children []*Menu `json:"children,omitempty"`
	Roles    []*Role `json:"roles,omitempty"`
}

// NewMenu creates a new menu instance
func NewMenu(name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) *Menu {
	return &Menu{
		ID:             uuid.New(),
		Name:           name,
		Slug:           slug,
		Description:    description,
		URL:            url,
		Icon:           icon,
		ParentID:       parentID,
		RecordLeft:     0, // Will be set by repository
		RecordRight:    0, // Will be set by repository
		RecordOrdering: recordOrdering,
		IsActive:       true,
		IsVisible:      true,
		Target:         "_self",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Children:       []*Menu{},
		Roles:          []*Role{},
	}
}

// UpdateMenu updates menu information
func (m *Menu) UpdateMenu(name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) {
	m.Name = name
	m.Slug = slug
	m.Description = description
	m.URL = url
	m.Icon = icon
	m.RecordOrdering = recordOrdering
	m.ParentID = parentID
	m.UpdatedAt = time.Now()
}

// Activate activates the menu
func (m *Menu) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// Deactivate deactivates the menu
func (m *Menu) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// Show makes the menu visible
func (m *Menu) Show() {
	m.IsVisible = true
	m.UpdatedAt = time.Now()
}

// Hide makes the menu invisible
func (m *Menu) Hide() {
	m.IsVisible = false
	m.UpdatedAt = time.Now()
}

// SetTarget sets the target for the menu link
func (m *Menu) SetTarget(target string) {
	m.Target = target
	m.UpdatedAt = time.Now()
}

// SoftDelete marks the menu as deleted
func (m *Menu) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
}

// IsDeleted checks if the menu is soft deleted
func (m *Menu) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsActiveMenu checks if the menu is active
func (m *Menu) IsActiveMenu() bool {
	return m.IsActive && !m.IsDeleted()
}

// IsVisibleMenu checks if the menu is visible
func (m *Menu) IsVisibleMenu() bool {
	return m.IsVisible && m.IsActive && !m.IsDeleted()
}

// AddChild adds a child menu
func (m *Menu) AddChild(child *Menu) {
	if child != nil {
		m.Children = append(m.Children, child)
	}
}

// RemoveChild removes a child menu
func (m *Menu) RemoveChild(childID uuid.UUID) {
	for i, child := range m.Children {
		if child.ID == childID {
			m.Children = append(m.Children[:i], m.Children[i+1:]...)
			break
		}
	}
}

// AddRole adds a role to the menu
func (m *Menu) AddRole(role *Role) {
	if role != nil {
		m.Roles = append(m.Roles, role)
	}
}

// RemoveRole removes a role from the menu
func (m *Menu) RemoveRole(roleID uuid.UUID) {
	for i, role := range m.Roles {
		if role.ID == roleID {
			m.Roles = append(m.Roles[:i], m.Roles[i+1:]...)
			break
		}
	}
}

// HasRole checks if the menu has a specific role
func (m *Menu) HasRole(roleID uuid.UUID) bool {
	for _, role := range m.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}

// Validate validates the menu data
func (m *Menu) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("menu name is required")
	}

	if strings.TrimSpace(m.Slug) == "" {
		return errors.New("menu slug is required")
	}

	if m.RecordOrdering < 0 {
		return errors.New("menu record ordering must be non-negative")
	}

	return nil
}

// IsRoot checks if the menu is a root menu (no parent)
func (m *Menu) IsRoot() bool {
	return m.ParentID == nil
}

// IsLeaf checks if the menu is a leaf menu (no children)
func (m *Menu) IsLeaf() bool {
	return len(m.Children) == 0
}

// GetDepth calculates the depth of the menu in the hierarchy
func (m *Menu) GetDepth() int {
	if m.IsRoot() {
		return 0
	}

	depth := 1
	parent := m.Parent
	for parent != nil && !parent.IsRoot() {
		depth++
		parent = parent.Parent
	}

	return depth
}

// GetWidth returns the width of the node in the nested set
func (m *Menu) GetWidth() int64 {
	return m.RecordRight - m.RecordLeft + 1
}

// IsDescendantOf checks if this menu is a descendant of the given menu
func (m *Menu) IsDescendantOf(ancestor *Menu) bool {
	return m.RecordLeft > ancestor.RecordLeft && m.RecordRight < ancestor.RecordRight
}

// IsAncestorOf checks if this menu is an ancestor of the given menu
func (m *Menu) IsAncestorOf(descendant *Menu) bool {
	return m.RecordLeft < descendant.RecordLeft && m.RecordRight > descendant.RecordRight
}
