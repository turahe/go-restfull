package request

type CreateRoleRequest struct {
	Name string `json:"name" binding:"required,min=2,max=50"`
}

type RoleListRequest struct {
	PageRequest
	SearchRequest
	Name string `form:"name" json:"name" binding:"omitempty,min=2,max=50"`
}
