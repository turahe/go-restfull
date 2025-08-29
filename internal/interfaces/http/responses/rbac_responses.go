// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

// RBACPolicyResource represents an RBAC policy in API responses.
// This struct defines a single Role-Based Access Control policy with subject,
// object, and action components following the standard RBAC model.
type RBACPolicyResource struct {
	// Subject is the entity (user, role, or group) that the policy applies to
	Subject string `json:"subject"`
	// Object is the resource or entity that the policy governs access to
	Object string `json:"object"`
	// Action is the operation or permission being granted/denied
	Action string `json:"action"`
}

// RBACPolicyCollection represents a collection of RBAC policies.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type RBACPolicyCollection struct {
	// Data contains the array of RBAC policy resources
	Data []RBACPolicyResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// RBACPolicyResourceResponse represents a single RBAC policy response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type RBACPolicyResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the RBAC policy resource
	Data RBACPolicyResource `json:"data"`
}

// RBACPolicyCollectionResponse represents a collection of RBAC policies response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type RBACPolicyCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the RBAC policy collection
	Data RBACPolicyCollection `json:"data"`
}

// RBACRoleResource represents an RBAC role in API responses.
// This struct defines a single role within the Role-Based Access Control system.
type RBACRoleResource struct {
	// Role is the name or identifier of the role
	Role string `json:"role"`
}

// RBACRoleCollection represents a collection of RBAC roles.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type RBACRoleCollection struct {
	// Data contains the array of RBAC role resources
	Data []RBACRoleResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// RBACRoleCollectionResponse represents a collection of RBAC roles response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type RBACRoleCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the RBAC role collection
	Data RBACRoleCollection `json:"data"`
}

// RBACUserResource represents an RBAC user in API responses.
// This struct defines a single user within the Role-Based Access Control system.
type RBACUserResource struct {
	// User is the name or identifier of the user
	User string `json:"user"`
}

// RBACUserCollection represents a collection of RBAC users.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type RBACUserCollection struct {
	// Data contains the array of RBAC user resources
	Data []RBACUserResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// RBACUserCollectionResponse represents a collection of RBAC users response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type RBACUserCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the RBAC user collection
	Data RBACUserCollection `json:"data"`
}

// NewRBACPolicyResource creates a new RBACPolicyResource from policy data.
// This function creates a policy resource with the three required components:
// subject (who), object (what), and action (how).
//
// Parameters:
//   - subject: The entity the policy applies to (user, role, or group)
//   - object: The resource or entity being accessed
//   - action: The operation or permission being granted/denied
//
// Returns:
//   - A new RBACPolicyResource with the provided policy components
func NewRBACPolicyResource(subject, object, action string) RBACPolicyResource {
	return RBACPolicyResource{
		Subject: subject,
		Object:  object,
		Action:  action,
	}
}

// NewRBACPolicyResourceResponse creates a new RBACPolicyResourceResponse.
// This function wraps an RBACPolicyResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - subject: The entity the policy applies to
//   - object: The resource or entity being accessed
//   - action: The operation or permission being granted/denied
//
// Returns:
//   - A new RBACPolicyResourceResponse with success status and policy data
func NewRBACPolicyResourceResponse(subject, object, action string) RBACPolicyResourceResponse {
	return RBACPolicyResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Policy operation successful",
		Data:            NewRBACPolicyResource(subject, object, action),
	}
}

// NewRBACPolicyCollection creates a new RBACPolicyCollection.
// This function transforms a slice of policy data arrays into a collection
// of RBACPolicyResource objects, validating that each policy has the required
// three components (subject, object, action).
//
// Parameters:
//   - policies: Slice of policy data arrays, each containing [subject, object, action]
//
// Returns:
//   - A new RBACPolicyCollection with all valid policies properly formatted
func NewRBACPolicyCollection(policies [][]string) RBACPolicyCollection {
	policyResources := make([]RBACPolicyResource, len(policies))
	for i, policy := range policies {
		// Ensure each policy has the required three components
		if len(policy) >= 3 {
			policyResources[i] = NewRBACPolicyResource(policy[0], policy[1], policy[2])
		}
	}

	return RBACPolicyCollection{
		Data: policyResources,
	}
}

// NewRBACPolicyCollectionResponse creates a new RBACPolicyCollectionResponse.
// This function wraps an RBACPolicyCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - policies: Slice of policy data arrays, each containing [subject, object, action]
//
// Returns:
//   - A new RBACPolicyCollectionResponse with success status and policy collection data
func NewRBACPolicyCollectionResponse(policies [][]string) RBACPolicyCollectionResponse {
	return RBACPolicyCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Policies retrieved successfully",
		Data:            NewRBACPolicyCollection(policies),
	}
}

// NewRBACRoleResource creates a new RBACRoleResource from role data.
// This function creates a role resource with the specified role name.
//
// Parameters:
//   - role: The name or identifier of the role
//
// Returns:
//   - A new RBACRoleResource with the provided role information
func NewRBACRoleResource(role string) RBACRoleResource {
	return RBACRoleResource{
		Role: role,
	}
}

// NewRBACRoleCollection creates a new RBACRoleCollection.
// This function transforms a slice of role names into a collection
// of RBACRoleResource objects.
//
// Parameters:
//   - roles: Slice of role names or identifiers
//
// Returns:
//   - A new RBACRoleCollection with all roles properly formatted
func NewRBACRoleCollection(roles []string) RBACRoleCollection {
	roleResources := make([]RBACRoleResource, len(roles))
	for i, role := range roles {
		roleResources[i] = NewRBACRoleResource(role)
	}

	return RBACRoleCollection{
		Data: roleResources,
	}
}

// NewRBACRoleCollectionResponse creates a new RBACRoleCollectionResponse.
// This function wraps an RBACRoleCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - roles: Slice of role names or identifiers
//
// Returns:
//   - A new RBACRoleCollectionResponse with success status and role collection data
func NewRBACRoleCollectionResponse(roles []string) RBACRoleCollectionResponse {
	return RBACRoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Roles retrieved successfully",
		Data:            NewRBACRoleCollection(roles),
	}
}

// NewRBACUserResource creates a new RBACUserResource from user data.
// This function creates a user resource with the specified user name.
//
// Parameters:
//   - user: The name or identifier of the user
//
// Returns:
//   - A new RBACUserResource with the provided user information
func NewRBACUserResource(user string) RBACUserResource {
	return RBACUserResource{
		User: user,
	}
}

// NewRBACUserCollection creates a new RBACUserCollection.
// This function transforms a slice of user names into a collection
// of RBACUserResource objects.
//
// Parameters:
//   - users: Slice of user names or identifiers
//
// Returns:
//   - A new RBACUserCollection with all users properly formatted
func NewRBACUserCollection(users []string) RBACUserCollection {
	userResources := make([]RBACUserResource, len(users))
	for i, user := range users {
		userResources[i] = NewRBACUserResource(user)
	}

	return RBACUserCollection{
		Data: userResources,
	}
}

// NewRBACUserCollectionResponse creates a new RBACUserCollectionResponse.
// This function wraps an RBACUserCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - users: Slice of user names or identifiers
//
// Returns:
//   - A new RBACUserCollectionResponse with success status and user collection data
func NewRBACUserCollectionResponse(users []string) RBACUserCollectionResponse {
	return RBACUserCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Users retrieved successfully",
		Data:            NewRBACUserCollection(users),
	}
}
