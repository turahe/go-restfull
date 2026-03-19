package request

type UserListRequest struct {
	Limit int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=200"`
	Name  string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
	Email string `form:"email" json:"email" binding:"omitempty,email,max=190"`
	Role  string `form:"role" json:"role" binding:"omitempty,min=2,max=50"`
	Page  int    `form:"page" json:"page" binding:"omitempty,min=1"`
}
