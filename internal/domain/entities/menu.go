// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Menu entity for managing
// hierarchical menu structures with nested set model support and role-based access control.
package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Menu represents a menu entity in the domain layer that supports
// hierarchical organization through nested set model implementation.
//
// The entity provides:
// - Hierarchical menu structure with parent-child relationships
// - Nested set model for efficient tree traversal and querying
// - Role-based access control for menu visibility
// - Menu state management (active/inactive, visible/hidden)
// - URL routing and target configuration
// - Audit trail with creation, update, and deletion tracking
type Menu struct {
	ID             uuid.UUID  `json:"id"`                        // Unique identifier for the menu
	Name           string     `json:"name"`                      // Display name for the menu item
	Slug           string     `json:"slug"`                      // URL-friendly identifier for the menu
	Description    string     `json:"description,omitempty"`     // Optional description of the menu
	URL            string     `json:"url,omitempty"`             // Target URL for the menu link
	Icon           string     `json:"icon,omitempty"`            // Icon identifier for the menu item
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`       // ID of parent menu (nil for root items)
	RecordLeft     *uint64    `json:"record_left,omitempty"`     // Left boundary for nested set model
	RecordRight    *uint64    `json:"record_right,omitempty"`    // Right boundary for nested set model
	RecordOrdering *uint64    `json:"record_ordering,omitempty"` // Display order within the same level
	RecordDepth    *uint64    `json:"record_depth,omitempty"`    // Depth level in the hierarchy
	IsActive       bool       `json:"is_active"`                 // Whether the menu is active/enabled
	IsVisible      bool       `json:"is_visible"`                // Whether the menu is visible to users
	Target         string     `json:"target,omitempty"`          // Link target (_blank, _self, etc.)
	CreatedBy      uuid.UUID  `json:"created_by"`                // ID of user who created this menu
	UpdatedBy      uuid.UUID  `json:"updated_by"`                // ID of user who last updated this menu
	DeletedBy      *uuid.UUID `json:"deleted_by,omitempty"`      // ID of user who deleted this menu (soft delete)
	CreatedAt      time.Time  `json:"created_at"`                // Timestamp when menu was created
	UpdatedAt      time.Time  `json:"updated_at"`                // Timestamp when menu was last updated
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`      // Timestamp when menu was soft deleted

	// Relationships
	Parent   *Menu   `json:"parent,omitempty"`   // Reference to parent menu item
	Children []*Menu `json:"children,omitempty"` // Collection of child menu items
	Roles    []*Role `json:"roles,omitempty"`    // Roles that can access this menu
}

// NewMenu creates a new menu instance with the provided details.
// This constructor initializes required fields and sets default values
// for active/visible status, target, and timestamps.
//
// Parameters:
//   - name: Display name for the menu item
//   - slug: URL-friendly identifier for the menu
//   - description: Optional description of the menu
//   - url: Target URL for the menu link
//   - icon: Icon identifier for the menu item
//   - parentID: Optional ID of parent menu (nil for root items)
//
// Returns:
//   - *Menu: Pointer to the newly created menu entity
//
// Default values:
//   - IsActive: true (menu is enabled by default)
//   - IsVisible: true (menu is visible by default)
//   - Target: "_self" (open in same window by default)
func NewMenu(name, slug, description, url, icon string, parentID *uuid.UUID) *Menu {
	return &Menu{
		ID:          uuid.New(),  // Generate new unique identifier
		Name:        name,        // Set menu name
		Slug:        slug,        // Set menu slug
		Description: description, // Set menu description
		URL:         url,         // Set target URL
		Icon:        icon,        // Set icon identifier
		ParentID:    parentID,    // Set parent menu ID
		IsActive:    true,        // Set as active by default
		IsVisible:   true,        // Set as visible by default
		Target:      "_self",     // Set default target
		CreatedAt:   time.Now(),  // Set creation timestamp
		UpdatedAt:   time.Now(),  // Set initial update timestamp
		Children:    []*Menu{},   // Initialize empty children slice
		Roles:       []*Role{},   // Initialize empty roles slice
	}
}

