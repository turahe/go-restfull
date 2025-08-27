// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/domain/services"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// CasbinRBACService implements the RBACService interface using Casbin for role-based access control.
// This service provides comprehensive RBAC functionality including permission checking, role management,
// policy management, and user-role assignments. It uses Casbin's powerful policy engine with file-based
// storage for policy persistence.
type CasbinRBACService struct {
	// enforcer holds the Casbin enforcer instance that manages policies and permissions
	enforcer *casbin.Enforcer
}

// NewCasbinRBACService creates a new Casbin RBAC service instance.
// This factory function initializes the Casbin enforcer with the configured model and policy files.
// It loads the initial policy and returns a fully configured RBAC service.
//
// Returns:
//   - services.RBACService: A new RBAC service instance
//   - error: Any error that occurred during service initialization
func NewCasbinRBACService() (services.RBACService, error) {
	cfg := config.GetConfig()

	// Create file adapter for policy storage (temporary solution)
	fileAdapter := fileadapter.NewAdapter(cfg.Casbin.Policy)

	// Load the RBAC model from the configured model file
	model, err := model.NewModelFromFile(cfg.Casbin.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	// Create the Casbin enforcer with the loaded model and adapter
	enforcer, err := casbin.NewEnforcer(model, fileAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load the initial policy from the policy file
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	return &CasbinRBACService{
		enforcer: enforcer,
	}, nil
}

// CheckPermission checks if a subject (user/role) has permission to perform an action on an object.
// This method normalizes the object path for consistent matching and delegates the permission
// check to the Casbin enforcer. It's the core method for access control decisions.
//
// Parameters:
//   - subject: The user or role requesting access
//   - object: The resource or endpoint being accessed
//   - action: The operation being performed (e.g., GET, POST, DELETE)
//
// Returns:
//   - bool: True if permission is granted, false if denied
//   - error: Any error that occurred during permission checking
func (c *CasbinRBACService) CheckPermission(subject, object, action string) (bool, error) {
	// Normalize the object path for consistent matching
	object = c.normalizePath(object)

	// Check permission using the Casbin enforcer
	allowed, err := c.enforcer.Enforce(subject, object, action)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return allowed, nil
}

// AddRoleForUser assigns a role to a user, enabling role-based access control.
// This method adds the role assignment to the policy and persists the changes
// to the underlying storage.
//
// Parameters:
//   - user: The username to assign the role to
//   - role: The role name to assign
//
// Returns:
//   - error: Any error that occurred during role assignment or policy persistence
func (c *CasbinRBACService) AddRoleForUser(user, role string) error {
	_, err := c.enforcer.AddRoleForUser(user, role)
	if err != nil {
		return fmt.Errorf("failed to add role for user: %w", err)
	}

	return c.enforcer.SavePolicy()
}

// RemoveRoleForUser removes a role assignment from a user.
// This method removes the role assignment from the policy and persists the changes
// to the underlying storage.
//
// Parameters:
//   - user: The username to remove the role from
//   - role: The role name to remove
//
// Returns:
//   - error: Any error that occurred during role removal or policy persistence
func (c *CasbinRBACService) RemoveRoleForUser(user, role string) error {
	_, err := c.enforcer.DeleteRoleForUser(user, role)
	if err != nil {
		return fmt.Errorf("failed to remove role for user: %w", err)
	}

	return c.enforcer.SavePolicy()
}

// GetRolesForUser retrieves all roles assigned to a specific user.
// This method is useful for displaying user permissions or checking user capabilities.
//
// Parameters:
//   - user: The username to get roles for
//
// Returns:
//   - []string: List of role names assigned to the user
//   - error: Any error that occurred during role retrieval
func (c *CasbinRBACService) GetRolesForUser(user string) ([]string, error) {
	roles, err := c.enforcer.GetRolesForUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}

	return roles, nil
}

// GetUsersForRole retrieves all users assigned to a specific role.
// This method is useful for administrative purposes and role management.
//
// Parameters:
//   - role: The role name to get users for
//
// Returns:
//   - []string: List of usernames assigned to the role
//   - error: Any error that occurred during user retrieval
func (c *CasbinRBACService) GetUsersForRole(role string) ([]string, error) {
	users, err := c.enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}

	return users, nil
}

// AddPolicy adds a new policy rule to the RBAC system.
// This method creates a permission rule that allows a subject to perform an action on an object.
// The object path is normalized for consistent matching, and the policy is persisted.
//
// Parameters:
//   - subject: The user or role the policy applies to
//   - object: The resource or endpoint the policy applies to
//   - action: The operation the policy allows
//
// Returns:
//   - error: Any error that occurred during policy addition or persistence
func (c *CasbinRBACService) AddPolicy(subject, object, action string) error {
	object = c.normalizePath(object)

	added, err := c.enforcer.AddPolicy(subject, object, action)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	if !added {
		return fmt.Errorf("policy already exists")
	}

	return c.enforcer.SavePolicy()
}

// RemovePolicy removes an existing policy rule from the RBAC system.
// This method deletes a permission rule and persists the changes. The object path
// is normalized for consistent matching.
//
// Parameters:
//   - subject: The user or role the policy applies to
//   - object: The resource or endpoint the policy applies to
//   - action: The operation the policy allows
//
// Returns:
//   - error: Any error that occurred during policy removal or persistence
func (c *CasbinRBACService) RemovePolicy(subject, object, action string) error {
	object = c.normalizePath(object)

	removed, err := c.enforcer.RemovePolicy(subject, object, action)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	if !removed {
		return fmt.Errorf("policy does not exist")
	}

	return c.enforcer.SavePolicy()
}

// GetPolicy retrieves all policy rules from the RBAC system.
// This method returns the complete set of permission rules for administrative
// and debugging purposes.
//
// Returns:
//   - [][]string: Matrix of policy rules, each row containing [subject, object, action]
//   - error: Any error that occurred during policy retrieval
func (c *CasbinRBACService) GetPolicy() ([][]string, error) {
	policy, err := c.enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return policy, nil
}

// LoadPolicy loads the policy from the underlying storage.
// This method is useful for refreshing the policy after external changes
// or for reloading policies in multi-instance deployments.
//
// Returns:
//   - error: Any error that occurred during policy loading
func (c *CasbinRBACService) LoadPolicy() error {
	return c.enforcer.LoadPolicy()
}

// SavePolicy persists the current policy to the underlying storage.
// This method ensures that all policy changes are permanently saved
// and available across service restarts.
//
// Returns:
//   - error: Any error that occurred during policy persistence
func (c *CasbinRBACService) SavePolicy() error {
	return c.enforcer.SavePolicy()
}

// normalizePath normalizes URL paths for consistent policy matching.
// This method ensures that paths are consistently formatted by:
// - Removing trailing slashes (except for root path)
// - Ensuring paths start with a forward slash
// This normalization is crucial for consistent permission checking across
// different URL formats.
//
// Parameters:
//   - path: The raw path string to normalize
//
// Returns:
//   - string: The normalized path string
func (c *CasbinRBACService) normalizePath(path string) string {
	// Remove trailing slash except for root path
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	// Ensure path starts with forward slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
