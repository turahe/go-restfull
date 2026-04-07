package request

type AssignRoleRequest struct {
	UserID uint   `json:"userId" binding:"required,gt=0"`
	Role   string `json:"role" binding:"required,min=2,max=50"`
}

type AddPermissionRequest struct {
	Role string `json:"role" binding:"required,min=2,max=50"`
	Obj  string `json:"obj" binding:"required,min=1,max=200"`
	Act  string `json:"act" binding:"required,min=1,max=50"`
}