// UpdateMenu updates menu information with new values.
// This method modifies the menu fields and automatically updates
// the UpdatedAt timestamp to reflect the change.
//
// Parameters:
//   - name: New display name for the menu
//   - slug: New URL-friendly identifier
//   - description: New menu description
//   - url: New target URL
//   - icon: New icon identifier
//   - recordOrdering: New display order within the same level
//   - parentID: New parent menu ID
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Menu) UpdateMenu(name, slug, description, url, icon string, recordOrdering uint64, parentID *uuid.UUID) {
	m.Name = name                      // Update menu name
	m.Slug = slug                      // Update menu slug
	m.Description = description        // Update menu description
	m.URL = url                        // Update target URL
	m.Icon = icon                      // Update icon identifier
	m.RecordOrdering = &recordOrdering // Update display order
	m.ParentID = parentID              // Update parent menu ID
	m.UpdatedAt = time.Now()           // Update modification timestamp
}

// Activate enables the menu item.
// This method sets IsActive to true and updates the UpdatedAt timestamp.
// Active menus are typically included in navigation and processing.
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Menu) Activate() {
	m.IsActive = true        // Enable the menu
	m.UpdatedAt = time.Now() // Update modification timestamp
}

// Deactivate disables the menu item.
// This method sets IsActive to false and updates the UpdatedAt timestamp.
// Inactive menus are typically excluded from navigation and processing.
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Menu) Deactivate() {
	m.IsActive = false       // Disable the menu
	m.UpdatedAt = time.Now() // Update modification timestamp
}

// Show makes the menu item visible to users.
// This method sets IsVisible to true and updates the UpdatedAt timestamp.
// Visible menus are displayed in the user interface.
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Menu) Show() {
	m.IsVisible = true       // Make menu visible
	m.UpdatedAt = time.Now() // Update modification timestamp
}

// Hide makes the menu item invisible to users.
// This method sets IsVisible to false and updates the UpdatedAt timestamp.
// Hidden menus are not displayed in the user interface.
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Menu) Hide() {
	m.IsVisible = false      // Make menu invisible
	m.UpdatedAt = time.Now() // Update modification timestamp
}

// SetTarget sets the target for the menu link.
// This method updates the Target field and automatically updates
// the UpdatedAt timestamp.
//
// Parameters:
//   - target: Link target value (e.g., "_blank", "_self", "_parent", "_top")
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Menu) SetTarget(target string) {
	m.Target = target        // Set link target
	m.UpdatedAt = time.Now() // Update modification timestamp
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
	return int64(*m.RecordRight - *m.RecordLeft + 1)
}

// IsDescendantOf checks if this menu is a descendant of the given menu
func (m *Menu) IsDescendantOf(ancestor *Menu) bool {
	return *m.RecordLeft > *ancestor.RecordLeft && *m.RecordRight < *ancestor.RecordRight
}

// IsAncestorOf checks if this menu is an ancestor of the given menu
func (m *Menu) IsAncestorOf(descendant *Menu) bool {
	return *m.RecordLeft < *descendant.RecordLeft && *m.RecordRight > *descendant.RecordRight
}

// IsSiblingOf checks if this menu is a sibling of the given menu
func (m *Menu) IsSiblingOf(other *Menu) bool {
	return m.ParentID != nil && other.ParentID != nil && *m.ParentID == *other.ParentID
}

// GetSiblingCount returns the number of siblings this menu has
func (m *Menu) GetSiblingCount() int {
	if m.IsRoot() {
		return 0
	}
	return len(m.Parent.Children) - 1
}

// GetPositionAmongSiblings returns the position of this menu among its siblings
func (m *Menu) GetPositionAmongSiblings() int {
	if m.IsRoot() {
		return 0
	}

	for i, sibling := range m.Parent.Children {
		if sibling.ID == m.ID {
			return i
		}
	}
	return -1
}

// GetNextSibling returns the next sibling menu
func (m *Menu) GetNextSibling() *Menu {
	if m.IsRoot() {
		return nil
	}

	position := m.GetPositionAmongSiblings()
	if position >= 0 && position < len(m.Parent.Children)-1 {
		return m.Parent.Children[position+1]
	}
	return nil
}

