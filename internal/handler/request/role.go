package request

type CreateRoleRequest struct {
	Name string `json:"name" binding:"required,min=2,max=50"`
}

type RoleListRequest struct {
	Limit int    `json:"limit" binding:"omitempty,min=1,max=200"`
	Name  string `json:"name" binding:"omitempty,min=2,max=50"`
	Page  int    `json:"page" binding:"omitempty,min=1"`
}
