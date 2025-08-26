package responses

// RBACPolicyResource represents an RBAC policy in API responses
type RBACPolicyResource struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// RBACPolicyCollection represents a collection of RBAC policies
type RBACPolicyCollection struct {
	Data  []RBACPolicyResource `json:"data"`
	Meta  CollectionMeta       `json:"meta"`
	Links CollectionLinks      `json:"links"`
}

// RBACPolicyResourceResponse represents a single RBAC policy response
type RBACPolicyResourceResponse struct {
	ResponseCode    int                `json:"response_code"`
	ResponseMessage string             `json:"response_message"`
	Data            RBACPolicyResource `json:"data"`
}

// RBACPolicyCollectionResponse represents a collection of RBAC policies response
type RBACPolicyCollectionResponse struct {
	ResponseCode    int                  `json:"response_code"`
	ResponseMessage string               `json:"response_message"`
	Data            RBACPolicyCollection `json:"data"`
}

// RBACRoleResource represents an RBAC role in API responses
type RBACRoleResource struct {
	Role string `json:"role"`
}

// RBACRoleCollection represents a collection of RBAC roles
type RBACRoleCollection struct {
	Data  []RBACRoleResource `json:"data"`
	Meta  CollectionMeta     `json:"meta"`
	Links CollectionLinks    `json:"links"`
}

// RBACRoleCollectionResponse represents a collection of RBAC roles response
type RBACRoleCollectionResponse struct {
	ResponseCode    int                `json:"response_code"`
	ResponseMessage string             `json:"response_message"`
	Data            RBACRoleCollection `json:"data"`
}

// RBACUserResource represents an RBAC user in API responses
type RBACUserResource struct {
	User string `json:"user"`
}

// RBACUserCollection represents a collection of RBAC users
type RBACUserCollection struct {
	Data  []RBACUserResource `json:"data"`
	Meta  CollectionMeta     `json:"meta"`
	Links CollectionLinks    `json:"links"`
}

// RBACUserCollectionResponse represents a collection of RBAC users response
type RBACUserCollectionResponse struct {
	ResponseCode    int                `json:"response_code"`
	ResponseMessage string             `json:"response_message"`
	Data            RBACUserCollection `json:"data"`
}

// NewRBACPolicyResource creates a new RBACPolicyResource from policy data
func NewRBACPolicyResource(subject, object, action string) RBACPolicyResource {
	return RBACPolicyResource{
		Subject: subject,
		Object:  object,
		Action:  action,
	}
}

// NewRBACPolicyResourceResponse creates a new RBACPolicyResourceResponse
func NewRBACPolicyResourceResponse(subject, object, action string) RBACPolicyResourceResponse {
	return RBACPolicyResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Policy operation successful",
		Data:            NewRBACPolicyResource(subject, object, action),
	}
}

// NewRBACPolicyCollection creates a new RBACPolicyCollection
func NewRBACPolicyCollection(policies [][]string) RBACPolicyCollection {
	policyResources := make([]RBACPolicyResource, len(policies))
	for i, policy := range policies {
		if len(policy) >= 3 {
			policyResources[i] = NewRBACPolicyResource(policy[0], policy[1], policy[2])
		}
	}

	return RBACPolicyCollection{
		Data: policyResources,
	}
}

// NewRBACPolicyCollectionResponse creates a new RBACPolicyCollectionResponse
func NewRBACPolicyCollectionResponse(policies [][]string) RBACPolicyCollectionResponse {
	return RBACPolicyCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Policies retrieved successfully",
		Data:            NewRBACPolicyCollection(policies),
	}
}

// NewRBACRoleResource creates a new RBACRoleResource from role data
func NewRBACRoleResource(role string) RBACRoleResource {
	return RBACRoleResource{
		Role: role,
	}
}

// NewRBACRoleCollection creates a new RBACRoleCollection
func NewRBACRoleCollection(roles []string) RBACRoleCollection {
	roleResources := make([]RBACRoleResource, len(roles))
	for i, role := range roles {
		roleResources[i] = NewRBACRoleResource(role)
	}

	return RBACRoleCollection{
		Data: roleResources,
	}
}

// NewRBACRoleCollectionResponse creates a new RBACRoleCollectionResponse
func NewRBACRoleCollectionResponse(roles []string) RBACRoleCollectionResponse {
	return RBACRoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Roles retrieved successfully",
		Data:            NewRBACRoleCollection(roles),
	}
}

// NewRBACUserResource creates a new RBACUserResource from user data
func NewRBACUserResource(user string) RBACUserResource {
	return RBACUserResource{
		User: user,
	}
}

// NewRBACUserCollection creates a new RBACUserCollection
func NewRBACUserCollection(users []string) RBACUserCollection {
	userResources := make([]RBACUserResource, len(users))
	for i, user := range users {
		userResources[i] = NewRBACUserResource(user)
	}

	return RBACUserCollection{
		Data: userResources,
	}
}

// NewRBACUserCollectionResponse creates a new RBACUserCollectionResponse
func NewRBACUserCollectionResponse(users []string) RBACUserCollectionResponse {
	return RBACUserCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Users retrieved successfully",
		Data:            NewRBACUserCollection(users),
	}
}