// GetPreviousSibling returns the previous sibling menu
func (m *Menu) GetPreviousSibling() *Menu {
	if m.IsRoot() {
		return nil
	}

	position := m.GetPositionAmongSiblings()
	if position > 0 {
		return m.Parent.Children[position-1]
	}
	return nil
}

// GetFirstChild returns the first child menu
func (m *Menu) GetFirstChild() *Menu {
	if len(m.Children) > 0 {
		return m.Children[0]
	}
	return nil
}

// GetLastChild returns the last child menu
func (m *Menu) GetLastChild() *Menu {
	if len(m.Children) > 0 {
		return m.Children[len(m.Children)-1]
	}
	return nil
}

// GetAncestors returns all ancestor menus (parent, grandparent, etc.)
func (m *Menu) GetAncestors() []*Menu {
	var ancestors []*Menu
	parent := m.Parent

	for parent != nil {
		ancestors = append([]*Menu{parent}, ancestors...)
		parent = parent.Parent
	}

	return ancestors
}

// GetDescendants returns all descendant menus (children, grandchildren, etc.)
func (m *Menu) GetDescendants() []*Menu {
	var descendants []*Menu

	for _, child := range m.Children {
		descendants = append(descendants, child)
		descendants = append(descendants, child.GetDescendants()...)
	}

	return descendants
}

// GetLeaves returns all leaf nodes (menus with no children)
func (m *Menu) GetLeaves() []*Menu {
	var leaves []*Menu

	if m.IsLeaf() {
		leaves = append(leaves, m)
	} else {
		for _, child := range m.Children {
			leaves = append(leaves, child.GetLeaves()...)
		}
	}

	return leaves
}

// GetPath returns the path from root to this menu
func (m *Menu) GetPath() []*Menu {
	path := []*Menu{m}
	ancestors := m.GetAncestors()
	path = append(ancestors, path...)
	return path
}

// GetPathString returns the path as a string representation
func (m *Menu) GetPathString() string {
	path := m.GetPath()
	var pathStrings []string

	for _, menu := range path {
		pathStrings = append(pathStrings, menu.Name)
	}

	return strings.Join(pathStrings, " > ")
}

// IsInPath checks if the given menu is in the path from root to this menu
func (m *Menu) IsInPath(target *Menu) bool {
	path := m.GetPath()

	for _, menu := range path {
		if menu.ID == target.ID {
			return true
		}
	}

	return false
}

// GetSubtreeSize returns the total number of nodes in the subtree rooted at this menu
func (m *Menu) GetSubtreeSize() int {
	size := 1 // Include self

	for _, child := range m.Children {
		size += child.GetSubtreeSize()
	}

	return size
}

// GetMaxDepth returns the maximum depth of the subtree rooted at this menu
func (m *Menu) GetMaxDepth() int {
	if m.IsLeaf() {
		return 0
	}

	maxDepth := 0
	for _, child := range m.Children {
		childDepth := child.GetMaxDepth()
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}

	return maxDepth + 1
}

// IsBalanced checks if the subtree rooted at this menu is balanced
func (m *Menu) IsBalanced() bool {
	if m.IsLeaf() {
		return true
	}

	depths := make(map[int]bool)
	for _, child := range m.Children {
		depths[child.GetMaxDepth()] = true
	}

	// All children should have the same depth for a balanced tree
	return len(depths) <= 1
}

// GetLevel returns the level of this menu in the tree (0 for root, 1 for first level, etc.)
func (m *Menu) GetLevel() int {
	return m.GetDepth()
}

// IsAtLevel checks if this menu is at the specified level
func (m *Menu) IsAtLevel(level int) bool {
	return m.GetLevel() == level
}

// GetNodesAtLevel returns all nodes at the specified level in the subtree
func (m *Menu) GetNodesAtLevel(level int) []*Menu {
	var nodes []*Menu

	if m.GetLevel() == level {
		nodes = append(nodes, m)
	} else if m.GetLevel() < level {
		for _, child := range m.Children {
			nodes = append(nodes, child.GetNodesAtLevel(level)...)
		}
	}

	return nodes
}
