package services

// RBACService defines the interface for role-based access control operations
type RBACService interface {
	// CheckPermission checks if a subject has permission to perform an action on an object
	CheckPermission(subject, object, action string) (bool, error)

	// AddRoleForUser adds a role to a user
	AddRoleForUser(user, role string) error

	// RemoveRoleForUser removes a role from a user
	RemoveRoleForUser(user, role string) error

	// GetRolesForUser gets all roles for a user
	GetRolesForUser(user string) ([]string, error)

	// GetUsersForRole gets all users for a role
	GetUsersForRole(role string) ([]string, error)

	// AddPolicy adds a policy rule
	AddPolicy(subject, object, action string) error

	// RemovePolicy removes a policy rule
	RemovePolicy(subject, object, action string) error

	// GetPolicy gets all policy rules
	GetPolicy() ([][]string, error)

	// LoadPolicy loads policy from storage
	LoadPolicy() error

	// SavePolicy saves policy to storage
	SavePolicy() error
}
