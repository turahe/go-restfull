package request

// CreateUserRequest is used by admins to provision accounts (same fields as public register).
type CreateUserRequest struct {
	Name            string `json:"name" binding:"required,min=2,max=100"`
	Email           string `json:"email" binding:"required,email,max=190"`
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	// RoleID references roles.id (see model.Role). When omitted, the service assigns the default "user" role (entities.RoleUser) by resolving it in the database.
	RoleID *uint `json:"roleId" binding:"omitempty,gt=0"`
}

type UserListRequest struct {
	PageRequest
	SearchRequest
	Name  string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
	Email string `form:"email" json:"email" binding:"omitempty,email,max=190"`
}
