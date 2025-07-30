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

// CasbinRBACService implements the RBACService interface using Casbin
type CasbinRBACService struct {
	enforcer *casbin.Enforcer
}

// NewCasbinRBACService creates a new Casbin RBAC service instance
func NewCasbinRBACService() (services.RBACService, error) {
	cfg := config.GetConfig()

	// Create file adapter for now (temporary solution)
	fileAdapter := fileadapter.NewAdapter(cfg.Casbin.Policy)

	// Load model from file
	model, err := model.NewModelFromFile(cfg.Casbin.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(model, fileAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load initial policy from file
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	return &CasbinRBACService{
		enforcer: enforcer,
	}, nil
}

// CheckPermission checks if a subject has permission to perform an action on an object
func (c *CasbinRBACService) CheckPermission(subject, object, action string) (bool, error) {
	// Normalize the object path
	object = c.normalizePath(object)

	// Check permission
	allowed, err := c.enforcer.Enforce(subject, object, action)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return allowed, nil
}

// AddRoleForUser adds a role to a user
func (c *CasbinRBACService) AddRoleForUser(user, role string) error {
	_, err := c.enforcer.AddRoleForUser(user, role)
	if err != nil {
		return fmt.Errorf("failed to add role for user: %w", err)
	}

	return c.enforcer.SavePolicy()
}

// RemoveRoleForUser removes a role from a user
func (c *CasbinRBACService) RemoveRoleForUser(user, role string) error {
	_, err := c.enforcer.DeleteRoleForUser(user, role)
	if err != nil {
		return fmt.Errorf("failed to remove role for user: %w", err)
	}

	return c.enforcer.SavePolicy()
}

// GetRolesForUser gets all roles for a user
func (c *CasbinRBACService) GetRolesForUser(user string) ([]string, error) {
	roles, err := c.enforcer.GetRolesForUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}

	return roles, nil
}

// GetUsersForRole gets all users for a role
func (c *CasbinRBACService) GetUsersForRole(role string) ([]string, error) {
	users, err := c.enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}

	return users, nil
}

// AddPolicy adds a policy rule
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

// RemovePolicy removes a policy rule
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

// GetPolicy gets all policy rules
func (c *CasbinRBACService) GetPolicy() ([][]string, error) {
	policy, err := c.enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return policy, nil
}

// LoadPolicy loads policy from storage
func (c *CasbinRBACService) LoadPolicy() error {
	return c.enforcer.LoadPolicy()
}

// SavePolicy saves policy to storage
func (c *CasbinRBACService) SavePolicy() error {
	return c.enforcer.SavePolicy()
}

// normalizePath normalizes the path for consistent matching
func (c *CasbinRBACService) normalizePath(path string) string {
	// Remove trailing slash except for root
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
